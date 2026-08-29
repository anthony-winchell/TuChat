package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleConnectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		return m.submitConnect()
	}

	var cmd tea.Cmd
	m.connect.input, cmd = m.connect.input.Update(msg)

	return m, cmd
}

func (m Model) submitConnect() (tea.Model, tea.Cmd) {
	m.connect.error = ""
	value := strings.TrimSpace(m.connect.input.Value())
	if value == "" {
		return m, nil
	}

	m.connection.addr = value
	return m, connectCmd(value)
}
