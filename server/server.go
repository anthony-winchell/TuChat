package main

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"time"
	"tuchat/protocol"
)

const defaultWelcomeMessage = "Welcome to {server}, {nickname}!\n\nType /help for commands."

func (s *Server) Start() {
	log.Println("Server started")
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
			s.handleConnection(conn)
		})
	}
}

func (s *Server) Shutdown() {
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

func (s *Server) addConnection(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conns[conn] = struct{}{}
}

func (s *Server) registerClient(conn net.Conn) (*Client, error) {
	var client *Client
	client = &Client{
		conn:    conn,
		decoder: json.NewDecoder(conn),
		encoder: json.NewEncoder(conn),
	}

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

	for {
		var message protocol.Message

		if err := client.decoder.Decode(&message); err != nil {
			return nil, err
		}

		switch message.Type {
		case "register":
			user, err := s.RegisterUser(
				message.Username,
				message.Password,
			)

			if err != nil {
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
				Type: "auth_success",
				Nickname: client.User().Nickname(),
			}); err != nil {
				return nil, err
			}
			break
		}
	}

	if err := client.Send(protocol.Message{
		Type: "server_name",
		Message: s.Name(),
	}); err != nil {
		return nil, err
	}

	if err := s.JoinRoom(client, "general", ""); err != nil {
		return nil, err
	}

	return client, nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		s.removeConnection(conn)
		conn.Close()
	}()

	client, err := s.registerClient(conn)
	if err != nil {
		log.Println(err)
		return
	}

	s.sendWelcome(client)

	log.Printf("%s Connected", client.User().Username())

	s.handleMessages(client)

	log.Printf("%s Disconnected", client.User().Username())

	room := client.Room()

	if room != nil {
		room.Remove(client)
		room.broadcastUserList()

		room.Broadcast(protocol.Message{
			Type: "leave",
			Username: client.User().Username(),
			Nickname: client.User().Nickname(),
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

		var msg protocol.Message

		if err := client.decoder.Decode(&msg); err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Println(err)
			}
			return
		}

		switch msg.Type {
		case "chat":
			s.broadcastMessage(msg.Message, client)
		case "command":
			if s.executeCommand(client, msg.Message) {
				return
			}
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

	for {
		var message protocol.Message
		if err := client.decoder.Decode(&message); err != nil {
			return err
		}

		if message.Type != "server_password" {
			if err := client.Send(protocol.Message{
				Type:    "error",
				Message: "Expected server password",
			}); err != nil {
				return err
			}
			continue
		}

		if s.CheckPassword(message.Password) {
			return nil
		}

		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: "Incorrect server password",
		}); err != nil {
			return err
		}
	}
}
