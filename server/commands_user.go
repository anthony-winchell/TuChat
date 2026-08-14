package main

import (
	"tuchat/protocol"
	"time"
	"strings"
	"log"
	"sort"
	"fmt"
	"strconv"
)

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

	room.broadcastUserList()

	return false
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

func (s *Server) commandUsers(client *Client) bool {
	room := client.Room()
	clients := room.Users()

	summaries := make([]protocol.UserSummary, 0, len(clients))
	for _, c := range clients {
		username := c.User().Username()
		summaries = append(summaries, protocol.UserSummary{
			Nickname: c.User().Nickname(),
			Admin:    room.IsAdmin(username),
			Owner:    room.IsOwner(username),
		})
	}

	if err := client.Send(protocol.Message{
		Type:  "users",
		Users: summaries,
	}); err != nil {
		log.Println(err)
	}

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

func (s *Server) commandQuit(client *Client) bool {
	sendSystem(client, "Goodbye!")
	return true
}