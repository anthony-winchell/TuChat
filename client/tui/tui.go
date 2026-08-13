package tui

import (
	"encoding/json"
	"net"
	"strings"
	"tuchat/protocol"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const serverAddr = "localhost:8080"

type Model struct {
	screen    screen
	authStage authStage

	authChoice  string
	pendingUser string

	authMenu list.Model

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

	return Model{
		input:    ti,
		authMenu: authMenu,
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

func (m Model) Init() tea.Cmd {
	return connectCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

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
		}
		return m, listenCmd(m.decoder)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if m.screen == screenAuth {
			return m.handleAuthKey(msg)
		}

		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd

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
		return "you're in chat now\n"
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
