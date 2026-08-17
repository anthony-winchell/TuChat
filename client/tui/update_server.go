package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleServerMessage(msg serverMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {

	case "auth_success":
		return m.handleAuthSuccess(msg)

	case "nick_success":
		return m.handleNickSuccess(msg)

	case "join_password_required":
		return m.handleJoinPasswordRequired(msg)

	case "chat", "pm", "system", "announcement",
		"welcome", "join", "leave":
		m.handleChatMessage(msg)

	case "users":
		m.chat.users = msg.Users

	case "rooms":
		m.chat.rooms = msg.Rooms

	case "roominfo":
		m.chat.roomName = msg.RoomName
		m.chat.roomTopic = msg.RoomTopic

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
