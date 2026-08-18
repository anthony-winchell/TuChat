package main

import (
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var ErrUsernameTaken = errors.New("Username already taken")
var ErrUsernameFormat = errors.New(`Username cannot contain ':'. Must be between 3 and 13 characters long`)
var ErrUserNotFound = errors.New("User not found")


func main() {
	server := &Server{
		name:      "TuChat",
		clients:   make(map[string]*Client),
		users:     make(map[string]*User),
		conns:     make(map[net.Conn]struct{}),
		chatLogs:  make(map[string]*ChatLog),
		rooms:     make(map[string]*Room),
		startTime: time.Now(),
	}

	server.InitializeCommands()

	if err := server.loadConfig(); err != nil {
		log.Println("No config found. Creating defaults...")
		createDefaultState(server)
		server.configureName()
		server.configureOwner()

		if err := server.SaveConfig(); err != nil {
			log.Println(err)
		}
	} else {
		log.Println("Config found. Server name: " + server.Name())

		if server.Owner() == "" {
			server.configureOwner()
			if err := server.SaveConfig(); err != nil {
				log.Println(err)
			}
		}
	}

	address := net.JoinHostPort(server.BindAddress(), strconv.Itoa(server.Port()))

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}

	server.listener = listener 

	log.Println("Starting Server... ")
	log.Println("Listening on: ", listener.Addr())
	if server.AdvertisedAddress() != "" {
		log.Println("Clients can connect using: ", server.AdvertisedAddress())
	}
	go server.Start()

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, os.Interrupt)

	<-signals

	server.SaveConfig()

	server.Shutdown()

}

func createDefaultState(server *Server) {
	server.SetName("TuChat")

	server.SetBindAddress("0.0.0.0")
	server.SetPort(8080)

	server.AddRoom(NewRoom("general"))
}
