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
	if m.err != nil {
		return "connection error: " + m.err.Error()
	}

	if m.screen == screenChat {
		header := m.renderHeader()

		if m.roomTopic == "" {
			header = "#" + m.roomName
		}

		body := lipgloss.JoinHorizontal(
			lipgloss.Top,
			viewportStyle.Render(m.viewport.View()),
			sidebarStyle.Render(m.renderSidebar()))

		return headerStyle.Render(header) +
			"\n\n" +
			body +
			"\n\n" +
			inputStyle.Render(m.chatInput.View())
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

	usersTitle := "USERS"
	roomsTitle := "ROOMS"

	if m.activeSidebar == tabUsers {
		usersTitle = sidebarTitleStyle.Render(usersTitle)
		roomsTitle = sidebarInactiveStyle.Render(roomsTitle)
	} else {
		usersTitle = sidebarInactiveStyle.Render(usersTitle)
		roomsTitle = sidebarTitleStyle.Render(roomsTitle)
	}

	b.WriteString(usersTitle + "  " + roomsTitle + "\n\n")

	switch m.activeSidebar {
	case tabUsers:
		for _, user := range m.users {
			line := "• " + user.Nickname
			if user.Owner {
				line += " (owner)"
			} else if user.Admin {
				line += " (admin)"
			}
			b.WriteString(sidebarItemStyle.Render(line) + "\n")
		}
	case tabRooms:
		for _, r := range m.rooms {
			lock := ""
			if r.HasPassword {
				lock = "🔒"
			}
			b.WriteString(sidebarItemStyle.Render(fmt.Sprintf("• %s (%d)%s", r.Name, r.Users, lock)) + "\n")
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

			nickname := nicknameColumnStyle.Render(
				name.Render(e.nickname),
			)

			b.WriteString(
				ts + nickname + e.text + "\n",
			)

		case "pm":
			label := fmt.Sprintf(
				"PM %s -> %s",
				e.nickname,
				e.target,
			)

			b.WriteString(
				ts +
					pmStyle.Render(label) +
					": " +
					e.text +
					"\n",
			)

		case "welcome":
			b.WriteString(renderWelcomeEntry(e))

		case "system":
			b.WriteString(
				ts +
					systemStyle.Render(e.text) +
					"\n",
			)

		case "room_joined":
			b.WriteString(
				ts +
					systemStyle.Render("Joined #"+e.text) +
					"\n",
			)

		case "topic":
			b.WriteString(
				ts +
					systemStyle.Render(
						e.nickname+" changed the topic to: "+e.text,
					) +
					"\n",
			)

		case "error":
			b.WriteString(
				errorStyle.Render("Error: "+e.text) +
					"\n",
			)

		case "join":
			b.WriteString(
				ts +
					joinLeaveStyle.Render(
						e.nickname+" joined the chat",
					) +
					"\n",
			)

		case "leave":
			b.WriteString(
				ts +
					joinLeaveStyle.Render(
						e.nickname+" left the chat",
					) +
					"\n",
			)
		}
	}

	return b.String()
}

func (m Model) renderHeader() string {
	title := "#" + m.roomName

	if m.roomTopic != "" {
		title += " - " + m.roomTopic
	}

	status := fmt.Sprintf(
		"Users: %d",
		len(m.users),
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
			b.WriteString(
				timestampStyle.Render(formatTime(e.timestamp)) +
					systemStyle.Render(line),
			)
		} else {
			b.WriteString(
				"\n" +
					systemStyle.Render(line),
			)
		}
	}

	b.WriteString("\n")

	return b.String()
}
