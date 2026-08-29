package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.ui.width == 0 || m.ui.height == 0 {
		return "\n connecting..."
	}
	if m.connection.state == connectionDisconnected {
		return m.renderDisconnected()
	}

	if m.screen == screenConnect {
		return m.renderConnect()
	}

	if m.screen == screenAuth {
		return m.renderAuth()
	}

	return m.renderChat()
}

func (m Model) renderDisconnected() string {
	var message string

	if m.err != nil {
		message = m.err.Error()
	}

	return lipgloss.NewStyle().
		Padding(2, 4).
		Render(
			"Disconnected from server\n\n" +
				message +
				"\n\n" +
				"Press q to quit",
		)
}
