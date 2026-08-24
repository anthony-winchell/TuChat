package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"tuchat/protocol"
)

func (m Model) switchSidebar() (tea.Model, tea.Cmd) {
	m.chat.activeSidebar =
		(m.chat.activeSidebar + 1) % 2

	if m.chat.activeSidebar == tabRooms {
		return m, sendCmd(
			m.connection.enc,
			protocol.Message{
				Type:    "command",
				Message: "/rooms",
			},
		)
	}

	return m, nil
}

func (m Model) handleRoomSidebarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.chat.awaitingRoomPassword != "" {
		return m.handleChatInputKey(msg)
	}

	switch msg.String() {

	case "up":
		if m.chat.selectedRoom > 0 {
			m.chat.selectedRoom--
		}
		return m, nil

	case "down":
		if m.chat.selectedRoom < len(m.chat.rooms)-1 {
			m.chat.selectedRoom++
		}
		return m, nil

	case "ctrl+j":
		return m.joinSelectedRoom()
	}

	return m.handleChatInputKey(msg)
}

func (m Model) joinSelectedRoom() (tea.Model, tea.Cmd) {
	if len(m.chat.rooms) == 0 {
		return m, nil
	}

	room := m.chat.rooms[m.chat.selectedRoom]

	if room.Name == m.chat.roomName {
		return m, nil
	}

	if room.HasPassword {
		m.beginRoomPassword(room.Name)
		return m, nil
	}

	return m, sendCmd(
		m.connection.enc,
		protocol.Message{
			Type:    "command",
			Message: "/join " + room.Name,
		},
	)
}

func (m *Model) beginRoomPassword(roomName string) {
	m.chat.awaitingRoomPassword = roomName

	m.chat.input.EchoMode = textinput.EchoPassword
	m.chat.input.Placeholder = "room password"
	m.chat.input.Prompt = "🔒 "

	m.addChatEntry(chatEntry{
		kind: "system",
		text: "#" + roomName +
			" requires a password. Enter it or press Esc to cancel.",
	})
}
