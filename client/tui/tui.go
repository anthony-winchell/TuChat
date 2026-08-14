package tui

import (
	"encoding/json"
	"net"
	"strings"
	"time"
	"tuchat/protocol"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"fmt"
)

const serverAddr = "localhost:8080"

type Model struct {
	screen    screen
	authStage authStage

	authChoice  string
	pendingUser string

	authMenu list.Model

	chatLog   []string
	viewport  viewport.Model
	width     int
	height    int
	chatInput textinput.Model

	activeSidebar sidebarTab
	users         []protocol.UserSummary
	rooms         []protocol.RoomSummary

	decoder *json.Decoder
	encoder *json.Encoder
	input   textinput.Model
	err     error
}

type serverMsg protocol.Message

type screen int

const (
	screenAuth screen = iota
	screenChat
)

type authStage int

const (
	stageMenu authStage = iota
	stageUsername
	stagePassword
)

type connectedMsg struct {
	decoder *json.Decoder
	encoder *json.Encoder
}

type connErrMsg struct {
	err error
}

type sentMsg struct{}

type sendErrMsg struct {
	err error
}

type authOption string

type sidebarTab int

const (
	tabUsers sidebarTab = iota
	tabRooms
)

func (m Model) renderSidebar() string {
	var b strings.Builder

	switch m.activeSidebar {
	case tabUsers: 
	  b.WriteString("USERS\n")
		for _, user := range m.users {
			line := "• " + user.Nickname
			if user.Owner {
				line += " (owner)"
			} else if user.Admin {
				line += " (admin)"
			}
			b.WriteString(line + "\n")
		}
	case tabRooms:
		b.WriteString("ROOMS\n")
		for _, r := range m.rooms {
			b.WriteString(fmt.Sprintf("• %s (%d)\n", r.Name, r.Users))
		}
	}

	return b.String()
}

func (a authOption) Title() string {
	return string(a)
}

func (a authOption) Description() string {
	return ""
}

func (a authOption) FilterValue() string {
	return string(a)
}

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "type your message here"
	ti.Focus()

	items := []list.Item{
		authOption("Login"),
		authOption("Register"),
	}

	authMenu := list.New(items, list.NewDefaultDelegate(), 0, 0)
	authMenu.Title = "TuChat"
	authMenu.SetSize(30, 10)
	authMenu.SetShowStatusBar(false)
	authMenu.SetFilteringEnabled(false)

	vp := viewport.New(0, 0)
	vp.Width = 60
	vp.Height = 15

	ci := textinput.New()
	ci.Placeholder = "message or /command"
	ci.Focus()

	return Model{
		input:     ti,
		authMenu:  authMenu,
		viewport:  vp,
		chatInput: ci,
	}
}

func connectCmd() tea.Cmd {
	return func() tea.Msg {
		conn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			return connErrMsg{err: err}
		}
		return connectedMsg{
			decoder: json.NewDecoder(conn),
			encoder: json.NewEncoder(conn),
		}
	}
}

func listenCmd(decoder *json.Decoder) tea.Cmd {
	return func() tea.Msg {
		var msg protocol.Message
		if err := decoder.Decode(&msg); err != nil {
			return connErrMsg{err: err}
		}
		return serverMsg(msg)
	}
}

func sendCmd(encoder *json.Encoder, msg protocol.Message) tea.Cmd {
	return func() tea.Msg {
		if err := encoder.Encode(msg); err != nil {
			return sendErrMsg{err: err}
		}
		return sentMsg{}
	}
}

func (m *Model) refreshViewport() {
	m.viewport.SetContent(strings.Join(m.chatLog, "\n"))
	m.viewport.GotoBottom()
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return "[" + t.Format("15:04") + "] "
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
		if value == "" {
			return m, nil
		}
		m.pendingUser = value
		m.authStage = stagePassword
		return m, nil

	case stagePassword:
		m.authStage = stageMenu
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

func (m Model) Init() tea.Cmd {
	return connectCmd()
}

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

		case "error":
			m.chatLog = append(m.chatLog, "Error: "+msg.Message)
			m.refreshViewport()
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

func (m Model) View() string {
	if m.err != nil {
		return "connection error: " + m.err.Error()
	}

	if m.screen == screenChat {
		chatPane := m.viewport.View()
		sidebar := m.renderSidebar()

		body := lipgloss.JoinHorizontal(
			lipgloss.Top,
			chatPane,
			sidebar,
		)

		return body + "\n\n" + m.chatInput.View()
	}

	if m.authStage == stageMenu {
		return m.authMenu.View()
	}

	label := m.authChoice + " - username:"
	if m.authStage == stagePassword {
		label = m.authChoice + " - password:"
	}

	return label + "\n\n" + m.input.View()
}
