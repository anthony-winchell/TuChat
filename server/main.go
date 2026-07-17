package main

import (
	"log"
	"net"
)

type Client struct {
	conn     net.Conn
	username string
}

type Server struct {
	listener net.Listener
	clients  []*Client
}

func (s *Server) Start() {
	defer s.listener.Close()
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
	if _, err := conn.Write([]byte("Connection Successful\n Enter Username:")); err != nil {
		return nil, err
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	username := string(buffer[:n-1])

	client:= &Client{
		conn:     conn,
		username: username,
	}

	s.clients = append(s.clients, client)

	return client, nil
}

func (s *Server) handleMessages(client *Client) {
	buffer := make([]byte, 1024)
	for {
		n, err := client.conn.Read(buffer)
		if err != nil {
			log.Println(err)
			return
		}

		message := string(buffer[:n])
		log.Printf("%s: %s", client.username, message)
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

	server.Start()

}
