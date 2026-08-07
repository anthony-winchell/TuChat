package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"tuchat/protocol"
)

type Command struct {
	Description string
	Handler     CommandFunc
	Usage       string
}

type CommandFunc func(*Server, *Client, []string) bool

func (s *Server) executeCommand(client *Client, input string) bool {
	parts := strings.Fields(input)

	if len(parts) == 0 {
		return false
	}

	command := parts[0]
	args := parts[1:]

	cmd, exists := s.Commands()[command]

	if !exists {
		sendError(client, "Unknown command: "+command)
		return false
	}

	return cmd.Handler(s, client, args)
}

func (s *Server) commandHelp(client *Client) bool {

	message := "Commands:\n"

	for name, command := range s.Commands() {
		message += fmt.Sprintf("%s - %s\n   Usage: %s\n",
		name,
		command.Description,
		command.Usage,
	)
	}

	sendSystem(client, message)

	return false
}

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
	sendSystem(client, "Goodbye!")
	return true
}

func (s *Server) commandPM(client *Client, args []string) bool {
	if len(args) < 2 {
		sendError(client, "Usage: /pm <username> <message>")
		return false
	}

	receiver := s.findClient(args[0])

	if receiver == nil {
		sendError(client, "User not found: "+args[0])
		return false
	}

	if receiver == client {
		sendError(client, "You cannot PM yourself")
		return false
	}

	message := strings.Join(args[1:], " ")

	if err := receiver.Send(protocol.Message{
		Type:     "pm",
		Username: client.User().Username(),
		Target:   args[0],
		Message:  message,
	}); err != nil {
		log.Println(err)
	}

	if err := client.Send(protocol.Message{
		Type:     "pm",
		Username: client.User().Username(),
		Target:   args[0],
		Message:  message,
	}); err != nil {
		log.Println(err)
	}
	return false
}

func (s *Server) commandJoinRoom(client *Client, roomName string, password string) bool {
	if err := s.JoinRoom(client, roomName, password); err != nil {
		sendError(client, err.Error())
		return false
	}

	return false
}

func (s *Server) commandRooms(client *Client) bool {
	rooms := s.RoomsSnapshot()

	sendSystem(client, fmt.Sprintf("Rooms: %d", len(rooms)))

	for _, room := range rooms {
		sendSystem(client, fmt.Sprintf("- %s", room.Name()+"\n   Users: "+strconv.Itoa(room.Size())))
	}

	return false
}

func (s *Server) commandRoom(client *Client) bool {

	room := client.Room()

	sendSystem(client, fmt.Sprintf("Room: %s", room.Name()))

	sendSystem(client, fmt.Sprintf("Owner: %s", room.Owner()))

	sendSystem(client, fmt.Sprintf("Admins: %s", strings.Join(room.AdminsUsernames(), ", ")))

	sendSystem(client, fmt.Sprintf("Topic: %s", room.Topic()))

	sendSystem(client, fmt.Sprintf("Users: %d", room.Size()))

	return false
}

func (s *Server) commandRenameRoom(name string, client *Client) bool {

	room := client.Room()

	if err := room.RequireOwner(client); err != nil {
		sendError(client, err.Error())

		return false
	}

	if err := s.RenameRoom(room, name); err != nil {
		sendError(client, err.Error())
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

func (s *Server) commandSetTopic(client *Client, topic string) bool {
	room := client.Room()

	if err := room.RequireAdmin(client); err != nil {
		sendError(client, err.Error())
		return false
	}

	room.SetTopic(topic)
	if err := s.SaveConfig(); err != nil {
		log.Println("Failed to save config: " + err.Error())
	}

	room.Broadcast(protocol.Message{
		Type:    "system",
		Message: client.User().Username() + " changed the topic to: " + topic,
	}, nil)

	return false
}

func (s *Server) commandAddAdmin(client *Client, targetUsername string) bool {

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

	if err := room.RequireOwner(client); err != nil {
		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: err.Error(),
		}); err != nil {
			log.Println(err)
		}
		return false
	}

	if err := room.AddAdmin(target); err != nil {
		if err := client.Send(protocol.Message{
			Type:    "error",
			Message: err.Error(),
		}); err != nil {
			log.Println(err)
		}
		return false
	}

	room.Broadcast(protocol.Message{
		Type:    "system",
		Message: client.User().Username() + " made " + target.User().Username() + " admin",
	}, nil)

	return false
}

func (s *Server) commandRemoveAdmin(client *Client, targetUsername string) bool {

	target := s.findClient(targetUsername)
	if target == nil {
		sendError(client, "User not found: "+targetUsername)
		return false
	}

	room := client.Room()

	if err := room.RequireOwner(client); err != nil {
		sendError(client, err.Error())
		return false
	}

	if err := room.RemoveAdmin(target); err != nil {
		sendError(client, err.Error())
		return false
	}

	room.Broadcast(protocol.Message{
		Type:    "system",
		Message: target.User().Username() + "s admin status was revoked by " + client.User().Username(),
	}, nil)

	return false
}

func (s *Server) commandSetPassword(client *Client, password string) bool {
	room := client.Room()

	if err := room.RequireOwner(client); err != nil {
		sendError(client, err.Error())
		return false
	}

	if err := room.SetPassword(password); err != nil {
		sendError(client, err.Error())
		return false
	} else {
		if err := s.SaveConfig(); err != nil {
			log.Println("Failed to save config: " + err.Error())
		}
		room.Broadcast(protocol.Message{
			Type:    "system",
			Message: client.User().Username() + " set a password",
		}, nil)
	}

	return false
}

func (s *Server) commandKickUser(client *Client, targetUsername string) bool {

	target := s.findClient(targetUsername)

	if target == nil {
		sendError(client, "User not found: "+targetUsername)
		return false
	}

	if target == client {
		sendError(client, "You cannot kick yourself")
		return false
	}

	room := client.Room()

	if err := room.RequireAdmin(client); err != nil {
		sendError(client, err.Error())
		return false
	}

	if err := room.KickUser(target); err != nil {
		sendError(client, err.Error())
		return false
	}

	if err := s.JoinRoom(target, "general", ""); err != nil {
		log.Println(err)
	}

	if err := s.SaveConfig(); err != nil {
		log.Println("Failed to save config: " + err.Error())
	}

	sendSystem(client, "You have been kicked from "+room.Name()+
		" by "+client.User().Username()+". You have been moved to #general")

	room.Broadcast(protocol.Message{
		Type:    "system",
		Message: client.User().Username() + " kicked " + target.User().Username(),
	}, nil)

	return false
}

func sendError(client *Client, msg string) {
	if err := client.Send(protocol.Message{
		Type:    "error",
		Message: msg,
	}); err != nil {
		log.Println(err)
	}
}

func sendSystem(client *Client, msg string) {
	if err := client.Send(protocol.Message{
		Type:    "system",
		Message: msg,
	}); err != nil {
		log.Println(err)
	}
}

func requireArgs(client *Client, args []string, count int, usage string) bool {
	if len(args) < count {
		sendError(client, "Usage: "+usage)
		return false
	}

	return true
}

func (s *Server) Commands() map[string]Command {
	return s.commands
}

func (s *Server) InitializeCommands() {
	s.commands = map[string]Command{
		"/users": {
			Description: "Lists all users",
			Handler: func(s *Server, c *Client, args []string) bool {
				return s.commandUsers(c)
			},
			Usage: "/users",
		},

		"/rooms": {
			Description: "Lists all rooms",
			Handler: func(s *Server, c *Client, args []string) bool {
				return s.commandRooms(c)
			},
			Usage: "/rooms",
		},

		"/room": {
			Description: "Current room details",
			Handler: func(s *Server, c *Client, args []string) bool {
				return s.commandRoom(c)
			},
			Usage: "/room",
		},

		"/topic": {
			Description: "Current room topic",
			Handler: func(s *Server, c *Client, args []string) bool {
				return s.commandTopic(c)
			},
			Usage: "/topic",
		},

		"/help": {
			Description: "Lists all commands",
			Handler: func(s *Server, c *Client, args []string) bool {
				return s.commandHelp(c)
			},
			Usage: "/help",
		},

		"/quit": {
			Description: "Disconnects from the server",
			Handler: func(s *Server, c *Client, args []string) bool {
				return s.commandQuit(c)
			},
			Usage: "/quit",
		},

		"/pm": {
			Description: "Send a private message to a user",
			Handler: func(s *Server, c *Client, args []string) bool {
				return s.commandPM(c, args)
			},
			Usage: "/pm <username> <message>",
		},

		"/join": {
			Description: "Joins a room",
			Handler: func(s *Server, c *Client, args []string) bool {

				if !requireArgs(c, args, 1, "/join <room> [password]") {
					return false
				}

				password := ""

				if len(args) > 1 {
					password = strings.Join(args[1:], " ")
				}

				return s.commandJoinRoom(c, args[0], password)
			},
			Usage: "/join <room> [password]",
		},

		"/kick": {
			Description: "Kicks a user from the room (admin only)",
			Handler: func(s *Server, c *Client, args []string) bool {

				if !requireArgs(c, args, 1, "/kick <username>") {
					return false
				}

				return s.commandKickUser(c, args[0])
			},
			Usage: "/kick <username>",
		},

		"/settopic": {
			Description: "Sets the room topic (admin only)",
			Handler: func(s *Server, c *Client, args []string) bool {

				if !requireArgs(c, args, 1, "/settopic <topic>") {
					return false
				}

				return s.commandSetTopic(c, strings.Join(args, " "))
			},
			Usage: "/settopic <topic>",
		},

		"/rename": {
			Description: "Renames the room",
			Handler: func(s *Server, c *Client, args []string) bool {

				if !requireArgs(c, args, 1, "/rename <name>") {
					return false
				}

				return s.commandRenameRoom(args[0], c)
			},
			Usage: "/rename <name>",
		},
	}
}