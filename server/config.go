package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ServerName string       `json:"server_name"`
	Owner      string       `json:"owner"`
	Rooms      []RoomConfig `json:"rooms"`
	Users      []UserConfig `json:"users"`
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

func (s *Server) SaveConfig() error {
	config := Config{
		ServerName: s.Name(),
		Owner:      s.Owner(),
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

func (s *Server) Owner() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.owner
}

func (s *Server) SetOwner(owner string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.owner = owner
}

func (s *Server) RequireServerOwner(client *Client) error {
	if client.User().Username() == s.Owner() {
		return nil
	}

	return errors.New("server owner permissions required")
}

func (s *Server) configureOwner() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Set server owner username:")
		fmt.Print("> ")

		owner, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		owner = strings.TrimSpace(owner)

		if err := ValidateUsername(owner); err != nil {
			fmt.Println(err.Error())
			continue
		}

		s.SetOwner(owner)
		return 
	}
}
