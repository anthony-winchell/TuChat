package main

import (
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	addr      = flag.String("addr", ":8080", "listen address")
	advertise = flag.String("advertise", "", "address clients should dial to reach this server")
	name      = flag.String("name", "", "server name (default: TuChat)")
	owner     = flag.String("owner", "", "server owner username")
)

var (
	ErrUsernameTaken  = errors.New("Username already taken")
	ErrUsernameFormat = errors.New(`Username cannot contain ':'. Must be between 3 and 13 characters long`)
	ErrUserNotFound   = errors.New("User not found")
)

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Println(err)
		return
	}

	server := &Server{
		name:      "TuChat",
		listener:  listener,
		clients:   make(map[string]*Client),
		users:     make(map[string]*User),
		conns:     make(map[net.Conn]struct{}),
		chatLogs:  make(map[string]*ChatLog),
		rooms:     make(map[string]*Room),
		startTime: time.Now(),
	}

	advertised := *advertise
	if advertised == "" {
		advertised = *addr
	}
	server.SetAdvertiseAddr(advertised)

	server.InitializeCommands()

	if err := server.loadConfig(); err != nil {
		log.Println("No config found. Creating defaults...")
		createDefaultState(server)

		if *name != "" {
			server.SetName(*name)
		} else {
			server.configureName()
		}

		if *owner != "" {
			server.SetOwner(*owner)
		} else {
			server.configureOwner()
		}

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

	log.Println("Starting Server... ")
	go server.Start()

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	<-signals

	server.SaveConfig()

	server.Shutdown()
}

func createDefaultState(server *Server) {
	server.SetName("TuChat")
	server.AddRoom(NewRoom("general"))
}
