package tui

import (
	"fmt"
	"strings"
	"time"
	"tuchat/protocol"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.MouseMsg:
		var cmd tea.Cmd
		m.chat.viewport, cmd = m.chat.viewport.Update(msg)
		if m.chat.viewport.AtBottom() {
			m.chat.newMessages = 0
		}

		return m, cmd

	case tea.WindowSizeMsg:
		m.ui.width = msg.Width
		m.ui.height = msg.Height

		const (
			sidebarWidth = 24
			gapWidth     = 2
			verticalGap  = 4
			headerHeight = 2
			inputHeight  = 3
		)

		m.auth.menu.SetSize(msg.Width-4, msg.Height-6)

		m.chat.viewport.Width = msg.Width - sidebarWidth - gapWidth - 6
		m.chat.viewport.Height = msg.Height - headerHeight - inputHeight - verticalGap

		m.refreshViewport()

		return m, nil

	case connectedMsg:
		m.connection.conn = msg.conn
		m.connection.dec = msg.decoder
		m.connection.enc = msg.encoder
		m.connection.state = connectionConnected

		return m, listenCmd(m.connection.dec)

	case connErrMsg:
		m.connection.state = connectionDisconnected
		m.err = msg.err
		fmt.Println("DEBUG: connection disconnected")
		return m, nil

	case serverMsg:
		switch msg.Type {

		case "auth_success":
			m.screen = screenChat
			m.chat.selfNickname = msg.Nickname

			m.auth.error = ""
			m.auth.stage = stageMenu
			m.auth.pendingUser = ""
			m.auth.input.Reset()
			m.auth.input.Blur()
			m.auth.input.Focus()

			return m, tea.Batch(listenCmd(m.connection.dec), sendCmd(m.connection.enc, protocol.Message{
				Type:    "command",
				Message: "/users",
			}), sendCmd(m.connection.enc, protocol.Message{
				Type:    "command",
				Message: "/room",
			}))

		case "nick_success":
			m.chat.selfNickname = msg.Nickname

		case "join_password_required":
			m.chat.awaitingRoomPassword = msg.Message

			m.chat.input.EchoMode = textinput.EchoPassword
			m.chat.input.Placeholder = "room password"
			m.chat.input.Prompt = "🔒 "

			m.chat.entries = append(m.chat.entries, chatEntry{
				kind: "system",
				text: "#" + msg.Message + " requires a password. Enter it or press Esc to cancel.",
			})
			m.refreshViewport()

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

		case "system", "announcement":
			m.addChatEntry(chatEntry{
				kind:      "system",
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

		case "users":
			m.chat.users = msg.Users

		case "rooms":
			m.chat.rooms = msg.Rooms

		case "roominfo":
			m.chat.roomName = msg.RoomName
			m.chat.roomTopic = msg.RoomTopic

		case "error":
			if m.screen == screenAuth {
				m.auth.error = msg.Message

				if m.auth.stage == stageAuthenticating && msg.Message == "invalid password" {
					m.auth.stage = stagePassword
					m.auth.input.EchoMode = textinput.EchoPassword
					m.auth.input.Placeholder = "password"
				} else {
					m.auth.stage = stageMenu
					m.auth.input.EchoMode = textinput.EchoNormal
					m.auth.input.Placeholder = "username"
					m.auth.input.Reset()
					m.auth.input.Blur()
				}
				return m, listenCmd(m.connection.dec)
			}
			m.chat.entries = append(m.chat.entries, chatEntry{
				kind: "error",
				text: msg.Message,
			})
			m.refreshViewport()
		}
		return m, listenCmd(m.connection.dec)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if m.connection.state == connectionDisconnected {
			if msg.String() == "q" {
				return m, tea.Quit
			}
			return m, nil
		}

		if msg.String() == "tab" {
			m.chat.activeSidebar = (m.chat.activeSidebar + 1) % 2
			if m.chat.activeSidebar == tabRooms {
				return m, sendCmd(m.connection.enc, protocol.Message{
					Type:    "command",
					Message: "/rooms",
				})
			}
			return m, nil
		}

		if m.screen == screenChat && m.chat.activeSidebar == tabRooms {
			if m.chat.awaitingRoomPassword != "" {
				return m.handleChatKey(msg)
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
				if len(m.chat.rooms) == 0 {
					return m, nil
				}

				room := m.chat.rooms[m.chat.selectedRoom]

				if room.Name == m.chat.roomName {
					return m, nil
				}

				if room.HasPassword {
					m.chat.awaitingRoomPassword = room.Name

					m.chat.input.EchoMode = textinput.EchoPassword
					m.chat.input.Placeholder = "room password"
					m.chat.input.Prompt = "🔒 "

					m.chat.entries = append(m.chat.entries, chatEntry{
						kind: "system",
						text: "#" + room.Name + " requires a password. Enter it or press Esc to cancel.",
					})
					m.refreshViewport()

					return m, nil
				}

				return m, sendCmd(m.connection.enc, protocol.Message{
					Type:    "command",
					Message: "/join " + room.Name,
				})

			default:
				return m.handleChatKey(msg)
			}
		}

		if m.screen == screenAuth {
			return m.handleAuthKey(msg)
		}

		return m.handleChatKey(msg)

	case sendErrMsg:
		m.connection.state = connectionDisconnected
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

func (m Model) handleAuthKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.auth.stage == stageMenu {
		if msg.String() == "enter" {
			selected, ok := m.auth.menu.SelectedItem().(authOption)
			if !ok {
				return m, nil
			}
			if selected == "Login" {
				m.auth.choice = "login"
			} else {
				m.auth.choice = "register"
			}
			m.auth.stage = stageUsername
			m.auth.input.Reset()
			m.auth.input.EchoMode = textinput.EchoNormal
			m.auth.input.Placeholder = "username"
			m.auth.input.Focus()

			return m, nil
		}

		var cmd tea.Cmd
		m.auth.menu, cmd = m.auth.menu.Update(msg)
		return m, cmd
	}

	if msg.String() == "esc" {
		switch m.auth.stage {
		case stageUsername:
			m.auth.stage = stageMenu
			m.auth.input.Blur()
			m.auth.input.Reset()
			m.auth.error = ""

		case stagePassword:
			m.auth.stage = stageUsername
			m.auth.input.Reset()
			m.auth.input.EchoMode = textinput.EchoNormal
			m.auth.input.Placeholder = "username"
			m.auth.error = ""
		}
		return m, nil
	}

	if msg.String() != "enter" {
		var cmd tea.Cmd
		m.auth.input, cmd = m.auth.input.Update(msg)
		return m, cmd
	}

	value := strings.TrimSpace(m.auth.input.Value())
	m.auth.input.Reset()

	switch m.auth.stage {

	case stageUsername:
		if value == "" {
			return m, nil
		}

		m.auth.pendingUser = value
		m.auth.stage = stagePassword
		m.auth.input.EchoMode = textinput.EchoPassword
		m.auth.input.Placeholder = "password"
		m.auth.input.Focus()

		return m, nil

	case stagePassword:
		if value == "" {
			return m, nil
		}

		m.auth.error = ""
		m.auth.stage = stageAuthenticating

		m.auth.input.EchoMode = textinput.EchoNormal
		m.auth.input.Placeholder = "username"

		return m, sendCmd(m.connection.enc, protocol.Message{
			Type:     m.auth.choice,
			Username: m.auth.pendingUser,
			Password: value,
		})
	}

	return m, nil
}

func (m Model) handleChatKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.chat.awaitingRoomPassword != "" && msg.String() == "esc" {
		m.chat.awaitingRoomPassword = ""

		m.chat.input.Reset()
		m.chat.input.EchoMode = textinput.EchoNormal
		m.chat.input.Placeholder = "message or /command"
		m.chat.input.Prompt = "> "

		m.chat.entries = append(m.chat.entries, chatEntry{
			kind: "system",
			text: "Cancelled join.",
		})
		m.refreshViewport()
		return m, nil
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

	if m.chat.awaitingRoomPassword != "" {
		room := m.chat.awaitingRoomPassword
		m.chat.awaitingRoomPassword = ""

		m.chat.input.Reset()
		m.chat.input.EchoMode = textinput.EchoNormal
		m.chat.input.Placeholder = "message or /command"
		m.chat.input.Prompt = "> "

		return m, sendCmd(m.connection.enc, protocol.Message{
			Type:    "command",
			Message: "/join " + room + " " + value,
		})
	}

	msgType := "chat"
	if strings.HasPrefix(value, "/") {
		msgType = "command"
	}

	return m, sendCmd(m.connection.enc, protocol.Message{
		Type:    msgType,
		Message: value,
	})
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return "[" + t.Format("15:04") + "] "
}

func (m *Model) refreshViewport() {
	m.chat.viewport.SetContent(renderEntries(m.chat.entries))
}

func (m *Model) addChatEntry(entry chatEntry) {
	wasAtBottom := m.chat.viewport.AtBottom()

	m.chat.entries = append(m.chat.entries, entry)
	m.chat.viewport.SetContent(renderEntries(m.chat.entries))

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
