package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type sidebarTab int

const (
	tabUsers sidebarTab = iota
	tabRooms
)

func (m Model) View() string {
	if m.connection.state == connectionDisconnected {
		return m.renderDisconnected()
	}

	if m.screen == screenChat {
		header := m.renderHeader()

		chatPane := viewportStyle.Render(m.chat.viewport.View())

		if newMessages := m.renderNewMessages(); newMessages != "" {
			chatPane = lipgloss.JoinVertical(
				lipgloss.Left,
				chatPane,
				newMessages,
			)
		}

		body := lipgloss.JoinHorizontal(
			lipgloss.Top,
			chatPane,
			sidebarStyle.Render(m.renderSidebar()),
		)

		return headerStyle.Render(header) +
			"\n\n" +
			body +
			"\n\n" +
			inputStyle.Render(m.chat.input.View())
	}

	var errorLine string

	if m.auth.error != "" {
		errorLine = errorStyle.Render("Error: " + m.auth.error + "\n\n")
	}

	if m.auth.stage == stageMenu {
		return errorLine + m.auth.menu.View()
	}

	if m.auth.stage == stageAuthenticating {
		return errorLine + "Authenticating...\n\n"
	}

	label := m.auth.choice + " - username:"

	if m.auth.stage == stagePassword {
		label = m.auth.choice + " - password:"
	}

	return errorLine + label + "\n\n" + m.auth.input.View()
}

func (m Model) renderSidebar() string {
	var b strings.Builder

	usersTitle := "USERS"
	roomsTitle := "ROOMS"

	if m.chat.activeSidebar == tabUsers {
		usersTitle = sidebarTitleStyle.Render(usersTitle)
		roomsTitle = sidebarInactiveStyle.Render(roomsTitle)
	} else {
		usersTitle = sidebarInactiveStyle.Render(usersTitle)
		roomsTitle = sidebarTitleStyle.Render(roomsTitle)
	}

	b.WriteString(usersTitle)
	b.WriteString("  ")
	b.WriteString(roomsTitle)
	b.WriteString("\n\n")

	switch m.chat.activeSidebar {
	case tabUsers:
		for _, user := range m.chat.users {
			line := "• " + user.Nickname
			if user.Owner {
				line += " (owner)"
			} else if user.Admin {
				line += " (admin)"
			}
			b.WriteString(sidebarItemStyle.Render(line))
			b.WriteString("\n\n")
			b.WriteString("Tab to switch tabs")
		}
	case tabRooms:
		for i, r := range m.chat.rooms {
			lock := ""
			if r.HasPassword {
				lock = "🔒"
			}

			line := fmt.Sprintf("%s (%d)%s", r.Name, r.Users, lock)

			if i == m.chat.selectedRoom {
				line = "> " + line
				b.WriteString(sidebarTitleStyle.Render(line))
			} else {
				line = "  " + line
				b.WriteString(sidebarItemStyle.Render(line))
			}

			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString("Ctrl+J to join")
		b.WriteString("\n")
		b.WriteString("↑/↓ Select")
		b.WriteString("\n")
		b.WriteString("Tab to switch tabs")
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

			nickname := nicknameColumnStyle.Render(
				name.Render(e.nickname),
			)

			b.WriteString(ts)
			b.WriteString(nickname)
			b.WriteString(e.text)
			b.WriteString("\n")

		case "pm":
			label := fmt.Sprintf(
				"PM %s -> %s",
				e.nickname,
				e.target,
			)

			b.WriteString(ts)
			b.WriteString(pmStyle.Render(label))
			b.WriteString(": ")
			b.WriteString(e.text)
			b.WriteString("\n")

		case "welcome":
			b.WriteString(renderWelcomeEntry(e))

		case "system":
			b.WriteString(ts)
			b.WriteString(systemStyle.Render(e.text))
			b.WriteString("\n")

		case "room_joined":
			b.WriteString(ts)
			b.WriteString(systemStyle.Render("Joined #" + e.text))
			b.WriteString("\n")

		case "topic":
			b.WriteString(ts)
			b.WriteString(systemStyle.Render(
				e.nickname + " changed the topic to: " + e.text,
			))
			b.WriteString("\n")

		case "error":
			b.WriteString(errorStyle.Render("Error: " + e.text))
			b.WriteString("\n")

		case "join":
			b.WriteString(ts)
			b.WriteString(joinLeaveStyle.Render(
				e.nickname + " joined the chat",
			))
			b.WriteString("\n")

		case "leave":
			b.WriteString(ts)
			b.WriteString(joinLeaveStyle.Render(
				e.nickname + " left the chat",
			))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m Model) renderHeader() string {
	title := "#" + m.chat.roomName

	if m.chat.roomTopic != "" {
		title += " - " + m.chat.roomTopic
	}

	status := fmt.Sprintf(
		"Users: %d",
		len(m.chat.users),
	)

	return headerStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			title,
			"  ",
			status,
		),
	)
}

func renderWelcomeEntry(e chatEntry) string {
	lines := strings.Split(e.text, "\n")

	var b strings.Builder

	for i, line := range lines {
		if i == 0 {
			b.WriteString(timestampStyle.Render(formatTime(e.timestamp)))
			b.WriteString(systemStyle.Render(line))
		} else {
			b.WriteString("\n")
			b.WriteString(systemStyle.Render(line))
		}
	}

	b.WriteString("\n")

	return b.String()
}

func (m Model) renderNewMessages() string {
	if m.chat.newMessages == 0 || m.chat.viewport.AtBottom() {
		return ""
	}

	return newMessagesStyle.Render(fmt.Sprintf("↓ %d new messages", m.chat.newMessages))
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
