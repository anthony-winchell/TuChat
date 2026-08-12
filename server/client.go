package main

import (
	"log"
	"tuchat/protocol"
	"strings"
)

func (c *Client) Send(msg protocol.Message) error {

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

func (c *Client) Room() *Room {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.room
}

func (c *Client) User() *User {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.user
}

func (c *Client) SetUser(user *User) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.user = user
}

func (c *Client) SetRoom(room *Room) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.room = room
}

func (s *Server) addClient(client *Client) error {
	if strings.ContainsRune(client.User().Username(), ':') {
		return ErrUsernameFormat
	}

	if len(client.User().Username()) < 3 || len(client.User().Username()) > 13 {
		return ErrUsernameFormat
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[client.User().Username()]; exists {
		return ErrUsernameTaken
	}

	s.clients[client.User().Username()] = client

	return nil
}

func (s *Server) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, client.User().Username())
}

func (s *Server) findClient(username string) *Client {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.clients[username]
}

func (s *Server) findClientByNickname(nickname string) *Client {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		if strings.EqualFold(client.User().Nickname(), nickname) {
			return client
		}
	}

	return nil
}
