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

func (s *Server) broadcastMessage(text string, sender *Client) {
	sender.Room().Broadcast(protocol.Message{
		Type:    "chat",
		Username: sender.User().Username(),
		Message: text,
	}, sender)
}

func (s *Server) leaveAlert(leaver *Client) {
	leaver.Room().Broadcast(protocol.Message{
		Type:     "leave",
		Username: leaver.User().Username(),
	}, leaver)
}