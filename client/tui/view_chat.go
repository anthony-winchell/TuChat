package tui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"time"
)

func (m Model) renderChat() string {
	header := m.renderHeader()

	chatPane := viewportStyle.Render(
		m.chat.viewport.View(),
	)

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
			b.WriteString(systemStyle.Render("Joined #"+e.text))
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

func renderWelcomeEntry(e chatEntry) string {
	lines := strings.Split(e.text, "\n")

	var b strings.Builder

	for i, line := range lines {
		if i == 0 {
			b.WriteString(
				timestampStyle.Render(formatTime(e.timestamp)),
			)
			b.WriteString(
				systemStyle.Render(line),
			)
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

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return "[" + t.Format("15:04") + "] "
}
