package main

import (
	"errors"
	"log"
	"strings"
	"time"
	"tuchat/protocol"

	"golang.org/x/crypto/bcrypt"
)

var ErrAlreadyInRoom = errors.New("already in room")

func (s *Server) JoinRoom(client *Client, roomName string, password string) error {
	room, newRoom, err := s.FindOrCreateRoom(roomName)
	if err != nil {
		return err
	}

	if newRoom {
		s.broadcastRoomList()
	}

	if client.Room() == room {
		return ErrAlreadyInRoom
	}

	if !room.CheckPassword(password) {
		return ErrIncorrectPassword
	}

	oldRoom := client.Room()

	if oldRoom != nil {
		oldRoom.Remove(client)

		oldRoom.Broadcast(protocol.Message{
			Type:    "system",
			Message: client.User().Nickname() + " left the room",
		}, client)

		oldRoom.broadcastUserList()
		s.broadcastRoomList()
	}

	room.Add(client)

	if newRoom {
		room.SetOwner(client.User().Username())

		if err := s.SaveConfig(); err != nil {
			log.Println("Failed to save config:", err)
		}

		if err := client.Send(protocol.Message{
			Type:    "system",
			Message: "Room created: " + roomName,
		}); err != nil {
			log.Println(err)
		}
	}

	if err := client.Send(protocol.Message{
		Type:    "system",
		Message: "Joined room: " + roomName,
		Timestamp: time.Now(),
	}); err != nil {
		log.Println(err)
	}

	room.Broadcast(protocol.Message{
		Type:    "system",
		Message: client.User().Nickname() + " joined the room",
	}, client)

	room.broadcastUserList()
	room.broadcastRoomInfo()

	s.broadcastRoomList()


	return nil
}


func (r *Room) broadcastUserList() {
	clients := r.Users()

	summaries := make([]protocol.UserSummary, 0, len(clients))

	for _, client := range clients {
		username := client.User().Username()
		summaries = append(summaries, protocol.UserSummary{
			Nickname: client.User().Nickname(),
			Admin: r.IsAdmin(username),
			Owner: r.IsOwner(username),

		})
	}
	r.Broadcast(protocol.Message{
		Type: "users",
		Users: summaries,
	}, nil)
}

func (s *Server) FindOrCreateRoom(name string) (*Room, bool, error) {
	name = strings.TrimSpace(name)

	if err := ValidateRoomName(name); err != nil {
		return nil, false, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if room, exists := s.rooms[name]; exists {
		return room, false, nil
	}

	room := NewRoom(name)
	s.rooms[name] = room

	return room, true, nil
}

func (s *Server) broadcastRoomList() {
	rooms := s.RoomsSnapshot()

	summaries := make([]protocol.RoomSummary, 0, len(rooms))

	for _, room := range rooms {
		summaries = append(summaries, protocol.RoomSummary{
			Name: room.Name(),
			Users: room.Size(),
			HasPassword: room.HasPassword(),
		})
	}

	s.sendToAll(protocol.Message{
		Type: "rooms",
		Rooms: summaries,
	}, nil)
}

func (r *Room) Remove(client *Client) {

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.clients, client.User().Username())


	client.mu.Lock()
	client.room = nil
	client.mu.Unlock()
}

func (r *Room) Add(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clients[client.User().Username()] = client

	client.mu.Lock()
	client.room = r
	client.mu.Unlock()
}

func (r *Room) Users() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]*Client, 0, len(r.clients))

	for _, client := range r.clients {
		clients = append(clients, client)
	}

	return clients
}

func (r *Room) FindByNickname(nickname string) *Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, client := range r.clients {
		if client.User().Nickname() == nickname {
			return client
		}
	}

	return nil

}

func (r *Room) Broadcast(msg protocol.Message, sender *Client) {
	clients := r.Users()

	for _, client := range clients {
		if client == sender {
			continue
		}

		if err := client.Send(msg); err != nil {
			log.Println(err)
		}
	}
}

func (r *Room) Name() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.name
}

func (r *Room) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.clients)
}

func (r *Room) IsAdmin(username string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.admins[username]

	if ok {
		return true
	}

	return false
}

func (r *Room) IsOwner(username string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.owner == username
}

func (r *Room) Has(client *Client) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.clients[client.User().Username()]
	return ok
}

func (r *Room) Topic() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.topic
}

func (r *Room) SetTopic(topic string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.topic = topic
}

func (r *Room) Admins() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]*Client, 0, len(r.admins))

	for username := range r.admins {
		if client, ok := r.clients[username]; ok {
			clients = append(clients, client)
		}
	}

	return clients
}

func (r *Room) AdminsUsernames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	usernames := make([]string, 0, len(r.admins))

	for username := range r.admins {
		usernames = append(usernames, username)
	}

	return usernames
}

func (r *Room) SetOwner(username string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.owner = username
	r.admins[username] = struct{}{}
}

func (r *Room) RequireAdmin(client *Client) error {

	if r.Name() == "general" {
		return errors.New("general is the default room and cannot be modified")
	}

	if r.IsOwner(client.User().Username()) {
		return nil
	}

	if !r.IsAdmin(client.User().Username()) {
		return errors.New("admin permissions required")
	}


	return nil
}

func (r *Room) RequireOwner(client *Client) error {
	if r.Name() == "general" {
		return errors.New("general is the default room and cannot be modified")
	}

	if client.User().Username() != r.Owner() {
		return errors.New("owner permissions required")
	}

	return nil
}

func (r *Room) KickUser(client *Client) error {
	username := client.User().Username()

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.clients[username]; !ok {
		return errors.New("user not in this room")
	}

	delete(r.clients, username)
	delete(r.admins, username)

	client.mu.Lock()
	client.room = nil
	client.mu.Unlock()

	return nil
}

func (r *Room) AddAdmin(client *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.clients[client.User().Username()]; !ok {
		return errors.New("user not in this room")
	}

	if _, ok := r.admins[client.User().Username()]; ok {
		return errors.New("user is already an admin")
	}

	r.admins[client.User().Username()] = struct{}{}

	return nil
}

func (r *Room) RemoveAdmin(client *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.clients[client.User().Username()]; !ok {
		return errors.New("user not in room")
	}

	if client.User().Username() == r.owner {
		return errors.New("cannot demote owner")
	}

	if _, ok := r.admins[client.User().Username()]; !ok {
		return errors.New("user is not an admin")
	}

	delete(r.admins, client.User().Username())

	return nil
}

func (r *Room) RestoreAdmin(username string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.admins[username] = struct{}{}
}

func (r *Room) SetPassword(password string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if password == "" {
		r.password = ""
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	r.password = string(hash)

	return nil
}

func (r *Room) HasPassword() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.password != ""
}

func (r *Room) CheckPassword(password string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.password == "" {
		return true
	}

	return bcrypt.CompareHashAndPassword([]byte(r.password), []byte(password)) == nil
}

func (r *Room) PasswordHash() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.password
}

func NewRoom(name string) *Room {
	return &Room{
		name:    name,
		clients: make(map[string]*Client),
		admins:  make(map[string]struct{}),
	}
}

func (r *Room) RestorePasswordHash(hash string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.password = hash
}

func (r *Room) Owner() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.owner
}

func (s *Server) RoomsSnapshot() []*Room {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rooms := make([]*Room, 0, len(s.rooms))

	for _, room := range s.rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

func (s *Server) RenameRoom(r *Room, newName string) error {
	oldName := r.Name()
	newName = strings.TrimSpace(newName)

	if err := ValidateRoomName(newName); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rooms[newName]; exists {
		return errors.New("room exists")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	chatLog, exists := s.chatLogs[oldName]

	if exists {
		if err := chatLog.Rename("logs/" + newName + ".log"); err != nil {
			return err
		}

		delete(s.chatLogs, oldName)

		s.chatLogs[newName] = chatLog
	}

	delete(s.rooms, r.name)

	r.name = newName

	s.rooms[newName] = r

	return nil
}

func (s *Server) DeleteRoom(r *Room) error {
	name := r.Name()

	if name == "general" {
		return errors.New("the #general room cannot be deleted")
	}

	s.mu.Lock()

	if _, exists := s.rooms[name]; !exists {
		s.mu.Unlock()
		return errors.New("room not found")
	}

	general := s.rooms["general"]

	delete(s.rooms, name)

	chatLog, hasChatLog := s.chatLogs[name]
	if hasChatLog {
		delete(s.chatLogs, name)
	}

	s.mu.Unlock()

	if hasChatLog {
		if err := chatLog.Delete(); err != nil {
			return err
		}
	}

	for _, client := range s.roomSnapshot(r) {
		if err := client.Send(protocol.Message{
			Type:    "system",
			Message: "#" + name + " has been deleted. Moving to #general",
		}); err != nil {
			log.Println(err)
		}

		r.Remove(client)
		general.Add(client)
	}

	general.broadcastUserList()
	s.broadcastRoomList()
	general.broadcastRoomInfo()

	return nil
}

func (s *Server) AddRoom(room *Room) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rooms[room.Name()]; exists {
		return errors.New("room exists")
	}

	s.rooms[room.Name()] = room

	return nil
}


func ValidateRoomName(name string) error {
	if name == "" {
		return errors.New("room name cannot be empty")
	}

	if len(name) < 2 || len(name) > 50 {
		return errors.New("room name must be between 2 and 50 characters long")
	}

	if strings.Contains(name, " ") {
		return errors.New("room name cannot contain spaces")
	}

	if strings.Contains(name, "/") {
		return errors.New("room name cannot contain forward slashes")
	}

	if strings.Contains(name, "\\") {
		return errors.New("room name cannot contain back slashes")
	}

	if strings.Contains(name, ".") {
		return errors.New("room name cannot contain periods")
	}

	if strings.ContainsAny(name, "#!@$%^&*()[]{}|;:,<>?~`") {
		return errors.New("room name cannot contain special characters")
	}

	return nil
}
