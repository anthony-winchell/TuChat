package main

import (
	"log"
	"net"
	"sync"
	"bufio"
	"strings"
)

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
			log.Println(err)
			return
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
}

func (s *Server) registerClient(conn net.Conn) (*Client, error) {
	if _, err := conn.Write([]byte("Choose A Username:\n")); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)

	username, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	 }

	username = strings.TrimSpace(username)

	client:= &Client{
		conn:     conn,
		username: username,
		reader:   reader,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients = append(s.clients, client)

	_, err = conn.Write([]byte("Thank you " + username + ". You may begin chatting.\n"))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (s *Server) handleMessages(client *Client) {
	for {
		message, err := client.reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}
		s.broadcastMessage(message, client)
	}
}

func (s *Server) broadcastMessage(message string, sender *Client) {

	log.Println("broadcasting: ", message)
	log.Println("Clients: ", len(s.clients))

	formattedMessage := sender.username + ": " + message
	s.mu.Lock()

	clients := make([]*Client, len(s.clients))
	copy(clients, s.clients)

	s.mu.Unlock()
	for _, client := range clients {
		if client == sender {
			continue
		}
		if _, err := client.conn.Write([]byte(formattedMessage)); err != nil {
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
