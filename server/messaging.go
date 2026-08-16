package main

import (
	"log"
	"time"
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

	time := time.Now()

	room := sender.Room()

	chatlog, err := s.getChatLog(room.Name())
	if err != nil {
		log.Println(err)
		return
	}

	err = chatlog.Write(LogEntry{
		Username:  sender.User().Username(),
		Nickname:  sender.User().Nickname(),
		Message:   text,
		Timestamp: time,
	})
	if err != nil {
		log.Println("Failed to write to chatlog:", err)
	}

	room.Broadcast(protocol.Message{
		Type:      "chat",
		Username:  sender.User().Username(),
		Nickname:  sender.User().Nickname(),
		Message:   text,
		Timestamp: time,
	}, nil)
}

