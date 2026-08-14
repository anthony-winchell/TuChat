package tui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type sidebarTab int

const (
	tabUsers sidebarTab = iota
	tabRooms
)

func (m Model) View() string {
	if m.err != nil {
		return "connection error: " + m.err.Error()
	}

	if m.screen == screenChat {
		header := fmt.Sprintf("#%s - %s", m.roomName, m.roomTopic)
		if m.roomTopic == "" {
			header = "#" + m.roomName
		}

		body := lipgloss.JoinHorizontal(lipgloss.Top, m.viewport.View(), m.renderSidebar())

		return header + "\n\n" + body + "\n\n" + m.chatInput.View()
	}

	var errorLine string
	if m.authError != "" {
		errorLine = "Error: " + m.authError + "\n\n"
	}

	if m.authStage == stageMenu {
		return errorLine + m.authMenu.View()
	}

	label := m.authChoice + " - username:"
	if m.authStage == stagePassword {
		label = m.authChoice + " - password:"
	}

	return errorLine + label + "\n\n" + m.input.View()
}

func (m Model) renderSidebar() string {
	var b strings.Builder

	switch m.activeSidebar {
	case tabUsers:
		b.WriteString("USERS\n")
		for _, user := range m.users {
			line := "• " + user.Nickname
			if user.Owner {
				line += " (owner)"
			} else if user.Admin {
				line += " (admin)"
			}
			b.WriteString(line + "\n")
		}
	case tabRooms:
		b.WriteString("ROOMS\n")
		for _, r := range m.rooms {
			b.WriteString(fmt.Sprintf("• %s (%d)\n", r.Name, r.Users))
		}
	}

	return b.String()
}