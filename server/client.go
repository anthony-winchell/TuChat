package main

import (
	"log"
	"tuchat/protocol"
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
