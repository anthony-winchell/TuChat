package main

import (
	"fmt"
	"log"
	"strings"
	"tuchat/protocol"
)

func (s *Server) commandUsers(client *Client) bool {
	users := client.Room().Users()
	usernames := make([]string, 0, len(users))

	for _, user := range users {
		usernames = append(usernames, user.User().Username())
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
		Type: "system",
		Message: fmt.Sprint(
			"Commands:\n",
			"/pm <username> <message>\n",
			"/join <room>\n",
			"/rooms\n",
			"/room\n",
			"/quit\n",
			"/help\n",
			"/users\n",
			"/topic\n",
			"/settopic <topic>\n",
			"/setpassword <password>\n",
		),
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
		Username: client.User().Username(),
		Target:   parts[1],
		Message:  message,
	}); err != nil {
		log.Println(err)
	}

	if err := client.Send(protocol.Message{
		Type:     "pm",
		Username: client.User().Username(),
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
	case "/join":
		if len(parts) < 2 {
			if err := client.Send(protocol.Message{
				Type:    "error",
				Message: "Usage: /join <room> [password]",
			}); err != nil {
				log.Println(err)
			}
			return false
		}

		password := ""

		if len(parts) >= 3 {
			password = strings.Join(parts[2:], " ") 
		}
		s.commandJoinRoom(client, parts[1], password)
		return false
	case "/rooms":
		return s.commandRooms(client)
	case "/room":
		return s.commandRoom(client)
	case "/rename":
		if len(parts) < 2 {
			if err := client.Send(protocol.Message{
				Type:    "error",
				Message: "Usage: /rename <room>",
			}); err != nil {
				log.Println(err)
			}
			return false
		}
		return s.commandRenameRoom(parts[1], client)
	case "/topic":
		return s.commandTopic(client)
	case "/settopic":
		if len(parts) < 2 {
			if err := client.Send(protocol.Message{
				Type:    "error",
				Message: "Usage: /settopic <topic>",
			}); err != nil {
				log.Println(err)
			}
			return false
		}
		topic := strings.Join(parts[1:], " ")
		s.commandSetTopic(client, topic)
		return false

	case "/promote":
		if len(parts) < 2 {
			if err := client.Send(protocol.Message{
				Type:    "error",
				Message: "Usage: /promote <username>",
			}); err != nil {
				log.Println(err)
			}
			return false
		}
		s.commandPromote(client, parts[1])
		return false
	case "/demote":
		if len(parts) < 2 {
			if err := client.Send(protocol.Message{
				Type:    "error",
				Message: "Usage: /demote <username>",
			}); err != nil {
				log.Println(err)
			}
			return false
		}
		s.commandDemote(client, parts[1])
		return false

	case "/setpassword": 
		password := ""

		if len(parts) >= 2 {
			password = strings.Join(parts[1:], " ")
		}

		return s.commandSetPassword(client, password)
		
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

func (s *Server) commandJoinRoom(client *Client, roomName string, password string) bool {
	if err := s.JoinRoom(client, roomName, password); err != nil {
		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: err.Error(),
		}); err != nil {
			log.Println(err)
		}
		return false
	}

	return false
}

func (s *Server) commandRooms(client *Client) bool {
	rooms := s.RoomsSnapshot()

	if err := client.Send(protocol.Message{
		Type:    "system",
		Message: "Rooms:",
	}); err != nil {
		log.Println(err)
	}

	for _, room := range rooms {
		if err := client.Send(
			protocol.Message{
				Type:    "system",
				Message: fmt.Sprintf("%s: %d users", room.Name(), room.Size()),
			},
		); err != nil {
			log.Println(err)
		}
	}

	return false
}

func (s *Server) commandRoom(client *Client) bool {

	room := client.Room()
	 
	if err := client.Send(protocol.Message{
		Type:    "system",
		Message: "Current Room: " + room.Name(),
	}); err != nil {
		log.Println(err)
	}

	if err := client.Send(protocol.Message{
		Type:    "system",
		Message: "Topic: " + room.Topic(),
	}); err != nil {
		log.Println(err)
	}

	if err := client.Send(protocol.Message{
		Type: "system",
		Message: fmt.Sprintf("Users: %d", room.Size()),
	}); err != nil {
		log.Println(err)
	}
	

	return false
}

func (s *Server) commandRenameRoom(name string, client *Client) bool {

	room := client.Room()

	if err := room.RequireOperator(client); err != nil {
		if err := client.Send(protocol.Message{
			Type: "system",
			Message: err.Error(),
		}); err != nil {
			log.Println(err)
		}
	
		return false
	}

	if err := s.RenameRoom(room, name); err != nil {
		if err := client.Send(protocol.Message{
			Type: "system",
			Message: err.Error(),
		}); err != nil {
			log.Println(err)
		}
		return false
	}
	if err := s.SaveConfig(); err != nil {
		log.Println("Failed to save config: " + err.Error())
	}
	return false
}

func (s *Server) commandTopic(client *Client) bool {
	room := client.Room()

	if err := client.Send(protocol.Message{
		Type:    "system",
		Message: "Topic: " + room.Topic(),
	}); err != nil {
		log.Println(err)
	}

	return false
}

func (s *Server) commandSetTopic(client *Client, topic string)  {
	room := client.Room()

	if err := room.RequireOperator(client); err != nil {
		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: err.Error(),
		}); err != nil {
			log.Println(err)
		}
		return
	}

	room.SetTopic(topic)
	if err := s.SaveConfig(); err != nil {
		log.Println("Failed to save config: " + err.Error())
	}

	room.Broadcast(protocol.Message{
		Type: "system",
		Message: client.User().Username() + " changed the topic to: " + topic,

	}, nil)
}

func (s *Server) commandPromote(client *Client, targetUsername string) bool {

	target := s.findClient(targetUsername)

	if target == nil {
		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: "User not found: " + targetUsername,
		}); err != nil {
			log.Println(err)
		}
		return false
	}

	room := client.Room()

	if err := room.RequireOperator(client); err != nil {
		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: err.Error(),
		}); err != nil {
			log.Println(err)
		}
		return false
	}

	if err := room.Promote(target); err != nil {
		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: err.Error(),
		}); err != nil {
			log.Println(err)
		}
		return false
	}

	room.Broadcast(protocol.Message{
		Type: "system",
		Message: client.User().Username() + " promoted " + target.User().Username() + " to operator",
	}, nil)

	return false
}

func (s *Server) commandDemote(client *Client, targetUsername string) bool {

	target := s.findClient(targetUsername)
	if target == nil {
		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: "User not found: " + targetUsername,
		}); err != nil {
			log.Println(err)
		}
		return false
	}

	room := client.Room()

	if err := room.RequireOperator(client); err != nil {
		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: err.Error(),
		}); err != nil {
			log.Println(err)
		}
		return false
	}

	if err := room.Demote(target); err != nil {
		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: err.Error(),
		}); err != nil {
			log.Println(err)
		}
		return false
	}

	room.Broadcast(protocol.Message{
		Type: "system",
		Message: client.User().Username() + " demoted " + target.User().Username() + " to user",
	}, nil)

	return false
}

func (s *Server) commandSetPassword(client *Client, password string) bool {
	room := client.Room()

	if err := room.RequireOperator(client); err != nil {
		if err := client.Send(protocol.Message{
			Type: "error",
			Message: err.Error(),
		}); err != nil {
			log.Println(err)
		}
		return false
	}

	if err := room.SetPassword(password); err != nil {
		if err := client.Send(protocol.Message{
			Type: "error",
			Message: err.Error(),
		}); err != nil {
			log.Println(err)
		}
		return false
	} else {
		if err := s.SaveConfig(); err != nil {
			log.Println("Failed to save config: " + err.Error())
		}
		room.Broadcast(protocol.Message{
			Type: "system",
			Message: client.User().Username() + " set a password",
		}, nil)
	}

	return false
}