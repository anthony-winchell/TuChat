package main

import (
	"golang.org/x/crypto/bcrypt"
)

func (u *User) Username() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.username
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

func NewUser(username string, password string) (*User, error) {
	u := &User{
		username: username,
	}

	if err := u.SetPassword(password); err != nil {
		return nil, err
	}

	return u, nil
}

func RestoreUser(username string, passwordHash string) *User {
	return &User{
		username: username,
		passwordHash: passwordHash,
	}
}