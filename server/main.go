package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
	"tuchat/protocol"
)

var ErrUsernameTaken = errors.New("Username already taken")
var ErrUsernameFormat = errors.New(`Username cannot contain ':'. Must be between 3 and 13 characters long`)

type Client struct {
	conn     net.Conn
	username string

	decoder  *json.Decoder
	encoder  *json.Encoder
	writeMu  sync.Mutex
}

type Server struct {
	mu       sync.RWMutex
	listener net.Listener

	name string

	clients  map[string]*Client
	conns  map[net.Conn]struct{}

	wg       sync.WaitGroup
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

func (s *Server) Shutdown() {
	if err := s.listener.Close(); err != nil { 
		log.Println(err) 
	}

	s.sendToAll(protocol.Message{
		Type: "system",
		Message: "Server shutting down",
	},nil)

	s.closeClients()

	s.wg.Wait()
}

func (s *Server) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, client.username)
}

func (s *Server) closeClients() {
	for _, conn := range s.connSnapshot() {
		conn.Close()
	}
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

	s.removeClient(client)

	s.leaveAlert(client)
}

func (s *Server) addConnection(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conns[conn] = struct{}{}
}

func (s *Server) removeConnection(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.conns, conn)
} 

func (s *Server) registerClient(conn net.Conn) (*Client, error) {
	var client *Client
	client = &Client{
			conn:     conn,
			decoder:  json.NewDecoder(conn),
			encoder:  json.NewEncoder(conn),
		}

	if err := client.Send(protocol.Message{
		Type: "username_prompt",
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
		if errors.Is(err, ErrUsernameTaken) {
			if err := client.Send(protocol.Message{
				Type: "error",
				Message: ErrUsernameTaken.Error(),
			}); err != nil {
				return nil, err
			}
			continue
		}
		if errors.Is(err, ErrUsernameFormat) {
			if err := client.Send(protocol.Message{
				Type: "error",
				Message: ErrUsernameFormat.Error(),
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

	s.joinAlert(client)

	return client, nil
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
				Type: "error",
				Message: "Unknown message type: " + msg.Type,
			}); err != nil {
				log.Println(err)
			}
		}
	}
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

func (s *Server) broadcastMessage(text string, sender *Client) {
	s.sendToAll(protocol.Message{
		Type: "chat",
		Username: sender.username,
		Message: text,
	}, sender)
}

func (s *Server) joinAlert(joiner *Client) {
	s.sendToAll(protocol.Message{
		Type: "join",
		Username: joiner.username,
	}, joiner)
}

func (s *Server) leaveAlert(leaver *Client) {
	s.sendToAll(protocol.Message{
		Type: "leave",
		Username: leaver.username,
	}, leaver)
}

func (s *Server) clientsSnapshot() []*Client {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]*Client, 0, len(s.clients))

	for _, client := range s.clients {
		clients = append(clients, client)
	}

	return clients
}

func (s *Server) connSnapshot() []net.Conn {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conns := make([]net.Conn, 0, len(s.conns))

	for conn := range s.conns {
		conns = append(conns, conn)
	}
	return conns
}

func (s *Server) sendToAll(message protocol.Message, except *Client) {
	for _, client := range s.clientsSnapshot() {
		if client == except {
			continue
		}
		if err := client.Send(message); err != nil {
			log.Println(err)
		}
	}
}

func (s *Server) executeCommand(client *Client, input string) bool {
	parts := strings.Fields(input)

	if len(parts) == 0 {
		return false
	}

	switch parts[0] {
	case "/quit":
		return s.commandQuit(client)
	case "/users":
		return s.commandUsers(client)
	case "/help":
		return s.commandHelp(client)
	case "/pm":
		return s.commandPM(client, parts)
	default:
		if err := client.Send(protocol.Message{
			Type: "error",
			Message: "Unknown command: " + parts[0],
		}); err != nil {
			log.Println(err)
		}
		return false
	}
}

func (s *Server) commandPM(client *Client, parts []string) bool {
	if len(parts) < 3{
			if err := client.Send(protocol.Message{
				Type: "error",
				Message: "Usage: /pm <username> <message>",
			}); err != nil {
				log.Println(err)
			}
			return false
		}

		receiver := s.findClient(parts[1])

		if receiver == nil {
			if err := client.Send(protocol.Message{
				Type: "error",
				Message: "User not found: " + parts[1],
			}); err != nil {
				log.Println(err)
			}
			return false
		}

		if receiver == client {
			if err := client.Send(protocol.Message{
				Type: "error",
				Message: "You cannot /pm yourself",
			}); err != nil {
				log.Println(err)
			}
			return false
		}

		message := strings.Join(parts[2:], " ")

		if err :=receiver.Send(protocol.Message{
			Type: "pm",
			Username: client.username, 
			Target: parts[1],
			Message: message,
		}); err != nil {
			log.Println(err)
		}

		if err := client.Send(protocol.Message{
			Type: "pm",
			Username: client.username,
			Target: parts[1],
			Message: message,
		}); err != nil {
			log.Println(err)
		}
		return false
}

func (s *Server) commandHelp(client *Client) bool {
	if err := client.Send(protocol.Message{
		Type: "system",
		Message: "Commands: /quit, /users, /pm <username> <message>",
	}); err != nil {
		log.Println(err)
	}

	return false
}

func (s *Server) commandQuit(client *Client) bool {
	if err :=client.Send(protocol.Message{
		Type: "system",
		Message: "Goodbye",
	}); err != nil {
		log.Println(err)
	}
	return true
}

func (s *Server) commandUsers(client *Client) bool {
	users := s.clientsSnapshot()
	usernames := make([]string, 0, len(users))

	for _, user := range users {
		usernames = append(usernames, user.username)
	}

	if err := client.Send(protocol.Message{
		Type: "users",
		Users: usernames,
	}); err != nil {
		log.Println(err)
	}

	return false
}

func (s *Server) findClient(username string) *Client {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.clients[username]
}

func (c *Client) Send(msg protocol.Message) error  {

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.encoder.Encode(msg)
}

func (c *Client) Close() {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if err := c.conn.Close(); err != nil {
		log.Println(err)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Println(err)
		return 
	}

	server := &Server{
		name:     "TuChat",
		listener: listener,
		clients:  make(map[string]*Client),
		conns: make(map[net.Conn]struct{}),
	}

	server.configureName()

	log.Println("Starting Server...")
	go server.Start()

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, os.Interrupt)

	<-signals

	server.Shutdown()

}
