package main

import (
	"log"
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

		"/leave": {
			Description: "Leaves the current room",
			Handler: func(s *Server, c *Client, args []string) bool {
				return s.commandLeaveRoom(c)
			},
			Usage: "/leave",
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
			Usage: "/setwelcome <message> (omit to reset to default. \n" +
				"{server} and {nickname} will replace with server name \n" +
				" and user's nickname)",
		},
		"/addadmin": {
			Description: "Adds an admin to the room (owner only)",
			Handler: func(s *Server, c *Client, args []string) bool {
				if !requireArgs(c, args, 1, "/addadmin <username>") {
					return false
				}
				return s.commandAddAdmin(c, args[0])
			},
			Usage: "/addadmin <username>",
		},
		"/removeadmin": {
			Description: "Removes an admin from the room (owner only)",
			Handler: func(s *Server, c *Client, args []string) bool {
				if !requireArgs(c, args, 1, "/removeadmin <username>") {
					return false
				}
				return s.commandRemoveAdmin(c, args[0])
			},
			Usage: "/removeadmin <username>",
		},
	}
}

func sendError(client *Client, msg string) {
	if err := client.Send(protocol.Message{
		Type:      "error",
		Message:   msg,
		Timestamp: time.Now(),
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
