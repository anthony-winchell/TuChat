package tui

import (
	"encoding/json"
	"github.com/charmbracelet/bubbletea"
	"net"
	"tuchat/protocol"
)

const serverAddr = "localhost:8080"

func connectCmd() tea.Cmd {
	return func() tea.Msg {
		conn, err := net.Dial("tcp", serverAddr)
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
