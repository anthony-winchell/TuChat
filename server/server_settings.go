package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"tuchat/protocol"

	"golang.org/x/crypto/bcrypt"
)

func (s *Server) AdvertiseAddr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.advertise
}

func (s *Server) SetAdvertiseAddr(addr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.advertise = addr
}

func (s *Server) Name() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.name
}

func (s *Server) SetName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.name = name
}

func (s *Server) configureName() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Name Your Server (leave blank for 'TuChat'):")

	name, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
		return
	}

	name = strings.TrimSpace(name)

	if name != "" {
		s.name = name
	}
}

func (s *Server) Owner() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.owner
}

func (s *Server) SetOwner(owner string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.owner = strings.ToLower(owner)
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

func (s *Server) SetPassword(password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	password = strings.TrimSpace(password)

	if password == "" {
		s.password = ""
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	s.password = string(hash)
	return nil
}

func (s *Server) HasPassword() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.password != ""
}

func (s *Server) CheckPassword(password string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.password == "" {
		return true
	}

	return bcrypt.CompareHashAndPassword([]byte(s.password), []byte(password)) == nil
}

func (s *Server) PasswordHash() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.password
}

func (s *Server) RestorePasswordHash(hash string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.password = hash
}

func (s *Server) WelcomeMessage() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.welcomeMessage
}

func (s *Server) SetWelcomeMessage(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.welcomeMessage = message
}

func (s *Server) sendWelcome(client *Client) {
	template := s.WelcomeMessage()

	if template == "" {
		template = defaultWelcomeMessage
	}

	if err := client.Send(
		protocol.Message{
			Type:       "welcome",
			Message:    template,
			Nickname:   client.User().Nickname(),
			ServerName: s.Name(),
		},
	); err != nil {
		log.Println(err)
	}
}
