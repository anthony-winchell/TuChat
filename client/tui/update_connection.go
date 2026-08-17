package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"fmt"
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
	fmt.Println("DEBUG: connection disconnected")
	return m, nil
}

func (m Model) handleSendError(msg sendErrMsg) (tea.Model, tea.Cmd) {
	m.connection.state = connectionDisconnected
		m.err = msg.err
		return m, nil
}