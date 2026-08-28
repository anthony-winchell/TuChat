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
	addr      = flag.String("addr", "", "listen address (default: :8080)")
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

	listenerAddr := *addr
	if listenerAddr == "" {
		listenerAddr = os.Getenv("TUCHAT_ADDR")
	}
	if listenerAddr == "" {
		listenerAddr = ":8080"
	}

	serverName := *name
	if serverName == "" {
		serverName = os.Getenv("TUCHAT_NAME")
	}

	serverOwner := *owner
	if serverOwner == "" {
		serverOwner = os.Getenv("TUCHAT_OWNER")
	}

	listener, err := net.Listen("tcp", listenerAddr)
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
		advertised = os.Getenv("TUCHAT_ADVERTISE")
	}
	if advertised == "" {
		advertised = listenerAddr
	}

	server.SetAdvertiseAddr(advertised)

	server.InitializeCommands()

	if err := server.loadConfig(); err != nil {
		log.Println("No config found. Creating defaults...")
		createDefaultState(server)

		if serverName != "" {
			server.SetName(serverName)
		} else {
			server.configureName()
		}

		if serverOwner != "" {
			server.SetOwner(serverOwner)
		} else {
			server.configureOwner()
		}

		if err := server.SaveConfig(); err != nil {
			log.Println(err)
		}
	} else {
		log.Println("Config found. Server name: " + server.Name())

		if server.Owner() == "" {
			if serverOwner != "" {
				server.SetOwner(serverOwner)
			} else {
				server.configureOwner()
			}
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
