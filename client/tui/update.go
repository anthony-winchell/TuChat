package tui

import(
	tea "github.com/charmbracelet/bubbletea"
	"tuchat/protocol"
	"strings"
	"time"
)

type screen int

const (
	screenAuth screen = iota
	screenChat
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		sidebarWidth := 20
		m.viewport.Width = msg.Width - sidebarWidth - 4
		m.viewport.Height = msg.Height - 6

		m.authMenu.SetSize(msg.Width-4, msg.Height-6)

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
			return m, tea.Batch(listenCmd(m.decoder), sendCmd(m.encoder, protocol.Message{
				Type:    "command",
				Message: "/users",
			}), sendCmd(m.encoder, protocol.Message{
				Type:    "command",
				Message: "/room",
			}))

		case "chat":
			m.chatLog = append(m.chatLog, formatTime(msg.Timestamp)+msg.Nickname+": "+msg.Message)
			m.refreshViewport()

		case "pm":
			m.chatLog = append(m.chatLog, formatTime(msg.Timestamp)+"[PM "+msg.Nickname+" -> "+msg.Target+"] "+msg.Message)
			m.refreshViewport()

		case "system", "welcome", "announcement":
			m.chatLog = append(m.chatLog, msg.Message)
			m.refreshViewport()

		case "join":
			m.chatLog = append(m.chatLog, msg.Nickname+" joined the chat")
			m.refreshViewport()

		case "leave":
			m.chatLog = append(m.chatLog, msg.Nickname+" left the chat")
			m.refreshViewport()

		case "users":
			m.users = msg.Users

		case "rooms":
			m.rooms = msg.Rooms

		case "roominfo":
			m.roomName = msg.Message
			m.roomTopic = msg.RoomTopic

		case "error":
			if m.screen == screenAuth {
				m.authError = msg.Message
			} else {
				m.chatLog = append(m.chatLog, msg.Message)
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

	msgType := "chat"
	if strings.HasPrefix(value, "/") {
		msgType = "command"
	}

	if msgType == "chat" {
		m.chatLog = append(m.chatLog, formatTime(time.Now())+"you: "+value)
		m.refreshViewport()
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
	m.viewport.SetContent(strings.Join(m.chatLog, "\n"))
	m.viewport.GotoBottom()
}

