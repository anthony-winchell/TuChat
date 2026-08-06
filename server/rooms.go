package main

import (
	"errors"
	"log"
	"strings"
	"tuchat/protocol"

	"golang.org/x/crypto/bcrypt"
)

var ErrAlreadyInRoom = errors.New("already in room")

func (s *Server) JoinRoom(client *Client, roomName string, password string) error {
	room, new := s.FindOrCreateRoom(roomName)

	if client.Room() == room {
		return ErrAlreadyInRoom
	}

	if !room.CheckPassword(password) {
		return errors.New("incorrect password")
	}

	oldRoom := client.Room()

	if oldRoom != nil {
		oldRoom.Remove(client)

		oldRoom.Broadcast(protocol.Message{
			Type:    "system",
			Message: client.Username() + " left the room",
		}, client)
	}

	if new {
		if err := client.Send(protocol.Message{
			Type:    "system",
			Message: "Room created: " + roomName,
		}); err != nil {
			log.Println(err)
		}
	}

	room.Add(client)

	if new {
		room.AddOperator(client)
		if err := s.SaveConfig(); err != nil {
			log.Println("Failed to save config: " + err.Error())
		}
	}

	if err := client.Send(protocol.Message{
		Type:    "system",
		Message: "Joined room: " + roomName,
	}); err != nil {
		log.Println(err)
	}

	room.Broadcast(protocol.Message{
		Type:    "system",
		Message: client.Username() + " joined the room",
	}, client)

	return nil
}

func (s *Server) FindOrCreateRoom(name string) (*Room, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name = strings.TrimSpace(name)

	new := false

	_, ok := s.rooms[name]
	if !ok {
		s.rooms[name] = &Room{
			name:      name,
			clients:   make(map[string]*Client),
			operators: make(map[string]struct{}),
		}

		new = true
	}

	return s.rooms[name], new
}

func (r *Room) Remove(client *Client) {

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.clients, client.Username())
	delete(r.operators, client.Username())

	client.mu.Lock()
	client.room = nil
	client.mu.Unlock()

}

func (r *Room) Add(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clients[client.Username()] = client

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

func (r *Room) IsOperator(username string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.operators[username]
	return ok
}

func (r *Room) Has(client *Client) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.clients[client.Username()]
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

func (r *Room) Operators() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]*Client, 0, len(r.operators))

	for username := range r.operators {
		if client, ok := r.clients[username]; ok {
			clients = append(clients, client)
		}
	}

	return clients
}

func (r *Room) AddOperator(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.operators[client.Username()] = struct{}{}
}

func (r *Room) RemoveOperator(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.operators, client.Username())
}

func (r *Room) RequireOperator(client *Client) error {

	if r.Name() == "general" {
		return errors.New("general is the default room and cannot be modified")
	}

	if !r.IsOperator(client.Username()) {
		return errors.New("operator permissions required")
	}

	return nil
}

func (r *Room) Promote(client *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.clients[client.Username()]; !ok {
		return errors.New("user not in this room")
	}

	if _, ok := r.operators[client.Username()]; ok {
		return errors.New("user is already an operator")
	}

	r.operators[client.Username()] = struct{}{}

	return nil
}

func (r *Room) Demote(client *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.clients[client.Username()]; !ok {
		return errors.New("user not in room")
	}

	if _, ok := r.operators[client.Username()]; !ok {
		return errors.New("user is not an operator")
	}

	delete(r.operators, client.Username())

	return nil
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
		name: 			name,
		clients: 		make(map[string]*Client),
		operators: 	make(map[string]struct{}),
	}
}

func (r *Room) RestorePasswordHash(hash string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.password = hash
}