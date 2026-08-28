package main

import (
	"bufio"
	"encoding/json"
	"net"
	"sync"
	"time"
	"tuchat/protocol"
)

type Client struct {
	conn net.Conn

	user *User

	encoder *json.Encoder
	input   *bufio.Scanner
	mu      sync.RWMutex

	outbox   chan protocol.Message
	done     chan struct{}
	stopOnce sync.Once

	room *Room
}

type User struct {
	mu           sync.RWMutex
	username     string
	nickname     string
	passwordHash string
}

type Room struct {
	mu sync.RWMutex

	name     string
	password string

	owner  string
	admins map[string]struct{}

	topic   string
	clients map[string]*Client
}

type Server struct {
	mu       sync.RWMutex
	saveMu   sync.Mutex
	listener net.Listener

	name      string
	advertise string

	welcomeMessage string

	owner string

	password string

	clients map[string]*Client
	conns   map[net.Conn]struct{}

	users map[string]*User

	rooms map[string]*Room

	chatLogs map[string]*ChatLog

	commands map[string]Command

	startTime time.Time

	wg sync.WaitGroup
}

type Config struct {
	ServerName         string       `json:"server_name"`
	Owner              string       `json:"owner"`
	ServerPasswordHash string       `json:"server_password_hash"`
	WelcomeMessage     string       `json:"welcome_message"`
	Rooms              []RoomConfig `json:"rooms"`
	Users              []UserConfig `json:"users"`
}

type UserConfig struct {
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	PasswordHash string `json:"password_hash"`
}

type RoomConfig struct {
	Name         string   `json:"name"`
	Topic        string   `json:"topic"`
	PasswordHash string   `json:"password_hash"`
	Owner        string   `json:"owner"`
	Admins       []string `json:"admins"`
}
