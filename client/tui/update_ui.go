package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.ui.width = msg.Width
	m.ui.height = msg.Height
	m.ui.layout = computeLayout(msg.Width, msg.Height)

	m.auth.menu.SetSize(msg.Width-4, msg.Height-6)

	m.chat.viewport.Width = m.ui.layout.viewportWidth
	m.chat.viewport.Height = m.ui.layout.viewportHeight

	m.refreshViewport()

	return m, nil
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.chat.viewport, cmd = m.chat.viewport.Update(msg)

	if m.chat.viewport.AtBottom() {
		m.chat.newMessages = 0
	}

	return m, cmd
}

func (m *Model) refreshViewport() {
	m.chat.viewport.SetContent(renderEntries(m.chat.entries, m.ui.layout.viewportWidth))
}
