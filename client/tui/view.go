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

		body := lipgloss.JoinHorizontal(
			lipgloss.Top, 
			viewportStyle.Render(m.viewport.View()), 
			sidebarStyle.Render(m.renderSidebar()))

		return headerStyle.Render(header) + "\n\n" + body + "\n\n" + inputStyle.Render(m.chatInput.View())
	}

	var errorLine string
	if m.authError != "" {
		errorLine = errorStyle.Render("Error: " + m.authError + "\n\n")
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
			lock := ""
			if r.HasPassword {
				lock = "🔒"
			}
			b.WriteString(fmt.Sprintf("• %s (%d)%s\n", r.Name, r.Users, lock))
		}
	}

	return b.String()
}

func renderEntries(entries []chatEntry) string {
	var b strings.Builder

	for _, e := range entries {
		ts := timestampStyle.Render(formatTime(e.timestamp))

		switch e.kind {
		case "chat":
			name := nicknameStyle
			if e.self {
				name = selfNickStyle
			}
			b.WriteString(ts + name.Render(e.nickname) + "	" + e.text + "\n")

		case "pm":
			label := fmt.Sprintf("PM %s -> %s", e.nickname, e.target)
			b.WriteString(ts + pmStyle.Render(label) + ": " + e.text + "\n")

		case "system":
			b.WriteString(ts + systemStyle.Render(e.text) + "\n")

		case "error":
			b.WriteString(errorStyle.Render("Error: "+e.text) + "\n")

		case "join":
			b.WriteString(ts + joinLeaveStyle.Render(e.nickname+" joined the chat") + "\n")

		case "leave":
			b.WriteString(ts + joinLeaveStyle.Render(e.nickname+" left the chat") + "\n")
		}
	}

	return b.String()
}