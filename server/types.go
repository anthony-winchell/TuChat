package main

import (
	"encoding/json"
	"net"
	"sync"
)

type Client struct {
	conn     net.Conn
	username string

	decoder *json.Decoder
	encoder *json.Encoder
	writeMu sync.Mutex
	mu      sync.RWMutex

	room *Room
}

type Room struct {
	name string 
	clients map[string]*Client
}

type Server struct {
	mu       sync.RWMutex
	listener net.Listener

	name string

	clients map[string]*Client
	conns   map[net.Conn]struct{}

	rooms map[string]*Room

	wg sync.WaitGroup
}