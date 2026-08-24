package tui

import (
	"strings"
	"tuchat/protocol"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) handleChatMessage(msg serverMsg) {
	switch msg.Type {

	case "chat":
		m.addChatEntry(chatEntry{
			kind:      "chat",
			timestamp: msg.Timestamp,
			nickname:  msg.Nickname,
			text:      msg.Message,
			self:      msg.Nickname == m.chat.selfNickname,
		})

	case "pm":
		m.addChatEntry(chatEntry{
			kind:      "pm",
			timestamp: msg.Timestamp,
			nickname:  msg.Nickname,
			target:    msg.Target,
			text:      msg.Message,
			self:      msg.Nickname == m.chat.selfNickname,
		})

	case "system":
		m.addChatEntry(chatEntry{
			kind:      "system",
			text:      msg.Message,
			timestamp: msg.Timestamp,
		})

	case "announcement":
		m.addChatEntry(chatEntry{
			kind:      "announcement",
			text:      msg.Message,
			timestamp: msg.Timestamp,
		})

	case "welcome":
		m.addChatEntry(chatEntry{
			kind:      "welcome",
			text:      renderWelcomeMessage(msg),
			timestamp: msg.Timestamp,
		})

	case "join":
		m.addChatEntry(chatEntry{
			kind:      "join",
			nickname:  msg.Nickname,
			timestamp: msg.Timestamp,
		})

	case "leave":
		m.addChatEntry(chatEntry{
			kind:      "leave",
			nickname:  msg.Nickname,
			timestamp: msg.Timestamp,
		})
	}
}

func (m Model) handleChatKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "tab" {
		return m.switchSidebar()
	}

	if m.chat.activeSidebar == tabRooms {
		return m.handleRoomSidebarKey(msg)
	}

	return m.handleChatInputKey(msg)
}

func (m Model) handleChatInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.chat.awaitingRoomPassword != "" {
		return m.handleRoomPasswordKey(msg)
	}

	if msg.String() != "enter" {
		var cmd tea.Cmd

		m.chat.input, cmd = m.chat.input.Update(msg)

		return m, cmd
	}

	value := strings.TrimSpace(m.chat.input.Value())
	m.chat.input.Reset()

	if value == "" {
		return m, nil
	}

	if value == "/clear" {
		m.chat.entries = nil
		m.chat.newMessages = 0
		m.addChatEntry(
			chatEntry{
				kind: "system",
				text: "Cleared chat history. Use /history to recover.",
			},
		)
		return m, nil
	}

	msgType := "chat"

	if strings.HasPrefix(value, "/") {
		msgType = "command"
	}

	return m, sendCmd(
		m.connection.enc,
		protocol.Message{
			Type:    msgType,
			Message: value,
		},
	)
}

func (m Model) handleRoomPasswordKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.cancelRoomPassword()
		return m, nil
	}

	if msg.String() != "enter" {
		var cmd tea.Cmd

		m.chat.input, cmd = m.chat.input.Update(msg)

		return m, cmd
	}

	password := strings.TrimSpace(m.chat.input.Value())

	if password == "" {
		return m, nil
	}

	room := m.chat.awaitingRoomPassword

	m.resetChatInput()

	return m, sendCmd(
		m.connection.enc,
		protocol.Message{
			Type:    "command",
			Message: "/join " + room + " " + password,
		},
	)
}

func (m *Model) cancelRoomPassword() {
	m.chat.awaitingRoomPassword = ""
	m.resetChatInput()

	m.addChatEntry(chatEntry{
		kind: "system",
		text: "Cancelled join.",
	})
}

func (m *Model) resetChatInput() {
	m.chat.input.Reset()
	m.chat.input.EchoMode = textinput.EchoNormal
	m.chat.input.Placeholder = "message or /command"
	m.chat.input.Prompt = "> "
}

func (m Model) handleNickSuccess(msg serverMsg) (tea.Model, tea.Cmd) {
	m.chat.selfNickname = msg.Nickname

	return m, nil
}

func (m *Model) addChatEntry(entry chatEntry) {
	wasAtBottom := m.chat.viewport.AtBottom()

	m.chat.entries = append(m.chat.entries, entry)
	m.chat.viewport.SetContent(renderEntries(m.chat.entries, m.ui.layout.viewportWidth))

	if wasAtBottom {
		m.chat.viewport.GotoBottom()
		m.chat.newMessages = 0
	} else {
		m.chat.newMessages++
	}
}

func renderWelcomeMessage(msg serverMsg) string {
	return strings.NewReplacer(
		"{server}", msg.ServerName,
		"{nickname}", msg.Nickname,
	).Replace(msg.Message)
}

const (
	tsWidth    = 8
	nickCol    = 8
	chatIndent = tsWidth + nickCol
)

func wrapAnsi(text string, width, indent int) string {
	if width <= indent {
		return text
	}

	wrapped := lipgloss.NewStyle().Width(width - indent).Render(text)
	if indent == 0 {
		return wrapped
	}

	return strings.ReplaceAll(wrapped, "\n", "\n"+strings.Repeat(" ", indent))
}
