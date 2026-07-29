package main

import (
	"tuchat/protocol"
	"log"
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