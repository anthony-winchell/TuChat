package main

import (
	"fmt"
	"log"
	"strings"
	"time"
	"tuchat/protocol"
)

func (s *Server) commandServerInfo(client *Client) bool {
	sendSystem(client, "Server: "+s.Name())
	sendSystem(client, "Uptime: "+time.Since(s.startTime).Round(time.Second).String())
	sendSystem(client, fmt.Sprintf("Rooms: %d", len(s.RoomsSnapshot())))
	sendSystem(client, fmt.Sprintf("Users: %d", len(s.clientsSnapshot())))
	sendSystem(client, "Server Address: "+s.AdvertiseAddr())

	return false
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
		Type:    "system",
		Message: client.User().Nickname() + " renamed the server to " + name,
	}, nil)

	s.sendToAll(protocol.Message{
		Type:    "server_name",
		Message: s.Name(),
	}, nil)

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
