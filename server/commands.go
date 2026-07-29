package main 

import (
	"tuchat/protocol"
	"log"
	"strings"
)

func (s *Server) commandUsers(client *Client) bool {
	users := s.roomSnapshot(client.room)
	usernames := make([]string, 0, len(users))

	for _, user := range users {
		usernames = append(usernames, user.username)
	}

	if err := client.Send(protocol.Message{
		Type:  "users",
		Users: usernames,
	}); err != nil {
		log.Println(err)
	}

	return false
}

func (s *Server) commandQuit(client *Client) bool {
	if err := client.Send(protocol.Message{
		Type:    "system",
		Message: "Goodbye",
	}); err != nil {
		log.Println(err)
	}
	return true
}

func (s *Server) commandHelp(client *Client) bool {
	if err := client.Send(protocol.Message{
		Type:    "system",
		Message: "Commands: /quit, /users, /pm <username> <message>",
	}); err != nil {
		log.Println(err)
	}

	return false
}

func (s *Server) commandPM(client *Client, parts []string) bool {
	if len(parts) < 3 {
		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: "Usage: /pm <username> <message>",
		}); err != nil {
			log.Println(err)
		}
		return false
	}

	receiver := s.findClient(parts[1])

	if receiver == nil {
		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: "User not found: " + parts[1],
		}); err != nil {
			log.Println(err)
		}
		return false
	}

	if receiver == client {
		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: "You cannot /pm yourself",
		}); err != nil {
			log.Println(err)
		}
		return false
	}

	message := strings.Join(parts[2:], " ")

	if err := receiver.Send(protocol.Message{
		Type:     "pm",
		Username: client.username,
		Target:   parts[1],
		Message:  message,
	}); err != nil {
		log.Println(err)
	}

	if err := client.Send(protocol.Message{
		Type:     "pm",
		Username: client.username,
		Target:   parts[1],
		Message:  message,
	}); err != nil {
		log.Println(err)
	}
	return false
}

func (s *Server) executeCommand(client *Client, input string) bool {
	parts := strings.Fields(input)

	if len(parts) == 0 {
		return false
	}

	switch parts[0] {
	case "/quit":
		return s.commandQuit(client)
	case "/users":
		return s.commandUsers(client)
	case "/help":
		return s.commandHelp(client)
	case "/pm":
		return s.commandPM(client, parts)
	default:
		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: "Unknown command: " + parts[0],
		}); err != nil {
			log.Println(err)
		}
		return false
	}
}
