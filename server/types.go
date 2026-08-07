package main

import (
	"encoding/json"
	"net"
	"sync"
)

type Client struct {
	conn     	net.Conn

	user      *User

	decoder 	*json.Decoder
	encoder 	*json.Encoder
	writeMu 	sync.Mutex
	mu      	sync.RWMutex

	room 			*Room
}

type User struct {
	mu 						sync.RWMutex
	username 			string 
	passwordHash 	string
}

type Room struct {
	mu sync.RWMutex

	name    	string
	password 	string

	owner   	string
	admins    map[string]struct{}

	topic   	string
	clients 	map[string]*Client
}

type Server struct {
	mu       	sync.RWMutex
	listener 	net.Listener

	name 			string

	clients 	map[string]*Client
	conns   	map[net.Conn]struct{}

	users 		map[string]*User

	rooms 		map[string]*Room

	commands 	map[string]Command

	wg 				sync.WaitGroup
}
