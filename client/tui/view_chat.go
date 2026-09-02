package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
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
		sidebarStyle.Width(m.ui.layout.sidebarWidth).Render(m.renderSidebar()),
	)

	inputView := m.chat.input.View()

	if m.chat.awaitingRoomPassword != "" {
		inputView = m.chat.secret.View()
	}

	typingLine := m.renderTypingLine()

	result := headerStyle.Render(header) + "\n\n" + body
	if typingLine != "" {
		result += "\n\n" + typingLine
	}
	result += "\n\n" + inputStyle.Render(inputView)
	return result
}

func (m Model) renderHeader() string {
	title := "#" + m.chat.roomName
	if m.chat.roomTopic != "" {
		title += " - " + m.chat.roomTopic
	}
	if m.connection.addr != "" {
		title += " (" + m.connection.addr + ")"
	}

	status := fmt.Sprintf("Users: %d", len(m.chat.users))

	return headerStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.serverName,
			"  |  ",
			title,
			"  ",
			status,
		),
	)
}

func renderEntries(entries []chatEntry, width int) string {
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
			b.WriteString(wrapAnsi(e.text, width, chatIndent))
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
			b.WriteString(wrapAnsi(e.text, width, tsWidth))
			b.WriteString("\n")

		case "welcome":
			b.WriteString(renderWelcomeEntry(e))

		case "system":
			b.WriteString(ts)
			b.WriteString(systemStyle.Render(wrapAnsi(e.text, width, tsWidth)))
			b.WriteString("\n")

		case "announcement":
			boxed := announcementStyle.Width(width - 2).Render(
				announcementTitleStyle.Render("ANNOUNCEMENT") + "\n\n\n" + e.text,
			)
			b.WriteString(boxed)
			b.WriteString("\n")

		case "topic":
			b.WriteString(ts)
			b.WriteString(systemStyle.Render(
				wrapAnsi(e.nickname+" changed the topic to: "+e.text, width, tsWidth),
			))
			b.WriteString("\n")

		case "error":
			b.WriteString(errorStyle.Render(wrapAnsi("Error: "+e.text, width, tsWidth)))
			b.WriteString("\n")

		case "join":
			b.WriteString(ts)
			b.WriteString(joinLeaveStyle.Render(
				wrapAnsi(e.nickname+" joined the chat", width, tsWidth),
			))
			b.WriteString("\n")

		case "leave":
			b.WriteString(ts)
			b.WriteString(joinLeaveStyle.Render(
				wrapAnsi(e.nickname+" left the chat", width, tsWidth),
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
