package main

import (
	"errors"
	"log"
	"time"
	"tuchat/protocol"
)

var ErrIncorrectPassword = errors.New("incorrect password")

func (s *Server) commandRoom(client *Client) bool {

	room := client.Room()

	if err := client.Send(protocol.Message{
		Type:          "roominfo",
		RoomName:      room.Name(),
		RoomOwner:     room.Owner(),
		RoomAdmins:    room.AdminsUsernames(),
		RoomTopic:     room.Topic(),
		RoomUserCount: room.Size(),
	}); err != nil {
		log.Println(err)
	}

	return false
}

func (s *Server) commandRooms(client *Client) bool {
	rooms := s.RoomsSnapshot()

	summaries := make([]protocol.RoomSummary, 0, len(rooms))
	for _, room := range rooms {
		summaries = append(summaries, protocol.RoomSummary{
			Name:        room.Name(),
			Users:       room.Size(),
			HasPassword: room.HasPassword(),
		})
	}

	if err := client.Send(protocol.Message{
		Type:  "rooms",
		Rooms: summaries,
	}); err != nil {
		log.Println(err)
	}

	return false
}

func (s *Server) commandJoinRoom(client *Client, roomName string, password string) bool {
	if err := s.JoinRoom(client, roomName, password); err != nil {
		if errors.Is(err, ErrIncorrectPassword) {
			if err := client.Send(protocol.Message{
				Type:      "join_password_required",
				Message:   roomName,
				Timestamp: time.Now(),
			}); err != nil {
				log.Println(err)
			}
			return false
		}
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
		Type:      "system",
		Message:   client.User().Nickname() + " renamed the room to " + name,
		Timestamp: time.Now(),
	}, nil)

	room.broadcastRoomInfo()

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
		Type:          "roominfo",
		RoomName:      room.Name(),
		RoomTopic:     room.Topic(),
		RoomOwner:     room.Owner(),
		RoomAdmins:    room.AdminsUsernames(),
		RoomUserCount: room.Size(),
		Timestamp:     time.Now(),
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
		Type:      "system",
		Message:   client.User().Nickname() + " changed the topic to: " + topic,
		Timestamp: time.Now(),
	}, nil)

	room.broadcastRoomInfo()

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
		Type:      "system",
		Message:   client.User().Nickname() + " made " + target.User().Nickname() + " admin",
		Timestamp: time.Now(),
	}, nil)

	room.broadcastUserList()

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
		Type:      "system",
		Message:   target.User().Nickname() + "s admin status was revoked by " + client.User().Nickname(),
		Timestamp: time.Now(),
	}, nil)

	room.broadcastUserList()

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
			Type:      "system",
			Message:   client.User().Nickname() + " set a password",
			Timestamp: time.Now(),
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
		Type:      "system",
		Message:   client.User().Nickname() + " kicked " + target.User().Nickname(),
		Timestamp: time.Now(),
	}, nil)

	room.broadcastUserList()

	return false
}

func (r *Room) broadcastRoomInfo() {
	r.Broadcast(protocol.Message{
		Type:          "roominfo",
		RoomName:      r.Name(),
		RoomOwner:     r.Owner(),
		RoomAdmins:    r.AdminsUsernames(),
		RoomTopic:     r.Topic(),
		RoomUserCount: r.Size(),
		Timestamp:     time.Now(),
	}, nil)
}
