package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.ui.width = msg.Width
	m.ui.height = msg.Height

	const (
		sidebarWidth = 24
		gapWidth     = 2
		verticalGap  = 4
		headerHeight = 2
		inputHeight  = 3
	)

	m.auth.menu.SetSize(msg.Width-4, msg.Height-6)

	m.chat.viewport.Width = msg.Width - sidebarWidth - gapWidth - 6
	m.chat.viewport.Height = msg.Height - headerHeight - inputHeight - verticalGap

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
	m.chat.viewport.SetContent(renderEntries(m.chat.entries))
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return "[" + t.Format("15:04") + "] "
}
