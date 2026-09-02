package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"tuchat/protocol"
)

func (m Model) handleServerMessage(msg serverMsg) (tea.Model, tea.Cmd) {
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
		return m.handleJoinedRoom(msg)

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

	case "ping":
		return m, tea.Batch(
			sendCmd(m.connection.enc, protocol.Message{Type: "pong"}),
			listenCmd(m.connection.conn, m.connection.dec),
		)
	}
	return m, listenCmd(m.connection.conn, m.connection.dec)
}

func (m Model) handleJoinPasswordRequired(msg serverMsg) (tea.Model, tea.Cmd) {
	blinkCmd := m.beginRoomPassword(msg.Message)

	return m, tea.Batch(listenCmd(m.connection.conn, m.connection.dec), blinkCmd)
}

func (m Model) handleServerError(msg serverMsg) (tea.Model, tea.Cmd) {
	if m.screen == screenAuth {
		return m.handleAuthError(msg)
	}

	m.addChatEntry(chatEntry{
		kind: "error",
		text: msg.Message,
	})

	return m, listenCmd(m.connection.conn, m.connection.dec)

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

func (m Model) handleJoinedRoom(msg serverMsg) (tea.Model, tea.Cmd) {
	if m.connection.preserveHistory {
		if msg.RoomName == m.chat.roomName {
			m.connection.preserveHistory = false
		}
		return m, listenCmd(m.connection.conn, m.connection.dec)
	}
	m.chat.entries = nil
	m.chat.newMessages = 0

	return m, listenCmd(m.connection.conn, m.connection.dec)
}

func (m *Model) handleServerName(msg serverMsg) (tea.Model, tea.Cmd) {
	m.serverName = msg.Message

	m.refreshViewport()
	return m, listenCmd(m.connection.conn, m.connection.dec)
}
