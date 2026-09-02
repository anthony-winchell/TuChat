package tui

import (
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"
)

func TestReconnectDelay(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{5, 16 * time.Second},
		{6, 30 * time.Second},
		{7, 30 * time.Second},
		{10, 30 * time.Second},
	}

	for _, tt := range tests {
		if got := reconnectDelay(tt.attempt); got != tt.want {
			t.Fatalf("reconnectDelay(%d) = %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

// TestSubmitAuthInputPreservesServerPassword drives the full first-connect
// submit sequence (server password, then username+login password) and checks
// that the login submission does not discard the remembered server password.
func TestSubmitAuthInputPreservesServerPassword(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	t.Cleanup(func() {
		serverConn.Close()
		clientConn.Close()
	})

	go func() {
		io.Copy(io.Discard, serverConn)
	}()

	m := New("localhost:8080")
	m.connection.conn = clientConn
	m.connection.enc = json.NewEncoder(clientConn)
	m.connection.dec = json.NewDecoder(clientConn)

	m.auth.choice = "login"
	m.auth.stage = stageServerPassword
	m.auth.input.SetValue("secret")
	model, _ := m.submitAuthInput()
	m = model.(Model)

	m.auth.stage = stageUsername
	m.auth.input.SetValue("alice")
	model, _ = m.submitAuthInput()
	m = model.(Model)

	m.auth.stage = stagePassword
	m.auth.input.SetValue("password")
	model, _ = m.submitAuthInput()
	m = model.(Model)

	if m.connection.creds.serverPassword != "secret" {
		t.Fatalf("serverPassword = %q, want %q", m.connection.creds.serverPassword, "secret")
	}
	if m.connection.creds.choice != "login" {
		t.Fatalf("choice = %q, want login", m.connection.creds.choice)
	}
	if m.connection.creds.username != "alice" {
		t.Fatalf("username = %q, want alice", m.connection.creds.username)
	}
	if m.connection.creds.password != "password" {
		t.Fatalf("password = %q, want password", m.connection.creds.password)
	}
}
