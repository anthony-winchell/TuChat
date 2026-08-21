package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	return m, listenCmd(m.connection.dec)
}

func (m Model) handleJoinPasswordRequired(msg serverMsg) (tea.Model, tea.Cmd) {
	m.chat.awaitingRoomPassword = msg.Message

	m.chat.input.EchoMode = textinput.EchoPassword
	m.chat.input.Placeholder = "room password"
	m.chat.input.Prompt = "🔒 "

	m.chat.entries = append(m.chat.entries, chatEntry{
		kind: "system",
		text: "#" + msg.Message + " requires a password. Enter it or press Esc to cancel.",
	})
	m.refreshViewport()

	return m, m.chat.input.Focus()
}

func (m Model) handleServerError(msg serverMsg) (tea.Model, tea.Cmd) {
	if m.screen == screenAuth {
		return m.handleAuthError(msg)
	}

	m.addChatEntry(chatEntry{
		kind: "error",
		text: msg.Message,
	})

	return m, listenCmd(m.connection.dec)

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

	return m, listenCmd(m.connection.dec)
}


func (m *Model) handleServerName(msg serverMsg) (tea.Model, tea.Cmd) {
	m.serverName = msg.Message

	m.refreshViewport()
	return m, listenCmd(m.connection.dec)
}
