package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
	"tuchat/protocol"
)

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

func (s *Server) sendWelcome(client *Client) {
	if err := client.Send(protocol.Message{
		Type: "welcome",
		Message: fmt.Sprintf(
			"%s\nWelcome to %s, %s!\n\nType /help for commands.\n%s",
			strings.Repeat("=", 35),
			s.name,
			client.username,
			strings.Repeat("=", 35),
		),
	}); err != nil {
		log.Println(err)
	}
}

func (s *Server) configureName() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Name Your Server (leave blank for 'TuChat'):")

	name, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
		return
	}

	name = strings.TrimSpace(name)

	if name != "" {
		s.name = name
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

	log.Printf("%s Connected", client.username)

	s.handleMessages(client)

	log.Printf("%s Disconnected", client.username)

	room := client.Room()

	s.removeClient(client)

	s.leaveAlert(client)

	if room != nil {
		room.Remove(client)
	}
}

func (s *Server) removeConnection(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.conns, conn)
}

func (s *Server) registerClient(conn net.Conn) (*Client, error) {
	var client *Client
	client = &Client{
		conn:    conn,
		decoder: json.NewDecoder(conn),
		encoder: json.NewEncoder(conn),
	}

	if err := client.Send(protocol.Message{
		Type:    "username_prompt",
		Message: "Choose a username:",
	}); err != nil {
		return nil, err
	}

	for {
		var message protocol.Message

		if err := client.decoder.Decode(&message); err != nil {
			return nil, err
		}

		if message.Type != "username" {
			continue
		}

		client.username = strings.TrimSpace(message.Username)

		err := s.addClient(client)

		if err != nil {
			if err := client.Send(protocol.Message{
				Type:    "error",
				Message: err.Error(),
			}); err != nil {
				return nil, err
			}
			if err := client.Send(protocol.Message{
				Type:    "username_prompt",
				Message: "Choose a username:",
			}); err != nil {
				return nil, err
			}
			continue
		}

		if err := client.Send(protocol.Message{
			Type: "username_accepted",
		}); err != nil {
			return nil, err
		}

		break
	}
	s.JoinRoom(client, "general", "")

	return client, nil
}

func (s *Server) addClient(client *Client) error {
	if strings.ContainsRune(client.username, ':') {
		return ErrUsernameFormat
	}

	if len(client.username) < 3 || len(client.username) > 13 {
		return ErrUsernameFormat
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[client.username]; exists {
		return ErrUsernameTaken
	}

	s.clients[client.username] = client

	return nil
}

func (s *Server) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, client.username)
}

func (s *Server) findClient(username string) *Client {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.clients[username]
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

func (s *Server) RoomsSnapshot() []*Room {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rooms := make([]*Room, 0, len(s.rooms))

	for _, room := range s.rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

func (s *Server) RenameRoom(r *Room, newName string)  error {
	newName = strings.TrimSpace(newName)

	if newName == "" {
		return errors.New("room name cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rooms[newName]; exists {
		return errors.New("room exists")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(s.rooms, r.name)

	r.name = newName

	s.rooms[newName] = r

	return nil
}

func (s *Server) Name() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.name
}

func (s *Server) SetName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.name = name
}

func (s *Server) AddRoom(room *Room) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rooms[room.Name()] = room
}