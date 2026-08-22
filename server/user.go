package main

import (
	"errors"
	"log"
	"strings"
	"tuchat/protocol"

	"golang.org/x/crypto/bcrypt"
)

func (u *User) Username() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.username
}

func (u *User) Nickname() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.nickname
}

func (u *User) SetNickname(nickname string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.nickname = nickname
}

func (u *User) SetUsername(username string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.username = username
}

func (u *User) SetPassword(password string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.passwordHash = string(hash)
	return nil
}

func (u *User) PasswordHash() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.passwordHash
}

func (u *User) CheckPassword(password string) error {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return bcrypt.CompareHashAndPassword([]byte(u.passwordHash), []byte(password))
}


func (s *Server) AddUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.Username()]; exists {
		return ErrUsernameTaken
	}

	s.users[user.Username()] = user

	return nil
}

func (s *Server) FindUser(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.users[username]; !exists {
		return nil, ErrUserNotFound
	}
	return s.users[username], nil

}

func (s *Server) RegisterUser(username string, password string) (*User, error) {

	if err := ValidateUsername(username); err != nil {
		return nil, err
	}

	user, err := NewUser(username, password)
	if err != nil {
		return nil, err
	}

	if err := s.AddUser(user); err != nil {
		return nil, err
	}

	if err := s.SaveConfig(); err != nil {
		log.Println("Failed to save config: " + err.Error())
	}

	return user, nil
}

func (s *Server) AuthenticateUser(username string, password string) (*User, error) {
	user, err := s.FindUser(username)
	if err != nil {
		return nil, errors.New(protocol.AuthInvalidCredentials)
	}

	if err := user.CheckPassword(password); err != nil {
		return nil, errors.New(protocol.AuthInvalidCredentials)
	}

	return user, nil
}

func ValidateUsername(username string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}

	if len(username) < 3 || len(username) > 13 {
		return ErrUsernameFormat
	}

	if strings.Contains(username, ":") {
		return ErrUsernameFormat
	}

	return nil
}

func ValidateNickname(nickname string) error {
	if nickname == "" {
		return errors.New("nickname cannot be empty")
	}

	if len(nickname) < 3 || len(nickname) > 13 {
		return errors.New("nickname must be between 3 and 13 characters long")
	}

	if strings.Contains(nickname, ":") {
		return errors.New("nickname cannot contain ':'")
	}

	if strings.ContainsAny(nickname, "\r\n\t") {
		return errors.New("nickname cannot contain whitespace")
	}

	return nil
}

func (s *Server) nicknameTaken(nickname string, except *Client) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		if client == except {
			continue
		}

		if strings.EqualFold(client.User().Nickname(), nickname) {
			return errors.New("nickname taken")
		}
	}

	return nil
}

func NewUser(username string, password string) (*User, error) {
	u := &User{
		username: username,
		nickname: username,
	}

	if err := u.SetPassword(password); err != nil {
		return nil, err
	}

	return u, nil
}

func RestoreUser(username string, nickname string, passwordHash string) *User {
	return &User{
		username:     username,
		nickname:     nickname,
		passwordHash: passwordHash,
	}
}
