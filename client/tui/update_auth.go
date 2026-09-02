package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"tuchat/protocol"
)

func (m Model) handleAuthKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.auth.stage == stageMenu {
		return m.handleAuthMenuKey(msg)
	}

	if msg.String() == "esc" {
		return m.handleAuthEscape()
	}

	if msg.String() != "enter" {
		var cmd tea.Cmd
		m.auth.input, cmd = m.auth.input.Update(msg)
		return m, cmd
	}

	return m.submitAuthInput()
}

func (m Model) handleAuthMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

		return m, m.auth.input.Focus()
	}

	var cmd tea.Cmd
	m.auth.menu, cmd = m.auth.menu.Update(msg)

	return m, cmd
}

func (m Model) handleAuthEscape() (tea.Model, tea.Cmd) {
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

func (m Model) submitAuthInput() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.auth.input.Value())
	m.auth.input.Reset()

	if value == "" {
		return m, nil
	}

	switch m.auth.stage {

	case stageServerPassword:
		return m, sendCmd(m.connection.enc,
			protocol.Message{
				Type:     "server_password",
				Password: value,
			})

	case stageUsername:
		m.auth.pendingUser = value
		m.auth.stage = stagePassword

		m.auth.input.EchoMode = textinput.EchoPassword
		m.auth.input.Placeholder = "password"

		return m, nil

	case stagePassword:
		m.auth.error = ""
		m.auth.stage = stageAuthenticating

		m.auth.input.EchoMode = textinput.EchoNormal
		m.auth.input.Placeholder = "username"

		m.connection.creds = reconnectCreds{
			choice:   m.auth.choice,
			username: m.auth.pendingUser,
			password: value,
		}

		return m, sendCmd(
			m.connection.enc,
			protocol.Message{
				Type:     m.auth.choice,
				Username: m.auth.pendingUser,
				Password: value,
			},
		)
	}

	return m, nil
}

func (m Model) handleAuthSuccess(msg serverMsg) (tea.Model, tea.Cmd) {
	restoring := m.chat.roomName != ""

	m.screen = screenChat
	m.chat.selfNickname = msg.Nickname

	m.auth.error = ""
	m.auth.stage = stageMenu
	m.auth.pendingUser = ""
	m.auth.input.Reset()
	m.auth.input.Blur()

	cmds := []tea.Cmd{
		listenCmd(m.connection.conn, m.connection.dec),
		sendCmd(m.connection.enc, protocol.Message{
			Type:    "command",
			Message: "/users",
		}),
		sendCmd(m.connection.enc, protocol.Message{
			Type:    "command",
			Message: "/room",
		}),
		m.chat.input.Focus(),
	}

	if restoring {
		cmds = append(cmds,
			sendCmd(m.connection.enc, protocol.Message{
				Type:    "command",
				Message: "/nick " + m.chat.selfNickname,
			}),
			sendCmd(m.connection.enc, protocol.Message{
				Type:    "command",
				Message: "/join " + m.chat.roomName,
			}),
		)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleAuthError(msg serverMsg) (tea.Model, tea.Cmd) {
	m.auth.error = msg.Message

	if m.auth.stage == stageServerPassword {
		m.auth.input.Reset()
		m.auth.input.EchoMode = textinput.EchoPassword
		m.auth.input.Placeholder = "server password"
		return m, listenCmd(m.connection.conn, m.connection.dec)
	}

	if m.auth.stage == stageAuthenticating &&
		msg.Message == protocol.AuthInvalidCredentials {

		m.auth.stage = stagePassword
		m.auth.input.EchoMode = textinput.EchoPassword
		m.auth.input.Placeholder = "password"

		return m, listenCmd(m.connection.conn, m.connection.dec)
	}

	m.auth.stage = stageMenu
	m.auth.input.EchoMode = textinput.EchoNormal
	m.auth.input.Placeholder = "username"
	m.auth.input.Reset()
	m.auth.input.Blur()

	return m, listenCmd(m.connection.conn, m.connection.dec)
}
