package main

import (
	"log"
	"tuchat/protocol"
)


func (s *Server) sendToAll(message protocol.Message, except *Client) {
	for _, client := range s.clientsSnapshot() {
		if client == except {
			continue
		}
		if err := client.Send(message); err != nil {
			log.Println(err)
		}
	}
}

func (s *Server) sendToRoom(message protocol.Message, sender *Client) {
	clients := s.roomSnapshot(sender.room)

	for _, client := range clients {
		if client == sender {
			continue
		}
		if err := client.Send(message); err != nil {
			log.Println(err)
		}
	}
}

func (s *Server) broadcastMessage(text string, sender *Client) {
	s.sendToRoom(protocol.Message{
		Type:     "chat",
		Username: sender.username,
		Message:  text,
	}, sender)
}

func (s *Server) joinAlert(joiner *Client) {
	s.sendToRoom(protocol.Message{
		Type:     "join",
		Username: joiner.username,
	}, joiner)
}

func (s *Server) leaveAlert(leaver *Client) {
	s.sendToRoom(protocol.Message{
		Type:     "leave",
		Username: leaver.username,
	}, leaver)
}