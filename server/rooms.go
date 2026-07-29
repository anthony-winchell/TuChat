package main

import (
	"errors"
)

func (s *Server) JoinRoom(client *Client, roomName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	room, ok := s.rooms[roomName]
	if !ok {
		return errors.New("Room not found")
	}

	room.clients[client.username] = client

	client.mu.Lock()
	client.room = room
	client.mu.Unlock()

	return nil
}

func (s *Server) leaveRoom(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client.room == nil {
		return
	}

	delete(client.room.clients, client.username)
	client.room = nil
}