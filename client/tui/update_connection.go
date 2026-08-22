package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleConnected(msg connectedMsg) (tea.Model, tea.Cmd) {
	m.connection.conn = msg.conn
	m.connection.dec = msg.decoder
	m.connection.enc = msg.encoder
	m.connection.state = connectionConnected

	return m, listenCmd(m.connection.dec)
}

func (m Model) handleConnectionError(msg connErrMsg) (tea.Model, tea.Cmd) {
	m.connection.state = connectionDisconnected
	m.err = msg.err
	return m, nil
}

func (m Model) handleSendError(msg sendErrMsg) (tea.Model, tea.Cmd) {
	m.connection.state = connectionDisconnected
	m.err = msg.err
	return m, nil
}

func (m Model) handleServerPasswordPrompt(msg serverMsg) (tea.Model, tea.Cmd) {
	m.auth.stage = stageServerPassword
	m.auth.error = ""

	m.auth.input.Reset()
	m.auth.input.EchoMode = textinput.EchoPassword
	m.auth.input.Placeholder = "server password"
	m.auth.input.Focus()

	return m, listenCmd(m.connection.dec)
}

func (m Model) handleAuthPrompt(msg serverMsg) (tea.Model, tea.Cmd) {
	if m.auth.stage == stageServerPassword {
		m.auth.stage = stageMenu
		m.auth.error = ""
		m.auth.input.Reset()
		m.auth.input.Blur()
	}

	return m, listenCmd(m.connection.dec)
}
