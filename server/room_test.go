package main

import (
	"strings"
	"testing"
)

func TestValidateRoomName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid room",
			input:   "general",
			wantErr: false,
		},
		{
			name:    "valid room with numbers",
			input:   "room123",
			wantErr: false,
		},
		{
			name:    "empty room",
			input:   "",
			wantErr: true,
		},
		{
			name:    "too short",
			input:   "a",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   strings.Repeat("a", 70),
			wantErr: true,
		},
		{
			name:    "contains space",
			input:   "my room",
			wantErr: true,
		},
		{
			name:    "contains slash",
			input:   "my/room",
			wantErr: true,
		},
		{
			name:    "contains backslash",
			input:   "my\\room",
			wantErr: true,
		},
		{
			name:    "contains period",
			input:   "my.room",
			wantErr: true,
		},
		{
			name:    "contains special character",
			input:   "my!room",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRoomName(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"ValidateRoomName(%q) error = %v, wantErr = %v",
					tt.input,
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestRoomAddAndRemove(t *testing.T) {
	room := NewRoom("test")

	client := newTestClient(t, "anthony")

	room.Add(client)

	if !room.Has(client) {
		t.Fatal("expected client to be in room")
	}

	if client.Room() != room {
		t.Fatal("expected client's room to be updated")
	}

	room.Remove(client)

	if room.Has(client) {
		t.Fatal("expected client to be removed from room")
	}

	if client.Room() != nil {
		t.Fatal("expected client's room to be nil")
	}
}

func TestRoomKickUser(t *testing.T) {
	room := NewRoom("test")

	client := newTestClient(t, "anthony")

	room.Add(client)

	if err := room.KickUser(client); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if room.Has(client) {
		t.Fatal("expected client to be kicked from room")
	}

	if client.Room() != nil {
		t.Fatal("expected kicked client's room to be nil")
	}
}

func TestRoomKickUserNotInRoom(t *testing.T) {
	room := NewRoom("test")

	client := newTestClient(t, "anthony")

	err := room.KickUser(client)

	if err == nil {
		t.Fatal("expected error, but got nil")
	}
}

func TestRoomPasswordHash(t *testing.T) {
	room := NewRoom("test")

	if room.PasswordHash() != "" {
		t.Fatal("expected password hash to initially be empty")
	}

	if err := room.SetPassword("password"); err != nil {
		t.Fatal(err)
	}

	hash := room.PasswordHash()

	if hash == "" {
		t.Fatal("expected password hash to be set")
	}

	if hash == "password" {
		t.Fatal("password should not be stored as plaintext")
	}
}

func TestRoomPasswordCanBeRemoved(t *testing.T) {
	room := NewRoom("test")

	if err := room.SetPassword("password"); err != nil {
		t.Fatal(err)
	}

	if err := room.SetPassword(""); err != nil {
		t.Fatal(err)
	}

	if room.HasPassword() {
		t.Fatal("expected room to have no password")
	}

	if !room.CheckPassword("") {
		t.Fatal("expected empty password after removal")
	}
}

func TestRoomAdminUsernames(t *testing.T) {
	room := NewRoom("test")

	client := newTestClient(t, "anthony")
	user := client.User()

	room.Add(client)

	if room.IsAdmin(user.Username()) {
		t.Fatal("user should not be initialized as admin")
	}

	if err := room.AddAdmin(client); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !room.IsAdmin(user.Username()) {
		t.Fatal("expexted user to be admin")
	}

	if err := room.RemoveAdmin(client); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if room.IsAdmin(user.Username()) {
		t.Fatal("user should not be admin")
	}
}

func TestRoomRestoreAdmin(t *testing.T) {
	room := NewRoom("test")

	client := newTestClient(t, "anthony")
	room.Add(client)

	if room.IsAdmin("anthony") {
		t.Fatal("expected user to initially not be admin")
	}

	room.RestoreAdmin("anthony")

	if !room.IsAdmin("anthony") {
		t.Fatal("expected restored user to be admin")
	}
}

func TestRoomRestorePasswordHash(t *testing.T) {
	room := NewRoom("test")

	original := NewRoom("original")

	if err := original.SetPassword("password"); err != nil {
		t.Fatal(err)
	}

	hash := original.PasswordHash()

	room.RestorePasswordHash(hash)

	if !room.HasPassword() {
		t.Fatal("expected restored room to have a password")
	}

	if !room.CheckPassword("password") {
		t.Fatal("expected restored password to be valid")
	}

	if room.CheckPassword("wrong") {
		t.Fatal("expected wrong password to be rejected")
	}
}

func TestRoomNameAndSize(t *testing.T) {
	room := NewRoom("test")

	if room.Name() != "test" {
		t.Fatal("expected room name to be 'test'")
	}

	if room.Size() != 0 {
		t.Fatalf("expected room size to be 0, got %d", room.Size())
	}

	client := newTestClient(t, "anthony")

	room.Add(client)

	if room.Size() != 1 {
		t.Fatalf("expected room size to be 1, got %d", room.Size())
	}
}

func TestRoomTopic(t *testing.T) {
	room := NewRoom("test")

	if room.Topic() != "" {
		t.Fatalf("expected room topic to be empty, got: %q", room.Topic())
	}

	room.SetTopic("Welcome")

	if room.Topic() != "Welcome" {
		t.Fatalf(
			"expected topic to be %q, got %q",
			"Welcome",
			room.Topic(),
		)
	}
}

func TestRoomOwner(t *testing.T) {
	room := NewRoom("test")

	if room.Owner() != "" {
		t.Fatalf("expected room owner to be empty, got: %q", room.Owner())
	}

	room.SetOwner("anthony")

	if room.Owner() != "anthony" {
		t.Fatalf(
			"expected owner to be %q, got %q",
			"anthony",
			room.Owner(),
		)
	}
}

func TestRoomUsers(t *testing.T) {
	room := NewRoom("test")

	anthony := newTestClient(t, "anthony")
	bob := newTestClient(t, "bob")

	room.Add(anthony)
	room.Add(bob)

	users := room.Users()

	if len(users) != 2 {
		t.Fatalf(
			"expected room to have 2 users, got %d",
			len(users),
		)
	}

	found := map[string]bool{}

	for _, client := range users {
		found[client.User().Username()] = true
	}

	if !found["anthony"] {
		t.Fatal("expected room to have user 'anthony'")
	}

	if !found["bob"] {
		t.Fatal("expected room to have user 'bob'")
	}
}

func TestRoomAdmins(t *testing.T) {
	room := NewRoom("test")

	anthony := newTestClient(t, "anthony")
	bob := newTestClient(t, "bob")

	room.Add(anthony)
	room.Add(bob)

	if err := room.AddAdmin(anthony); err != nil {
		t.Fatal(err)
	}

	if err := room.AddAdmin(bob); err != nil {
		t.Fatal(err)
	}

	admins := room.Admins()

	if len(admins) != 2 {
		t.Fatalf("expected 2 admins, got %d", len(admins))
	}

	found := map[string]bool{}

	for _, client := range admins {
		found[client.User().Username()] = true
	}

	if !found["anthony"] {
		t.Fatal("expected anthony to be an admin")
	}

	if !found["bob"] {
		t.Fatal("expected bob to be an admin")
	}
}

func TestRoomAdminsUsernames(t *testing.T) {
	room := NewRoom("test")

	anthony := newTestClient(t, "anthony")
	bob := newTestClient(t, "bob")

	room.Add(anthony)
	room.Add(bob)

	if err := room.AddAdmin(anthony); err != nil {
		t.Fatal(err)
	}

	if err := room.AddAdmin(bob); err != nil {
		t.Fatal(err)
	}

	admins := room.AdminsUsernames()

	if len(admins) != 2 {
		t.Fatalf(
			"expected room to have 2 admins, got %d",
			len(admins),
		)
	}

	found := map[string]bool{}

	for _, admin := range admins {
		found[admin] = true
	}

	if !found["anthony"] {
		t.Fatal("expected room to have admin 'anthony'")
	}

	if !found["bob"] {
		t.Fatal("expected room to have admin 'bob'")
	}
}

func TestRoomAddAdminNotInRoom(t *testing.T) {
	room := NewRoom("test")

	client := newTestClient(t, "anthony")

	err := room.AddAdmin(client)

	if err == nil {
		t.Fatal("expected error when adding user not in room")
	}
}

func TestRoomAddAdminAlreadyAdmin(t *testing.T) {
	room := NewRoom("test")

	client := newTestClient(t, "anthony")

	room.Add(client)

	if err := room.AddAdmin(client); err != nil {
		t.Fatal(err)
	}

	if err := room.AddAdmin(client); err == nil {
		t.Fatal("expected error when adding user already admin")
	}
}

func TestRoomRemoveAdminNotInRoom(t *testing.T) {
	room := NewRoom("test")

	client := newTestClient(t, "anthony")

	err := room.RemoveAdmin(client)

	if err == nil {
		t.Fatal("expected error when removing user not in room")
	}
}

func TestRoomRemoveAdminNotAdmin(t *testing.T) {
	room := NewRoom("test")

	client := newTestClient(t, "anthony")

	room.Add(client)

	err := room.RemoveAdmin(client)
	if err == nil {
		t.Fatal("expected error when removing user who is not an admin")
	}
}

func TestRoomRequireOwner(t *testing.T) {
	room := NewRoom("test")

	owner := newTestClient(t, "anthony")
	client := newTestClient(t, "bob")

	room.Add(owner)
	room.Add(client)
	room.SetOwner("anthony")

	if err := room.RequireOwner(owner); err != nil {
		t.Fatalf("expected owner to have permission: %v", err)
	}

	if err := room.RequireOwner(client); err == nil {
		t.Fatal("expected client to not have permission")
	}
}

func TestRoomRequireOwnerGeneralRoom(t *testing.T) {
	room := NewRoom("general")

	owner := newTestClient(t, "anthony")
	room.Add(owner)
	room.SetOwner("anthony")

	if err := room.RequireOwner(owner); err == nil {
		t.Fatal("expected general room to reject owner operations")
	}
}

func TestRoomRequireAdmin(t *testing.T) {
	room := NewRoom("test")

	admin := newTestClient(t, "anthony")
	user := newTestClient(t, "bob")

	room.Add(admin)
	room.Add(user)

	if err := room.AddAdmin(admin); err != nil {
		t.Fatal(err)
	}

	if err := room.RequireAdmin(admin); err != nil {
		t.Fatalf("expected admin to have permission: %v", err)
	}

	if err := room.RequireAdmin(user); err == nil {
		t.Fatal("expected regular user to be rejected")
	}
}

func TestRoomCannotRemoveOwnerAsAdmin(t *testing.T) {
	room := NewRoom("test")

	owner := newTestClient(t, "anthony")
	room.Add(owner)
	room.SetOwner("anthony")

	if err := room.RemoveAdmin(owner); err == nil {
		t.Fatal("expected error when demoting owner")
	}
}

func TestRoomAddDuplicateClient(t *testing.T) {
	room := NewRoom("test")

	client := newTestClient(t, "anthony")

	room.Add(client)
	room.Add(client)

	if room.Size() != 1 {
		t.Fatalf("expected room size to remain 1, got %d", room.Size())
	}
}

func TestRoomRemoveClientNotInRoom(t *testing.T) {
	room := NewRoom("test")

	client := newTestClient(t, "anthony")

	room.Remove(client)

	if room.Size() != 0 {
		t.Fatalf("expected room size to remain 0, got %d", room.Size())
	}

	if client.Room() != nil {
		t.Fatal("expected client's room to remain nil")
	}
}

func TestRoomAdminPersistsAfterLeaving(t *testing.T) {
	room := NewRoom("test")

	client := newTestClient(t, "anthony")

	room.Add(client)

	if err := room.AddAdmin(client); err != nil {
		t.Fatal(err)
	}

	room.Remove(client)

	if room.Has(client) {
		t.Fatal("expected client to no longer be in room")
	}

	if !room.IsAdmin("anthony") {
		t.Fatal("expected client to remain an admin after leaving room")
	}
}

func TestRoomOwnerRemainsAdmin(t *testing.T) {
	room := NewRoom("test")

	owner := newTestClient(t, "anthony")

	room.Add(owner)
	room.SetOwner("anthony")

	if !room.IsAdmin("anthony") {
		t.Fatal("expected owner to be admin")
	}

	admins := room.AdminsUsernames()

	found := false
	for _, username := range admins {
		if username == "anthony" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected owner to appear in admin usernames")
	}
}

func newTestClient(t *testing.T, username string) *Client {

	t.Helper()

	user, err := NewUser(username, "password")
	if err != nil {
		t.Fatal(err)
	}

	return &Client{
		user: user,
	}
}
