package main

import (
	"testing"
)

func TestNewUser(t *testing.T) {
	user, err := NewUser("anthony", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Username() != "anthony" {
		t.Fatal("expected username to be anthony")
	}

	if user.PasswordHash() == "password" {
		t.Fatal("password should not be stored as plaintext")
	}

	if user.PasswordHash() == "" {
		t.Fatal("expected password hash to not be set")
	}

	if user.Nickname() != "anthony" {
		t.Fatal("expected nickname to be anthony")
	}
}

func TestUserUsername(t *testing.T) {
	user, err := NewUser("anthony", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Username() != "anthony" {
		t.Fatalf("expected username %q, got %q", "anthony", user.Username())
	}

	user.SetUsername("bob")

	if user.Username() != "bob" {
		t.Fatalf("expected username %q, got %q", "bob", user.Username())
	}

}

func TestUserNickname(t *testing.T) {
	user, err := NewUser("anthony", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Nickname() != "anthony" {
		t.Fatalf("expected nickname %q, got %q", "anthony", user.Nickname())
	}

	user.SetNickname("Anthony")

	if user.Nickname() != "Anthony" {
		t.Fatalf("expected nickname %q, got %q", "Anthony", user.Nickname())
	}
}

func TestUserPassword(t *testing.T) {
	user, err := NewUser("anthony", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := user.CheckPassword("password"); err != nil {
		t.Fatal("expected password to be valid")
	}

	if err := user.CheckPassword("wrong"); err == nil {
		t.Fatal("expected password to be invalid")
	}
}

func TestSetUserPassword(t *testing.T) {
	user, err := NewUser("anthony", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	oldHash := user.PasswordHash()

	if err := user.SetPassword("newpassword"); err != nil {
		t.Fatal("expected password to be valid")
	}

	newHash := user.PasswordHash()

	if newHash == oldHash {
		t.Fatal("expected password hash to be updated")
	}

	if newHash == "newpassword" {
		t.Fatal("expected password hash to not be stored as plaintext")
	}

	if newHash == "" {
		t.Fatal("expected password hash to not be empty")
	}

	if err := user.CheckPassword("newpassword"); err != nil {
		t.Fatalf("expected password to be valid: %v", err)
	}

	if err := user.CheckPassword("password"); err == nil {
		t.Fatal("expected password to be invalid")
	}

}

func TestUserPasswordHash(t *testing.T) {
	user, err := NewUser("anthony", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hash := user.PasswordHash()

	if hash == "" {
		t.Fatal("expected password hash to be non-empty")
	}

	if hash == "password" {
		t.Fatal("expected password hash to not be stored as plaintext")
	}
}

func TestRestoreUser(t *testing.T) {
	original, err := NewUser("anthony", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hash := original.PasswordHash()

	restored := RestoreUser("anthony", "Anthony", hash)

	if restored.Username() != "anthony" {
		t.Fatalf("expected username %q, got %q", "anthony", restored.Username())
	}

	if restored.Nickname() != "Anthony" {
		t.Fatalf("expected nickname %q, got %q", "Anthony", restored.Nickname())
	}

	if restored.PasswordHash() != hash {
		t.Fatalf("expected password hash %q, got %q", hash, restored.PasswordHash())
	}

	if err := restored.CheckPassword("password"); err != nil {
		t.Fatal("expected restored password hash to match original")
	}
	if err := restored.CheckPassword("wrong"); err == nil {
		t.Fatal("expected wrong password to be rejected")
	}

}

func TestUserSetUsernameDoesNotChangeNickname(t *testing.T) {
	user, err := NewUser("anthony", "password")
	if err != nil {
		t.Fatal(err)
	}

	user.SetNickname("Anthony")

	user.SetUsername("bob")

	if user.Username() != "bob" {
		t.Fatalf("expected username %q, got %q", "bob", user.Username())
	}

	if user.Nickname() != "Anthony" {
		t.Fatalf(
			"expected nickname to remain %q, got %q",
			"Anthony",
			user.Nickname(),
		)
	}
}
