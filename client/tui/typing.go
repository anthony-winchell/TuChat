package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"tuchat/protocol"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	typingDebounce = 2 * time.Second
	typingExpiry   = 3 * time.Second
)

func typingTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return typingTickMsg(t)
	})
}

func (m *Model) typingCmd() tea.Cmd {
	value := strings.TrimSpace(m.chat.input.Value())

	if value == "" {
		return sendCmd(m.connection.enc, protocol.Message{
			Type: "typing_stop",
		})
	}

	if strings.HasPrefix(value, "/") {
		return nil
	}

	if m.chat.lastTypingSent.IsZero() || time.Since(m.chat.lastTypingSent) >= typingDebounce {
		m.chat.lastTypingSent = time.Now()
		return sendCmd(m.connection.enc, protocol.Message{
			Type: "typing_start",
		})
	}

	return nil
}

func (m *Model) addTypingUser(username string) {
	m.chat.typingUsers[username] = time.Now()
}

func (m *Model) removeTypingUser(username string) {
	delete(m.chat.typingUsers, username)
}

func (m *Model) pruneTypingUsers() {
	now := time.Now()
	for username, lastTyping := range m.chat.typingUsers {
		if now.Sub(lastTyping) >= typingExpiry {
			delete(m.chat.typingUsers, username)
		}
	}
}

func (m Model) typingNicknames() string {
	if len(m.chat.typingUsers) == 0 {
		return ""
	}

	names := make([]string, 0, len(m.chat.typingUsers))
	for username := range m.chat.typingUsers {
		names = append(names, username)
	}
	sort.Strings(names)

	const maxShown = 3
	if len(names) > maxShown {
		names = append(names[:maxShown], fmt.Sprintf("+%d more", len(names)-maxShown))
	}

	verb := "is typing…"
	if len(m.chat.typingUsers) > 1 {
		verb = "are typing…"
	}

	return strings.Join(names, ", ") + " " + verb
}

func (m Model) renderTypingLine() string {
	line := m.typingNicknames()
	if line == "" {
		return ""
	}
	return typingStyle.Render(line)
}
