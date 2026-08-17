package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case connectedMsg:
		return m.handleConnected(msg)

	case connErrMsg:
		return m.handleConnectionError(msg)

	case serverMsg:
		return m.handleServerMessage(msg)

	case sendErrMsg:
		return m.handleSendError(msg)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if m.connection.state == connectionDisconnected {
		return m.handleDisconnectedKey(msg)
	}

	if m.screen == screenAuth {
		return m.handleAuthKey(msg)
	}

	return m.handleChatKey(msg)
}

func (m Model) handleDisconnectedKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "q" {
		return m, tea.Quit
	}
	return m, nil
}
