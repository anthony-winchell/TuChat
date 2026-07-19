package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

var ErrUsernameTaken = errors.New("Username already taken")
var ErrUsernameFormat = errors.New(`Username cannot contain ':', '\n' or spaces. Must be between 3 and 13 characters long`)

type Client struct {
	conn     net.Conn
	username string
	reader   *bufio.Reader
}

type Server struct {
	mu       sync.Mutex
	listener net.Listener
	clients  []*Client
}

func (s *Server) Start() {
	defer s.listener.Close()
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
		go s.handleConnection(conn)
	}
}

func (s *Server) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, c := range s.clients {
		if c == client {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			return
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	client, err := s.registerClient(conn)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("%s Connected", client.username)

	s.handleMessages(client)

	log.Printf("%s Disconnected", client.username)

	s.removeClient(client)

	s.leaveAlert(client)
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

	_, err := conn.Write([]byte("Thank you " + client.username + ". You may begin chatting.\n"))
	if err != nil {
		return nil, err
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
		s.broadcastMessage(message, client)
	}
}

func (s *Server) addClient(client *Client) error {
	if strings.ContainsAny(client.username, `: \n`) {
		return ErrUsernameFormat
	}

	if len(client.username) < 3 || len(client.username) > 13 {
		return ErrUsernameFormat
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, c := range s.clients {
		if c.username == client.username {
			return ErrUsernameTaken
		}
	}

	s.clients = append(s.clients, client)

	return nil
}

func (s *Server) broadcastMessage(message string, sender *Client) {
	formattedMessage := sender.username + ": " + strings.TrimSpace(message) + "\n"

	s.sendToAll(formattedMessage, sender)
}

func (s *Server) joinAlert(joiner *Client) {
	s.sendToAll(joiner.username+" has joined the chat.\n", joiner)
}

func (s *Server) leaveAlert(leaver *Client) {
	s.sendToAll(leaver.username+" has left the chat.\n", leaver)
}

func (s *Server) clientsSnapshot() []*Client {
	s.mu.Lock()
	defer s.mu.Unlock()

	clients := make([]*Client, len(s.clients))
	copy(clients, s.clients)

	return clients
}

func (s *Server) sendToAll(message string, except *Client) {
	for _, client := range s.clientsSnapshot() {
		if client == except {
			continue
		}
		if _, err := client.conn.Write([]byte(message)); err != nil {
			log.Println(err)
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Println(err)
		return
	}

	server := &Server{
		listener: listener,
	}
	log.Println("Starting Server...")
	server.Start()
}
