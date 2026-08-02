package main

import( 
	"net"
)

func (s *Server) roomSnapshot(room *Room) []*Client {
	room.mu.RLock()
	defer room.mu.RUnlock()

	clients := make([]*Client, 0, len(room.clients))

	for _, client := range room.clients {
		clients = append(clients, client)
	}	

	return clients
}

func (s *Server) clientsSnapshot() []*Client {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]*Client, 0, len(s.clients))

	for _, client := range s.clients {
		clients = append(clients, client)
	}

	return clients
}

func (s *Server) connSnapshot() []net.Conn {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conns := make([]net.Conn, 0, len(s.conns))

	for conn := range s.conns {
		conns = append(conns, conn)
	}
	return conns
}

