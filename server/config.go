package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	ServerName 	string `json:"server_name"`
	Rooms      	[]RoomConfig `json:"rooms"`
	
}


type RoomConfig struct {
	Name     		 	string `json:"name"`
	Topic    		 	string `json:"topic"`
	PasswordHash 	string `json:"password_hash"`
}

func (s *Server) SaveConfig() error {
	config := Config{
		ServerName: s.Name(),
	}

	for _, room := range s.RoomsSnapshot() {
		config.Rooms = append(config.Rooms, RoomConfig{
			Name:      		room.Name(),
			Topic:     		room.Topic(),
			PasswordHash: room.PasswordHash(),
		})
	}

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	if err = os.WriteFile("config.json", data, 0644); err != nil {
		return err
	}
	return nil
}

func (s *Server) loadConfig() error {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return err
	}

	var config Config 
	if err = json.Unmarshal(data, &config); err != nil {
		return err 
	}

	s.SetName(config.ServerName)
	
	for _, roomConfig := range config.Rooms {
		room := NewRoom(roomConfig.Name)
		room.SetTopic(roomConfig.Topic)
		room.RestorePasswordHash(roomConfig.PasswordHash)

		s.AddRoom(room)
	}

	return nil 
}


