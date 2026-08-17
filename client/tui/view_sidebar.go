package tui

import (
	"fmt"
	"strings"
)

type sidebarTab int

const (
	tabUsers sidebarTab = iota
	tabRooms
)

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

			b.WriteString(
				sidebarItemStyle.Render(line),
			)
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString("Tab to switch tabs")

	case tabRooms:
		for i, r := range m.chat.rooms {
			lock := ""
			selected := i == m.chat.selectedRoom
			current := r.Name == m.chat.roomName

			marker := " "

			if selected {
				marker = ">"
			}

			if current {
				marker = "✓"
			}

			if r.HasPassword {
				lock = "🔒"
			}

			line := fmt.Sprintf(
				"%s %s (%d)%s",
				marker,
				r.Name,
				r.Users,
				lock,
			)

			if i == m.chat.selectedRoom {
				b.WriteString(
					sidebarTitleStyle.Render(line),
				)
			} else {
				b.WriteString(
					sidebarItemStyle.Render(line),
				)
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
