package main

import (
	"log"
	"tuchat/protocol"
)

func (s *Server) JoinRoom(client *Client, roomName string) error {
	room := s.FindOrCreateRoom(roomName)

	if client.Room() != nil {
		client.Room().Remove(client)
	}

	room.Add(client)

	return nil
}

func (s *Server) FindOrCreateRoom(name string) *Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.rooms[name]
	if !ok {
		s.rooms[name] = &Room{
			name:      name,
			clients:   make(map[string]*Client),
			operators: make(map[string]struct{}),
		}
	}

	return s.rooms[name]
}

func (r *Room) Remove(client *Client) {

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.clients, client.Username())

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
