package main

import (
	"errors"
	"net"
	"testing"
)

func newTestServer() *Server {
	return &Server{
		clients:  make(map[string]*Client),
		users:    make(map[string]*User),
		rooms:    make(map[string]*Room),
		conns:    make(map[net.Conn]struct{}),
		chatLogs: make(map[string]*ChatLog),
		commands: make(map[string]Command),

		authRateLimit:    NewRateLimitMap(5, 5.0/30.0),
		messageRateLimit: NewRateLimitMap(10, 5.0),
	}
}

func TestServerAddUser(t *testing.T) {
	server := newTestServer()

	user, err := NewUser("anthony", "password")
	if err != nil {
		t.Fatal(err)
	}

	if err := server.AddUser(user); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, err := server.FindUser("anthony")
	if err != nil {
		t.Fatalf("unexpected error finding user: %v", err)
	}

	if found != user {
		t.Fatal("expected server to return the same user")
	}
}

func TestServerAddDuplicateUser(t *testing.T) {
	server := newTestServer()

	first, err := NewUser("anthony", "password")
	if err != nil {
		t.Fatal(err)
	}

	second, err := NewUser("anthony", "different")
	if err != nil {
		t.Fatal(err)
	}

	if err := server.AddUser(first); err != nil {
		t.Fatal(err)
	}

	err = server.AddUser(second)

	if !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf(
			"expected ErrUsernameTaken, got %v",
			err,
		)
	}
}

func TestServerFindUser(t *testing.T) {
	server := newTestServer()

	user, err := NewUser("anthony", "password")
	if err != nil {
		t.Fatal(err)
	}

	if err := server.AddUser(user); err != nil {
		t.Fatal(err)
	}

	found, err := server.FindUser("anthony")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if found.Username() != "anthony" {
		t.Fatalf(
			"expected username %q, got %q",
			"anthony",
			found.Username(),
		)
	}
}

func TestServerFindUserNotFound(t *testing.T) {
	server := newTestServer()

	_, err := server.FindUser("anthony")

	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf(
			"expected ErrUserNotFound, got %v",
			err,
		)
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid username",
			input:   "anthony",
			wantErr: false,
		},
		{
			name:    "minimum length",
			input:   "abc",
			wantErr: false,
		},
		{
			name:    "maximum length",
			input:   "abcdefghijklm",
			wantErr: false,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "too short",
			input:   "ab",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   "abcdefghijklmn",
			wantErr: true,
		},
		{
			name:    "contains colon",
			input:   "anthony:test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"ValidateUsername(%q) error = %v, wantErr = %v",
					tt.input,
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestServerRegisterUser(t *testing.T) {
	server := newTestServer()

	user, err := server.RegisterUser("anthony", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user == nil {
		t.Fatal("expected registered user")
	}

	if user.Username() != "anthony" {
		t.Fatalf(
			"expected username %q, got %q",
			"anthony",
			user.Username(),
		)
	}

	found, err := server.FindUser("anthony")
	if err != nil {
		t.Fatalf("expected registered user to be stored: %v", err)
	}

	if found != user {
		t.Fatal("expected stored user to be the registered user")
	}
}

func TestServerRegisterDuplicateUser(t *testing.T) {
	server := newTestServer()

	_, err := server.RegisterUser("anthony", "password")
	if err != nil {
		t.Fatalf("unexpected error registering first user: %v", err)
	}

	_, err = server.RegisterUser("anthony", "password")

	if !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf(
			"expected ErrUsernameTaken, got %v",
			err,
		)
	}
}

func TestServerRegisterInvalidUser(t *testing.T) {
	server := newTestServer()

	tests := []string{
		"",
		"ab",
		"abcdefghijklmn",
		"anthony:test",
	}

	for _, username := range tests {
		t.Run(username, func(t *testing.T) {
			_, err := server.RegisterUser(username, "password")

			if err == nil {
				t.Fatalf(
					"expected registration of %q to fail",
					username,
				)
			}
		})
	}
}

func TestServerAuthenticateUser(t *testing.T) {
	server := newTestServer()

	_, err := server.RegisterUser("anthony", "password")
	if err != nil {
		t.Fatalf("unexpected error registering user: %v", err)
	}

	user, err := server.AuthenticateUser("anthony", "password")
	if err != nil {
		t.Fatalf("unexpected authentication error: %v", err)
	}

	if user.Username() != "anthony" {
		t.Fatalf(
			"expected username %q, got %q",
			"anthony",
			user.Username(),
		)
	}
}

func TestServerAuthenticateUserErrorsAreIndistinguishable(t *testing.T) {
	server := newTestServer()

	if _, err := server.RegisterUser("anthony", "password"); err != nil {
		t.Fatal(err)
	}

	_, unknownUser := server.AuthenticateUser("ghost", "password")
	_, badPassword := server.AuthenticateUser("anthony", "wrong")

	if unknownUser == nil || badPassword == nil {
		t.Fatal("expected both authentication attempts to fail")
	}

	if unknownUser.Error() != badPassword.Error() {
		t.Fatalf(
			"auth failures must not leak account existence: unknown user = %q, bad password = %q",
			unknownUser.Error(),
			badPassword.Error(),
		)
	}
}
func TestServerNicknameTaken(t *testing.T) {
	server := newTestServer()

	anthony := newTestClient(t, "anthony")
	anthony.User().SetNickname("Anthony")

	server.clients["anthony"] = anthony

	err := server.nicknameTaken("Anthony", nil)

	if err == nil {
		t.Fatal("expected nickname to be taken")
	}
}

func TestServerNicknameTakenCaseInsensitive(t *testing.T) {
	server := newTestServer()

	anthony := newTestClient(t, "anthony")
	anthony.User().SetNickname("Anthony")

	server.clients["anthony"] = anthony

	err := server.nicknameTaken("ANTHONY", nil)

	if err == nil {
		t.Fatal("expected nickname to be taken")
	}
}

func TestServerNicknameAvailable(t *testing.T) {
	server := newTestServer()

	anthony := newTestClient(t, "anthony")
	anthony.User().SetNickname("Anthony")

	server.clients["anthony"] = anthony

	err := server.nicknameTaken("Bob", nil)

	if err != nil {
		t.Fatalf(
			"expected nickname to be available, got %v",
			err,
		)
	}
}

func TestServerNicknameTakenExcludesClient(t *testing.T) {
	server := newTestServer()

	anthony := newTestClient(t, "anthony")
	anthony.User().SetNickname("Anthony")

	server.clients["anthony"] = anthony

	err := server.nicknameTaken("Anthony", anthony)

	if err != nil {
		t.Fatalf(
			"expected client's own nickname to be allowed, got %v",
			err,
		)
	}
}

func TestServerNicknameTakenWithMultipleClients(t *testing.T) {
	server := newTestServer()

	anthony := newTestClient(t, "anthony")
	anthony.User().SetNickname("Anthony")

	bob := newTestClient(t, "bob")
	bob.User().SetNickname("Bob")

	server.clients["anthony"] = anthony
	server.clients["bob"] = bob

	if err := server.nicknameTaken("Anthony", nil); err == nil {
		t.Fatal("expected Anthony to be taken")
	}

	if err := server.nicknameTaken("Bob", nil); err == nil {
		t.Fatal("expected Bob to be taken")
	}

	if err := server.nicknameTaken("Charlie", nil); err != nil {
		t.Fatalf(
			"expected Charlie to be available, got %v",
			err,
		)
	}
}
