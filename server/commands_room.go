package main 

import (
	"fmt"
	"strconv"
	"tuchat/protocol"
	"strings"
	"log"
)

func (s *Server) commandRoom(client *Client) bool {

	room := client.Room()

	sendSystem(client, fmt.Sprintf("Room: %s", room.Name()))

	sendSystem(client, fmt.Sprintf("Owner: %s", room.Owner()))

	sendSystem(client, fmt.Sprintf("Admins: %s", strings.Join(room.AdminsUsernames(), ", ")))

	sendSystem(client, fmt.Sprintf("Topic: %s", room.Topic()))

	sendSystem(client, fmt.Sprintf("Users: %d", room.Size()))

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

func (s *Server) commandJoinRoom(client *Client, roomName string, password string) bool {
	if err := s.JoinRoom(client, roomName, password); err != nil {
		sendError(client, err.Error())
		return false
	}

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