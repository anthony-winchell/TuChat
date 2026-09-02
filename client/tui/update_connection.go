package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"tuchat/protocol"
)

func (m Model) handleConnected(msg connectedMsg) (tea.Model, tea.Cmd) {
	m.connection.conn = msg.conn
	m.connection.dec = msg.decoder
	m.connection.enc = msg.encoder
	m.connection.state = connectionConnected

	restoring := m.screen == screenChat
	m.screen = screenAuth
	if restoring {
		m.connection.preserveHistory = true
	}

	return m, listenCmd(m.connection.conn, m.connection.dec)
}

func (m Model) handleConnectionError(msg connErrMsg) (tea.Model, tea.Cmd) {
	if m.screen == screenConnect {
		m.connect.error = msg.err.Error()
		return m, m.connect.input.Focus()
	}
	if m.chat.roomName != "" {
		m.connection.state = connectionReconnecting
		m.connection.reconnectAttempt++
		return m, reconnectCmd(m.connection.reconnectAttempt)
	}
	m.connection.state = connectionDisconnected
	m.err = msg.err
	return m, nil
}

func (m Model) handleSendError(msg sendErrMsg) (tea.Model, tea.Cmd) {
	if m.chat.roomName != "" {
		m.connection.state = connectionReconnecting
		m.connection.reconnectAttempt++
		return m, reconnectCmd(m.connection.reconnectAttempt)
	}
	m.connection.state = connectionDisconnected
	m.err = msg.err
	return m, nil
}

func (m Model) handleReconnect(msg reconnectMsg) (tea.Model, tea.Cmd) {
	m.connection.reconnectAttempt = msg.attempt
	m.connection.state = connectionConnecting
	return m, connectCmd(m.connection.addr)
}

func (m Model) handleServerPasswordPrompt(msg serverMsg) (tea.Model, tea.Cmd) {
	m.screen = screenAuth
	m.auth.stage = stageServerPassword
	m.auth.error = ""

	m.auth.input.Reset()
	m.auth.input.EchoMode = textinput.EchoPassword
	m.auth.input.Placeholder = "server password"
	blinkCmd := m.auth.input.Focus()

	return m, tea.Batch(listenCmd(m.connection.conn, m.connection.dec), blinkCmd)
}

func (m Model) handleAuthPrompt(msg serverMsg) (tea.Model, tea.Cmd) {
	if m.auth.stage == stageServerPassword {
		m.auth.stage = stageMenu
		m.auth.error = ""
		m.auth.input.Reset()
		m.auth.input.Blur()
	}

	if m.chat.roomName != "" && m.connection.creds.choice != "" {
		return m, tea.Batch(
			listenCmd(m.connection.conn, m.connection.dec),
			sendCmd(m.connection.enc, protocol.Message{
				Type:     m.connection.creds.choice,
				Username: m.connection.creds.username,
				Password: m.connection.creds.password,
			}),
		)
	}

	return m, listenCmd(m.connection.conn, m.connection.dec)
}
