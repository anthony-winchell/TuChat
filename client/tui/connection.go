package tui

import (
	"encoding/json"
	"net"
	"os"

	"tuchat/protocol"

	tea "github.com/charmbracelet/bubbletea"
)

func serverAddress() string {
	if addr := os.Getenv("TUCHAT_ADDR"); addr != "" {
		return addr
	}
	return "localhost:8080"
}

func connectCmd() tea.Cmd {
	return func() tea.Msg {
		conn, err := net.Dial("tcp", serverAddress())
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

func listenCmd(decoder *json.Decoder) tea.Cmd {
	return func() tea.Msg {
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
