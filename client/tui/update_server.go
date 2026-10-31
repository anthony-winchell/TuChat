package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleServerMessage is the only place that re-arms listenCmd after a
// server message. Handlers must not return listenCmd themselves: a missing
// one stops the client reading, and a duplicate means two concurrent reads
// on the same decoder.
func (m Model) handleServerMessage(msg serverMsg) (tea.Model, tea.Cmd) {
	next, cmd := m.dispatchServerMessage(msg)
	return next, tea.Batch(cmd, listenCmd(m.connection.dec))
}

func (m Model) dispatchServerMessage(msg serverMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {

	case "server_password_prompt":
		return m.handleServerPasswordPrompt(msg)

	case "auth_prompt":
		return m.handleAuthPrompt(msg)

	case "auth_success":
		return m.handleAuthSuccess(msg)

	case "nick_success":
		return m.handleNickSuccess(msg)

	case "room_joined":
		return m.handleJoinedRoom()

	case "join_password_required":
		return m.handleJoinPasswordRequired(msg)

	case "chat", "pm", "system", "announcement",
		"welcome", "join", "leave":
		m.handleChatMessage(msg)

	case "users":
		m.chat.users = msg.Users

	case "rooms":
		m.handleRoomSelection(msg)

	case "roominfo":
		m.chat.roomName = msg.RoomName
		m.chat.roomTopic = msg.RoomTopic

	case "server_name":
		m.handleServerName(msg)

	case "error":
		return m.handleServerError(msg)
	}
	return m, nil
}

func (m Model) handleJoinPasswordRequired(msg serverMsg) (tea.Model, tea.Cmd) {
	return m, m.beginRoomPassword(msg.Message)
}

func (m Model) handleServerError(msg serverMsg) (tea.Model, tea.Cmd) {
	if m.screen == screenAuth {
		return m.handleAuthError(msg)
	}

	m.addChatEntry(chatEntry{
		kind: "error",
		text: msg.Message,
	})

	return m, nil
}

func (m *Model) handleRoomSelection(msg serverMsg) {
	selectedRoom := ""

	if m.chat.selectedRoom >= 0 &&
		m.chat.selectedRoom < len(m.chat.rooms) {
		selectedRoom = m.chat.rooms[m.chat.selectedRoom].Name
	}

	m.chat.rooms = msg.Rooms

	m.chat.selectedRoom = 0

	for i, room := range m.chat.rooms {
		if room.Name == selectedRoom {
			m.chat.selectedRoom = i
			break
		}
	}

}

func (m Model) handleJoinedRoom() (tea.Model, tea.Cmd) {
	m.chat.entries = nil
	m.chat.newMessages = 0

	return m, nil
}

func (m *Model) handleServerName(msg serverMsg) {
	m.serverName = msg.Message

	m.refreshViewport()
}
