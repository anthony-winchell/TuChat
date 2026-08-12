package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
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

func (s *Server) commandServerRename(client *Client, name string) bool {
	if err := s.RequireServerOwner(client); err != nil {
		sendError(client, err.Error())
		return false
	}

	name = strings.TrimSpace(name)

	if name == "" {
		sendError(client, "Server name cannot be empty")
		return false
	}

	s.SetName(name)

	if err := s.SaveConfig(); err != nil {
		log.Println("Failed to save config: " + err.Error())
	}

	s.sendToAll(protocol.Message{
		Type: "system",
		Message: client.User().Nickname() + " renamed the server to " + name,
	}, nil)

	return false
}

func (s *Server) commandSetServerPassword(client *Client, password string) bool {
	if err := s.RequireServerOwner(client); err != nil {
		sendError(client, err.Error())
		return false
	}

	if err := s.SetPassword(password); err != nil {
		sendError(client, err.Error())
		return false
	}

	if err := s.SaveConfig(); err != nil {
		log.Println("Failed to save config: " + err.Error())
	}

	sendSystem(client, "Server password updated")

	return false
}

func (s *Server) commandSetWelcomeMessage(client *Client, message string) bool {
	if err := s.RequireServerOwner(client); err != nil {
		sendError(client, err.Error())
		return false
	}

	message = strings.ReplaceAll(message, `\n`, "\n")

	s.SetWelcomeMessage(message)

	if err := s.SaveConfig(); err != nil {
		log.Println("Failed to save config: " + err.Error())
	}

	sendSystem(client, "Welcome message updated")

	return false
}

func (s *Server) commandAnnounce(client *Client, message string) bool {
	if err := s.RequireServerOwner(client); err != nil {
		sendError(client, err.Error())
		return false
	}

	s.sendToAll(protocol.Message{
		Type:    "announcement",
		Message: message,
	}, nil)

	return false
}

func (s *Server) commandNick(client *Client, args []string) bool {
	if len(args) < 1 {
		sendError(client, "Usage: /nick <new nickname>")
		return false
	}
	oldNickname := client.User().Nickname()

	nickname := args[0]
	nickname = strings.TrimSpace(nickname)

	if err := ValidateNickname(nickname); err != nil {
		sendError(client, err.Error())
		return false
	}

	if err := s.nicknameTaken(nickname, client); err != nil {
		sendError(client, err.Error())
		return false
	}

	client.User().SetNickname(nickname)

	if err := s.SaveConfig(); err != nil {
		log.Println("Failed to save config: " + err.Error())
	}

	room := client.Room()

	room.Broadcast(protocol.Message{
		Type:     "system",
		Username: nickname,
		Message:  oldNickname + " changed their nickname to " + nickname,
	}, nil)

	return false
}

func (s *Server) commandHelp(client *Client) bool {

	message := "Commands:\n"

	names := make([]string, 0, len(s.commands))
	for name := range s.commands {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		command := s.commands[name]

		message += fmt.Sprintf(
			"%s - %s\n    Usage: %s\n",
			name,
			command.Description,
			command.Usage,
		)
	}

	sendSystem(client, message)

	return false
}

func (s *Server) commandServerInfo(client *Client) bool {
	sendSystem(client, "Server: "+s.Name())
	sendSystem(client, "Uptime: "+time.Since(s.startTime).Round(time.Second).String())
	sendSystem(client, fmt.Sprintf("Rooms: %d", len(s.RoomsSnapshot())))
	sendSystem(client, fmt.Sprintf("Users: %d", len(s.clientsSnapshot())))

	return false
}

func (s *Server) commandUsers(client *Client) bool {
	users := client.Room().Users()
	usernames := make([]string, 0, len(users))

	for _, user := range users {
		usernames = append(usernames, user.User().Nickname())
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
		sendError(client, "Usage: /pm <nickname> <message>")
		return false
	}

	time := time.Now()

	targetNickname := args[0]

	receiver := s.findClientByNickname(targetNickname)

	if receiver == nil {
		sendError(client, "User not found: "+targetNickname)
		return false
	}

	if receiver == client {
		sendError(client, "You cannot PM yourself")
		return false
	}

	message := strings.Join(args[1:], " ")

	msg := protocol.Message{
		Type:      "pm",
		Username:  client.User().Username(),
		Nickname:  client.User().Nickname(),
		Target:    receiver.User().Nickname(),
		Message:   message,
		Timestamp: time,
	}

	if err := receiver.Send(msg); err != nil {
		log.Println("Failed to send PM:", err)
		return false
	}

	if err := client.Send(msg); err != nil {
		log.Println("Failed to send PM confirmation:", err)
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

	room.Broadcast(protocol.Message{
		Type:    "system",
		Message: client.User().Nickname() + " renamed the room to " + name,
	}, nil)
	return false
}

func (s *Server) commandDeleteRoom(client *Client) bool {

	room := client.Room()

	if err := room.RequireOwner(client); err != nil {
		sendError(client, err.Error())
		return false
	}

	if err := s.DeleteRoom(room); err != nil {
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
		Message: client.User().Nickname() + " changed the topic to: " + topic,
	}, nil)

	return false
}

func (s *Server) commandAddAdmin(client *Client, targetUsername string) bool {

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

	if err := room.AddAdmin(target); err != nil {
		sendError(client, err.Error())
		return false
	}

	room.Broadcast(protocol.Message{
		Type:    "system",
		Message: client.User().Nickname() + " made " + target.User().Nickname() + " admin",
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
		Message: target.User().Nickname() + "s admin status was revoked by " + client.User().Nickname(),
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
			Message: client.User().Nickname() + " set a password",
		}, nil)
	}

	return false
}

func (s *Server) commandKickUser(client *Client, targetNickname string) bool {

	target := s.findClientByNickname(targetNickname)

	if target == nil {
		sendError(client, "User not found: "+targetNickname)
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

	sendSystem(target, "You have been kicked from "+room.Name()+
		" by "+client.User().Nickname()+". You have been moved to #general")

	room.Broadcast(protocol.Message{
		Type:    "system",
		Message: client.User().Nickname() + " kicked " + target.User().Nickname(),
	}, nil)

	return false
}

func (s *Server) commandHistory(client *Client, args []string) bool {
	count := 20

	if len(args) > 0 {
		parsed, err := strconv.Atoi(args[0])
		if err != nil || parsed <= 0 {
			sendError(client, "Usage: /history [positive number] (default: 20)")
			return false
		}
		if parsed > 300 {
			sendError(client, "History limit is 300 messages")
			return false
		}
		count = parsed
	}
	room := client.Room()

	chatLog, err := s.getChatLog(room.Name())
	if err != nil {
		sendError(client, "Unable to access chat history")
		log.Println(err)
		return false
	}

	entries, err := chatLog.Recent(count)
	if err != nil {
		sendError(client, "Unable to access chat history")
		log.Println(err)
		return false
	}

	for _, entry := range entries {
		sendSystem(client, fmt.Sprintf("%s: %s", entry.Nickname, entry.Message))
	}

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
			Usage: "/pm <nickname> <message>",
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

		"/delete": {
			Description: "Deletes the room (owner only)",
			Handler: func(s *Server, c *Client, args []string) bool {
				return s.commandDeleteRoom(c)
			},
			Usage: "/delete",
		},
		"/history": {
			Description: "Lists the last N messages in the room",
			Handler: func(s *Server, c *Client, args []string) bool {
				return s.commandHistory(c, args)
			},
			Usage: "/history [positive number] (default: 20)",
		},

		"/nick": {
			Description: "Change your nickname",
			Handler: func(s *Server, c *Client, args []string) bool {
				if !requireArgs(c, args, 1, "/nick <nickname>") {
					return false
				}

				if len(args) > 1 {
					sendError(c, "Nickname cannot contain spaces")
					return false
				}

				return s.commandNick(c, args)
			},
			Usage: "/nick <nickname>",
		},
		"/serverinfo": {
			Description: "Lists basic server information",
			Handler: func(s *Server, c *Client, args []string) bool {
				return s.commandServerInfo(c)
			},
			Usage: "/serverinfo",
		},
		"/serverrename": {
			Description: "Renames the server (owner only)",
			Handler: func(s *Server, c *Client, args []string) bool {
				if !requireArgs(c, args, 1, "/serverrename <name>") {
					return false
				}
				return s.commandServerRename(c, strings.Join(args, " "))
			},
			Usage: "/serverrename <name>",
		},

		"/announce": {
			Description: "Sends a server-wide announcement (owner only)",
			Handler: func(s *Server, c *Client, args []string) bool {
				if !requireArgs(c, args, 1, "/announce <message>") {
					return false
				}
				return s.commandAnnounce(c, strings.Join(args, " "))
			},
			Usage: "/announce <message>",
		},
		"/setpassword": {
			Description: "Sets or clears the room password (owner only)",
			Handler: func(s *Server, c *Client, args []string) bool {
				password := ""
				if len(args) > 0 {
					password = strings.Join(args, " ")
				}
				return s.commandSetPassword(c, password)
			},
			Usage: "/setpassword [password] (omit to clear)",
		},
		"/setserverpassword": {
			Description: "Sets or clears the server password (owner only)",
			Handler: func(s *Server, c *Client, args []string) bool {
				password := ""
				if len(args) > 0 {
					password = strings.Join(args, " ")
				}
				return s.commandSetServerPassword(c, password)
			},
			Usage: "/setserverpassword [password] (omit to clear)",
		},
		"/setwelcome": {
			Description: "Sets the welcome message for new users (owner only)",
			Handler: func(s *Server, c *Client, args []string) bool {
				message := ""
				if len(args) > 0 {
					message = strings.Join(args, " ")
				}
				return s.commandSetWelcomeMessage(c, message)
			},
			Usage: "/setwelcome <message> (omit to reset to default. {server} and {nickname} will replace with server name and user's nickname)",
		},
	}
}
