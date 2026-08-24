package main

import (
	"errors"
	"testing"
)

func TestServerAddClient(t *testing.T) {
	server := newTestServer()

	client := newTestClient(t, "anthony")

	if err := server.addClient(client); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := server.findClient("anthony")

	if found != client {
		t.Fatal("expected server to contain added client")
	}
}

func TestServerAddClientDuplicate(t *testing.T) {
	server := newTestServer()

	first := newTestClient(t, "anthony")
	second := newTestClient(t, "anthony")

	if err := server.addClient(first); err != nil {
		t.Fatalf("unexpected error adding first client: %v", err)
	}

	err := server.addClient(second)

	if !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf(
			"expected ErrUsernameTaken, got %v",
			err,
		)
	}
}

func TestServerAddClientUsernameTooShort(t *testing.T) {
	server := newTestServer()

	client := newTestClient(t, "ab")

	err := server.addClient(client)

	if !errors.Is(err, ErrUsernameFormat) {
		t.Fatalf(
			"expected ErrUsernameFormat, got %v",
			err,
		)
	}
}

func TestServerAddClientUsernameTooLong(t *testing.T) {
	server := newTestServer()

	client := newTestClient(t, "abcdefghijklmn")

	err := server.addClient(client)

	if !errors.Is(err, ErrUsernameFormat) {
		t.Fatalf(
			"expected ErrUsernameFormat, got %v",
			err,
		)
	}
}

func TestServerAddClientUsernameWithColon(t *testing.T) {
	server := newTestServer()

	client := newTestClient(t, "anthony:test")

	err := server.addClient(client)

	if !errors.Is(err, ErrUsernameFormat) {
		t.Fatalf(
			"expected ErrUsernameFormat, got %v",
			err,
		)
	}
}

func TestServerFindClient(t *testing.T) {
	server := newTestServer()

	client := newTestClient(t, "anthony")

	if err := server.addClient(client); err != nil {
		t.Fatal(err)
	}

	found := server.findClient("anthony")

	if found != client {
		t.Fatal("expected findClient to return the correct client")
	}
}

func TestServerFindClientNotFound(t *testing.T) {
	server := newTestServer()

	found := server.findClient("anthony")

	if found != nil {
		t.Fatal("expected findClient to return nil")
	}
}

func TestServerRemoveClient(t *testing.T) {
	server := newTestServer()

	client := newTestClient(t, "anthony")

	if err := server.addClient(client); err != nil {
		t.Fatal(err)
	}

	server.removeClient(client)

	if found := server.findClient("anthony"); found != nil {
		t.Fatal("expected client to be removed")
	}
}

func TestServerRemoveClientDoesNotAffectOtherClients(t *testing.T) {
	server := newTestServer()

	anthony := newTestClient(t, "anthony")
	bob := newTestClient(t, "bob")

	if err := server.addClient(anthony); err != nil {
		t.Fatal(err)
	}

	if err := server.addClient(bob); err != nil {
		t.Fatal(err)
	}

	server.removeClient(anthony)

	if server.findClient("anthony") != nil {
		t.Fatal("expected anthony to be removed")
	}

	if server.findClient("bob") != bob {
		t.Fatal("expected bob to remain connected")
	}
}

func TestServerFindClientByNickname(t *testing.T) {
	server := newTestServer()

	client := newTestClient(t, "anthony")
	client.User().SetNickname("Anthony")

	if err := server.addClient(client); err != nil {
		t.Fatal(err)
	}

	found := server.findClientByNickname("Anthony")

	if found != client {
		t.Fatal("expected client to be found by nickname")
	}
}

func TestServerFindClientByNicknameNotFound(t *testing.T) {
	server := newTestServer()

	client := newTestClient(t, "anthony")

	if err := server.addClient(client); err != nil {
		t.Fatal(err)
	}

	found := server.findClientByNickname("Bob")

	if found != nil {
		t.Fatal("expected findClientByNickname to return nil")
	}
}

func TestServerFindClientByNicknameCaseInsensitive(t *testing.T) {
	server := newTestServer()

	client := newTestClient(t, "anthony")
	client.User().SetNickname("Anthony")

	if err := server.addClient(client); err != nil {
		t.Fatalf("unexpected error adding client: %v", err)
	}

	found := server.findClientByNickname("anthony")

	if found != client {
		t.Fatal("expected nickname lookup to be case-insensitive")
	}
}

func TestServerFindClientByNicknameMultipleClients(t *testing.T) {
	server := newTestServer()

	anthony := newTestClient(t, "anthony")
	anthony.User().SetNickname("Anthony")

	bob := newTestClient(t, "bob")
	bob.User().SetNickname("Bob")

	if err := server.addClient(anthony); err != nil {
		t.Fatal(err)
	}

	if err := server.addClient(bob); err != nil {
		t.Fatal(err)
	}

	found := server.findClientByNickname("Bob")

	if found != bob {
		t.Fatal("expected Bob to be returned")
	}

	found = server.findClientByNickname("Anthony")

	if found != anthony {
		t.Fatal("expected Anthony to be returned")
	}
}
