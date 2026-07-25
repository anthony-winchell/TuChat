package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"strings"
	"sync"
	"time"
	"os"
	"os/signal"
	"fmt"
)

var ErrUsernameTaken = errors.New("Username already taken")
var ErrUsernameFormat = errors.New(`Username cannot contain ':'. Must be between 3 and 13 characters long`)

type Client struct {
	conn     net.Conn
	username string
	reader   *bufio.Reader
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
	client.Send(fmt.Sprintf(
		"%s\n"+
			"Welcome to %s, %s!\n\n"+
			`Type "/help" for a list of commands.`+"\n"+
			"%s",
		strings.Repeat("=", 35),
		s.name,
		client.username,
		strings.Repeat("=", 35),
	))
}

func (s *Server) Shutdown() {
	if err := s.listener.Close(); err != nil { 
		log.Println(err) 
	}

	s.sendToAll("Server shutting down.", nil)

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

	reader := bufio.NewReader(conn)
	var client *Client
	if _, err := conn.Write([]byte("Choose a Username:\n")); err != nil {
		return nil, err
	}
	for {
		username, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		username = strings.TrimSpace(username)

		if username == "" {
			if _, err := conn.Write([]byte("Username cannot be empty. Please choose another:\n")); err != nil {
				return nil, err
			}
			continue
		}

		client = &Client{
			conn:     conn,
			username: username,
			reader:  reader,
		}

		err = s.addClient(client)

		if errors.Is(err, ErrUsernameFormat) {
			if _, err := conn.Write([]byte(ErrUsernameFormat.Error() + " Please choose another:\n")); err != nil {
				return nil, err
			}
			continue
		}

		if errors.Is(err, ErrUsernameTaken) {
			if _, err := conn.Write([]byte("Username already taken. Please choose another:\n")); err != nil {
				return nil, err
			}
			continue
		}

		break
	}

	s.joinAlert(client)

	return client, nil
}

func (s *Server) handleMessages(client *Client) {
	for {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Minute))

		message, err := client.reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		message = strings.TrimSpace(message)

		if message == "" {
			continue
		}
		if strings.HasPrefix(message, "/") {
			if s.executeCommand(client, message) {
				return
			}
			continue
		}
		s.broadcastMessage(message, client)
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

func (s *Server) broadcastMessage(message string, sender *Client) {
	formattedMessage := sender.username + ": " + strings.TrimSpace(message)

	s.sendToAll(formattedMessage, sender)
}

func (s *Server) joinAlert(joiner *Client) {
	s.sendToAll(joiner.username+" has joined the chat.", joiner)
}

func (s *Server) leaveAlert(leaver *Client) {
	s.sendToAll(leaver.username+" has left the chat.", leaver)
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

func (s *Server) sendToAll(message string, except *Client) {
	for _, client := range s.clientsSnapshot() {
		if client == except {
			continue
		}
		client.Send(message)
	}
}

func (s *Server) executeCommand(client *Client, input string) bool {
	parts := strings.Fields(input)

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
		client.Send("Invalid Command. Try /help.")
		return false
	}
}

func (s *Server) commandPM(client *Client, parts []string) bool {
	if len(parts) < 3{
			client.Send("Usage: /pm <username> <message>")
			return false
		}

		receiver := s.findClient(parts[1])

		if receiver == nil {
			client.Send("User not found.")
			return false
		}

		if receiver == client {
			client.Send("You cannot /pm yourself.")
			return false
		}

		message := strings.Join(parts[2:], " ")

		receiver.Send("(Private) " + client.username + ": " + message)

		client.Send("(To " + receiver.username + ") " + message)
		return false
}

func (s *Server) commandHelp(client *Client) bool {
	client.Send("Available Commands:")
		client.Send("/users")
		client.Send("/help")
		client.Send("/quit")
		client.Send("/pm <username> <message>")

		return false
}

func (s *Server) commandQuit(client *Client) bool {
	client.Send("Goodbye.")
	return true
}

func (s *Server) commandUsers(client *Client) bool {
	users := s.clientsSnapshot()

	client.Send("Active users:")

	for _, user := range users {
		client.Send(user.username)
	}

	return false
}

func (s *Server) findClient(username string) *Client {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.clients[username]
}

func (c *Client) Send(message string) {

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if _, err := c.conn.Write([]byte(message + "\n")); err != nil {
		log.Println(err)
	}
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
