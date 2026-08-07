package main

import (
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
)

var ErrUsernameTaken = errors.New("Username already taken")
var ErrUsernameFormat = errors.New(`Username cannot contain ':'. Must be between 3 and 13 characters long`)
var ErrUserNotFound = errors.New("User not found")

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Println(err)
		return
	}

	server := &Server{
		name:     "TuChat",
		listener: listener,
		clients:  make(map[string]*Client),
		users:    make(map[string]*User),
		conns:    make(map[net.Conn]struct{}),
		rooms:    make(map[string]*Room),
	}

	server.InitializeCommands()
	
	if err := server.loadConfig(); err != nil {
		log.Println("No config found. Creating defaults...")
		createDefaultState(server)
		server.configureName()

		if err := server.SaveConfig(); err != nil {
			log.Println(err)
		}
	} else {
		log.Println("Config found. Server name: " + server.Name())

	}

	log.Println("Starting Server... ")
	go server.Start()

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, os.Interrupt)

	<-signals

	server.SaveConfig()

	server.Shutdown()

}


func createDefaultState(server *Server) {
	server.SetName("TuChat")
	server.AddRoom(NewRoom("general"))
}