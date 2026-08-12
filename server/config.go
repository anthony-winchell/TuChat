package main

import (
	"encoding/json"
	"os"
)

func (s *Server) SaveConfig() error {
	config := Config{
		ServerName: s.Name(),
		Owner:      s.Owner(),
		ServerPasswordHash: s.PasswordHash(),
		WelcomeMessage:     s.WelcomeMessage(),
	}

	for _, user := range s.usersSnapshot() {
		config.Users = append(config.Users, UserConfig{
			Username:     user.Username(),
			Nickname:     user.Nickname(),
			PasswordHash: user.PasswordHash(),
		})
	}

	for _, room := range s.RoomsSnapshot() {
		config.Rooms = append(config.Rooms, RoomConfig{
			Name:         room.Name(),
			Topic:        room.Topic(),
			PasswordHash: room.PasswordHash(),
			Owner:        room.Owner(),
			Admins:       room.AdminsUsernames(),
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
	s.SetOwner(config.Owner)
	s.RestorePasswordHash(config.ServerPasswordHash)
	s.SetWelcomeMessage(config.WelcomeMessage)

	for _, roomConfig := range config.Rooms {
		room := NewRoom(roomConfig.Name)
		room.SetTopic(roomConfig.Topic)
		room.RestorePasswordHash(roomConfig.PasswordHash)
		room.SetOwner(roomConfig.Owner)

		for _, username := range roomConfig.Admins {
			room.RestoreAdmin(username)
		}

		if err := s.AddRoom(room); err != nil {
			return err
		}
	}

	for _, userConfig := range config.Users {
		user := RestoreUser(userConfig.Username, userConfig.Nickname, userConfig.PasswordHash)
		if err := s.AddUser(user); err != nil {
			return err
		}
	}

	return nil
}
