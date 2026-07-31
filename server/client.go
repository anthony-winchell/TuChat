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

func (c *Client) Username() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.username
}
