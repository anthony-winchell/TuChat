package main

import (
	"errors"
	"log"
	"net"
	"runtime/debug"
	"time"
	"tuchat/protocol"
)

const defaultWelcomeMessage = "Welcome to {server}, {nickname}!\n\nType /help for commands.\n\n" +
	"Server Address: {address}"
const maxAuthAttempts = 3
const authTimeout = 2 * time.Minute
const maxMessageSize = 1 << 20

const heartbeatInterval = 60 * time.Second
const heartbeatTimeout = 2 * time.Minute

func (s *Server) Start() {
	log.Println("Server started")

	s.wg.Go(s.heartbeatLoop)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Println(err)
			continue
		}

		s.addConnection(conn)

		s.wg.Go(func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("recovered panic in connection handler: %v\n%s", r, debug.Stack())
				}
			}()
			s.handleConnection(conn)
		})
	}
}

func (s *Server) Shutdown() {
	close(s.done)

	if err := s.listener.Close(); err != nil {
		log.Println(err)
	}

	s.sendToAll(protocol.Message{
		Type:    "system",
		Message: "Server shutting down",
	}, nil)

	s.closeClients()

	s.wg.Wait()
}

func (s *Server) heartbeatLoop() {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			now := time.Now()
			for _, c := range s.clientsSnapshot() {
				if now.Sub(c.lastPongTime()) > heartbeatTimeout {
					c.stop()
					continue
				}
				c.Send(protocol.Message{Type: "ping"})
			}
		}
	}
}

func (s *Server) addConnection(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conns[conn] = struct{}{}
}

func (s *Server) registerClient(conn net.Conn) (*Client, error) {
	var client *Client
	client = newClient(conn)

	ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	if !s.authRateLimit.Allow(ip) {
		client.Send(protocol.Message{
			Type:    "error",
			Message: "Rate limit exceeded. Wait a few seconds and try again.",
		})
		return nil, errors.New("rate limit exceeded for " + ip)
	}

	client.conn.SetReadDeadline(time.Now().Add(authTimeout))

	if s.HasPassword() {
		if err := s.requireServerPassword(client); err != nil {
			return nil, err
		}
	}

	if err := client.Send(protocol.Message{
		Type:    "auth_prompt",
		Message: "Login or register",
	}); err != nil {
		return nil, err
	}

	attempts := 0
	for {
		message, err := client.readMessage()
		if err != nil {
			return nil, err
		}

		switch message.Type {
		case "register":
			user, err := s.RegisterUser(
				message.Username,
				message.Password,
			)

			if err != nil {
				attempts++
				if attempts >= maxAuthAttempts {
					client.Send(protocol.Message{
						Type:    "error",
						Message: "Too many failed attempts",
					})
					return nil, errors.New("too many auth attempts")
				}
				client.Send(protocol.Message{
					Type:    "error",
					Message: err.Error(),
				})
				continue
			}

			client.SetUser(user)

		case "login":
			user, err := s.AuthenticateUser(message.Username, message.Password)

			if err != nil {
				attempts++
				if attempts >= maxAuthAttempts {
					client.Send(protocol.Message{
						Type:    "error",
						Message: "Too many failed attempts",
					})
					return nil, errors.New("too many auth attempts")
				}
				client.Send(protocol.Message{
					Type:    "error",
					Message: err.Error(),
				})
				continue
			}

			client.SetUser(user)
		}

		if client.User() != nil {
			if err := s.addClient(client); err != nil {
				client.Send(protocol.Message{
					Type:    "error",
					Message: err.Error(),
				})
				client.SetUser(nil)
				continue
			}
			if err := client.Send(protocol.Message{
				Type:     "auth_success",
				Nickname: client.User().Nickname(),
			}); err != nil {
				return nil, err
			}
			break
		}
	}

	if err := client.Send(protocol.Message{
		Type:    "server_name",
		Message: s.Name(),
	}); err != nil {
		return nil, err
	}

	if err := s.JoinRoom(client, "general", ""); err != nil {
		return nil, err
	}

	client.startWriter(&s.wg)

	return client, nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		s.removeConnection(conn)
		conn.Close()
	}()

	client, err := s.registerClient(conn)
	if err != nil {
		if client != nil {
			client.stop()
		}
		log.Println(err)
		return
	}
	defer client.stop()

	s.sendWelcome(client)

	log.Printf("%s Connected", client.User().Username())

	s.handleMessages(client)

	log.Printf("%s Disconnected", client.User().Username())

	room := client.Room()

	if room != nil {
		room.Remove(client)
		room.broadcastUserList()

		room.Broadcast(protocol.Message{
			Type:      "leave",
			Username:  client.User().Username(),
			Nickname:  client.User().Nickname(),
			Timestamp: time.Now(),
		}, nil)
	}

	s.removeClient(client)
}

func (s *Server) removeConnection(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.conns, conn)
}

func (s *Server) closeClients() {
	for _, conn := range s.connSnapshot() {
		conn.Close()
	}
}

func (s *Server) handleMessages(client *Client) {
	for {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Minute))

		msg, err := client.readMessage()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Println(err)
			}
			return
		}

		if err := protocol.ValidateMessage(&msg); err != nil {
			if err := client.Send(protocol.Message{
				Type:    "error",
				Message: err.Error(),
			}); err != nil {
				log.Println(err)
			}
			continue
		}

		switch msg.Type {
		case "chat":
			if !s.messageRateLimit.Allow(client.User().Username()) {
				if err := client.Send(protocol.Message{
					Type:    "error",
					Message: "You are sending messages too fast.",
				}); err != nil {
					log.Println(err)
				}
				continue
			}
			s.broadcastMessage(msg.Message, client)
		case "command":
			if s.executeCommand(client, msg.Message) {
				return
			}
		case "pong":
			client.markPong()
		default:
			if err := client.Send(protocol.Message{
				Type:    "error",
				Message: "Unknown message type: " + msg.Type,
			}); err != nil {
				log.Println(err)
			}
		}
	}
}

func (s *Server) requireServerPassword(client *Client) error {
	if err := client.Send(protocol.Message{
		Type:    "server_password_prompt",
		Message: "Server password required",
	}); err != nil {
		return err
	}

	attempts := 0
	for {
		msg, err := client.readMessage()
		if err != nil {
			return err
		}

		if msg.Type != "server_password" {
			if err := client.Send(protocol.Message{
				Type:    "error",
				Message: "Expected server password",
			}); err != nil {
				return err
			}
			continue
		}

		if s.CheckPassword(msg.Password) {
			return nil
		}

		attempts++
		if attempts >= maxAuthAttempts {
			return errors.New("too many auth attempts")
		}

		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: "Incorrect server password",
		}); err != nil {
			return err
		}
	}
}
