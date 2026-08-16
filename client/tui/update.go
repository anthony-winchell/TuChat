package tui

import (
	"strings"
	"time"
	"tuchat/protocol"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenAuth screen = iota
	screenChat
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.MouseMsg:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		if m.viewport.AtBottom(){
			m.newMessages = 0
		}

		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		const (
			sidebarWidth = 24
			gapWidth     = 2
			verticalGap  = 4
			headerHeight = 2
			inputHeight  = 3
		)

		m.authMenu.SetSize(msg.Width-4, msg.Height-6)

		m.viewport.Width = msg.Width - sidebarWidth - gapWidth - 6
		m.viewport.Height = msg.Height - headerHeight - inputHeight - verticalGap

		m.refreshViewport()

		return m, nil

	case connectedMsg:
		m.decoder = msg.decoder
		m.encoder = msg.encoder
		return m, listenCmd(m.decoder)

	case connErrMsg:
		m.err = msg.err
		return m, nil

	case serverMsg:
		switch msg.Type {
		case "auth_success":
			m.screen = screenChat
			m.selfNickname = msg.Nickname
			return m, tea.Batch(listenCmd(m.decoder), sendCmd(m.encoder, protocol.Message{
				Type:    "command",
				Message: "/users",
			}), sendCmd(m.encoder, protocol.Message{
				Type:    "command",
				Message: "/room",
			}))

		case "nick_success":
			m.selfNickname = msg.Nickname

		case "join_password_required":
			m.awaitingRoomPassword = msg.Message

			m.chatInput.EchoMode = textinput.EchoPassword
			m.chatInput.Placeholder = "room password"
			m.chatInput.Prompt = "🔒 "

			m.chatLog = append(m.chatLog, chatEntry{
				kind: "system",
				text: "#" + msg.Message + " requires a password. Enter it or press Esc to cancel.",
			})
			m.refreshViewport()

		case "chat":
			m.addChatEntry(chatEntry{
				kind: "chat",
				timestamp: msg.Timestamp,
				nickname:  msg.Nickname,
				text:      msg.Message,
				self:      msg.Nickname == m.selfNickname,
			})

		case "pm":
			m.addChatEntry(chatEntry{
				kind:      "pm",
				timestamp: msg.Timestamp,
				nickname:  msg.Nickname,
				target:    msg.Target,
				text:      msg.Message,
				self:      msg.Nickname == m.selfNickname,
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
			m.users = msg.Users

		case "rooms":
			m.rooms = msg.Rooms

		case "roominfo":
			m.roomName = msg.RoomName
			m.roomTopic = msg.RoomTopic

		case "error":
			if m.screen == screenAuth {
				m.authError = msg.Message
			} else {
				m.chatLog = append(m.chatLog, chatEntry{
					kind: "error",
					text: msg.Message,
				})
				m.refreshViewport()
			}
		}
		return m, listenCmd(m.decoder)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if msg.String() == "tab" {
			m.activeSidebar = (m.activeSidebar + 1) % 2
			if m.activeSidebar == tabRooms {
				return m, sendCmd(m.encoder, protocol.Message{
					Type:    "command",
					Message: "/rooms",
				})
			}
			return m, nil
		}

		if m.screen == screenChat && m.activeSidebar == tabRooms {
			if m.awaitingRoomPassword != "" {
				return m.handleChatKey(msg)
			}

			switch msg.String() {
			case "up":
				if m.selectedRoom > 0 {
					m.selectedRoom--
				}
				return m, nil
			case "down":
				if m.selectedRoom < len(m.rooms)-1 {
					m.selectedRoom++
				}
				return m, nil

			case "ctrl+j":
				if len(m.rooms) == 0 {
					return m, nil
				}

				room := m.rooms[m.selectedRoom]

				if room.Name == m.roomName {
					return m, nil
				}

				if room.HasPassword {
					m.awaitingRoomPassword = room.Name

					m.chatInput.EchoMode = textinput.EchoPassword
					m.chatInput.Placeholder = "room password"
					m.chatInput.Prompt = "🔒 "

					m.chatLog = append(m.chatLog, chatEntry{
						kind: "system",
						text: "#" + room.Name + " requires a password. Enter it or press Esc to cancel.",
					})
					m.refreshViewport()

					return m, nil
				}

				return m, sendCmd(m.encoder, protocol.Message{
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

	case sentMsg:
		return m, nil

	case sendErrMsg:
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

func (m Model) handleAuthKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.authStage == stageMenu {
		if msg.String() == "enter" {
			selected, ok := m.authMenu.SelectedItem().(authOption)
			if !ok {
				return m, nil
			}
			if selected == "Login" {
				m.authChoice = "login"
			} else {
				m.authChoice = "register"
			}
			m.authStage = stageUsername
			m.input.Placeholder = "username"
			return m, nil
		}

		var cmd tea.Cmd
		m.authMenu, cmd = m.authMenu.Update(msg)
		return m, cmd
	}

	if msg.String() != "enter" {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	value := strings.TrimSpace(m.input.Value())
	m.input.Reset()

	switch m.authStage {
	case stageUsername:
		m.input.Placeholder = "password"
		if value == "" {
			return m, nil
		}
		m.pendingUser = value
		m.authStage = stagePassword
		return m, nil

	case stagePassword:
		m.authStage = stageMenu
		m.authError = ""
		return m, sendCmd(m.encoder, protocol.Message{
			Type:     m.authChoice,
			Username: m.pendingUser,
			Password: value,
		})
	}

	return m, nil
}

func (m Model) handleChatKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.awaitingRoomPassword != "" && msg.String() == "esc" {
		m.awaitingRoomPassword = ""

		m.chatInput.Reset()
		m.chatInput.EchoMode = textinput.EchoNormal
		m.chatInput.Placeholder = "message or /command"
		m.chatInput.Prompt = "> "

		m.chatLog = append(m.chatLog, chatEntry{
			kind: "system",
			text: "Cancelled join.",
		})
		m.refreshViewport()
		return m, nil
	}

	if msg.String() != "enter" {
		var cmd tea.Cmd
		m.chatInput, cmd = m.chatInput.Update(msg)
		return m, cmd
	}

	value := strings.TrimSpace(m.chatInput.Value())
	m.chatInput.Reset()

	if value == "" {
		return m, nil
	}

	if m.awaitingRoomPassword != "" {
		room := m.awaitingRoomPassword
		m.awaitingRoomPassword = ""

		m.chatInput.Reset()
		m.chatInput.EchoMode = textinput.EchoNormal
		m.chatInput.Placeholder = "message or /command"
		m.chatInput.Prompt = "> "

		return m, sendCmd(m.encoder, protocol.Message{
			Type:    "command",
			Message: "/join " + room + " " + value,
		})
	}

	msgType := "chat"
	if strings.HasPrefix(value, "/") {
		msgType = "command"
	}

	return m, sendCmd(m.encoder, protocol.Message{
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
	m.viewport.SetContent(renderEntries(m.chatLog))
}

func (m *Model) addChatEntry(entry chatEntry) {
	wasAtBottom := m.viewport.AtBottom()

	m.chatLog = append(m.chatLog, entry)
	m.viewport.SetContent(renderEntries(m.chatLog))

	if wasAtBottom {
		m.viewport.GotoBottom()
		m.newMessages = 0
	} else {
		m.newMessages++
	}
}

func renderWelcomeMessage(msg serverMsg) string {
	return strings.NewReplacer(
		"{server}", msg.ServerName,
		"{nickname}", msg.Nickname,
	).Replace(msg.Message)
}

