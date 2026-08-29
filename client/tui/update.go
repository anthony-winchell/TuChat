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

	default:
		return m.forwardToActiveWidget(msg)
	}
}

func (m Model) forwardToActiveWidget(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case m.screen == screenConnect:
		m.connect.input, cmd = m.connect.input.Update(msg)

	case m.screen == screenAuth:
		if m.auth.stage == stageMenu {
			m.auth.menu, cmd = m.auth.menu.Update(msg)
			return m, cmd
		}
		m.auth.input, cmd = m.auth.input.Update(msg)

	case m.chat.awaitingRoomPassword != "":
		m.chat.secret, cmd = m.chat.secret.Update(msg)

	default:
		m.chat.input, cmd = m.chat.input.Update(msg)
	}

	return m, cmd
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if m.connection.state == connectionDisconnected {
		return m.handleDisconnectedKey(msg)
	}

	if m.screen == screenConnect {
		return m.handleConnectKey(msg)
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
