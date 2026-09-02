package tui

import (
	"encoding/json"
	"net"
	"time"

	"tuchat/protocol"

	tea "github.com/charmbracelet/bubbletea"
)

func connectCmd(addr string) tea.Cmd {
	return func() tea.Msg {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return connErrMsg{err: err}
		}
		return connectedMsg{
			conn:    conn,
			decoder: json.NewDecoder(conn),
			encoder: json.NewEncoder(conn),
		}
	}
}

func listenCmd(conn net.Conn, decoder *json.Decoder) tea.Cmd {
	return func() tea.Msg {
		conn.SetReadDeadline(time.Now().Add(150 * time.Second))

		var msg protocol.Message
		if err := decoder.Decode(&msg); err != nil {
			return connErrMsg{err: err}
		}
		return serverMsg(msg)
	}
}

func sendCmd(encoder *json.Encoder, msg protocol.Message) tea.Cmd {
	return func() tea.Msg {
		if err := encoder.Encode(msg); err != nil {
			return sendErrMsg{err: err}
		}
		return nil
	}
}

func reconnectDelay(attempt int) time.Duration {
	shift := min(attempt-1, 5)
	if shift < 0 {
		shift = 0
	}
	delay := time.Duration(1<<shift) * time.Second
	if delay > 30*time.Second {
		delay = 30 * time.Second
	}
	return delay
}

func reconnectCmd(attempt int) tea.Cmd {
	return tea.Tick(reconnectDelay(attempt), func(time.Time) tea.Msg {
		return reconnectMsg{attempt: attempt}
	})
}
