package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"time"
	"tuchat/protocol"
)

const outboxSize = 64
const writeTimeout = 10 * time.Second

func newClient(conn net.Conn) *Client {
	c := &Client{
		conn:     conn,
		input:    bufio.NewScanner(conn),
		encoder:  json.NewEncoder(conn),
		outbox:   make(chan protocol.Message, outboxSize),
		done:     make(chan struct{}),
		lastPong: time.Now(),
	}

	c.input.Buffer(make([]byte, 64*1024), maxMessageSize)
	return c
}

func (c *Client) startWriter(wg *sync.WaitGroup) {
	wg.Go(func() {
		for {
			select {
			case <-c.done:
				return
			case msg := <-c.outbox:
				c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))

				if err := c.encoder.Encode(msg); err != nil {
					c.stop()
					return
				}
			}
		}
	})
}

func (c *Client) markPong() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastPong = time.Now()
}

func (c *Client) lastPongTime() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastPong
}

func (c *Client) stop() {
	c.stopOnce.Do(func() {
		close(c.done)
		c.conn.Close()
	})
}

func (c *Client) Send(msg protocol.Message) error {
	select {
	case <-c.done:
		return errors.New("connection closed")

	default:
	}
	select {
	case c.outbox <- msg:
		return nil

	default:
		go c.stop()
		return errors.New("outbox full; client evicted")
	}
}

func (c *Client) readMessage() (protocol.Message, error) {
	var msg protocol.Message

	if !c.input.Scan() {
		if err := c.input.Err(); err != nil {
			return msg, err
		}
		return msg, io.EOF
	}

	if err := json.Unmarshal(c.input.Bytes(), &msg); err != nil {
		return msg, err
	}

	return msg, nil
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

	return s.clients[strings.ToLower(username)]
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
