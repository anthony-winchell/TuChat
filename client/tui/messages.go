package tui

import (
	"tuchat/protocol"
	"encoding/json"
	tea "github.com/charmbracelet/bubbletea"
	"net"
)

type connectedMsg struct {
	decoder *json.Decoder
	encoder *json.Encoder
}

type connErrMsg struct {
	err error
}

type serverMsg protocol.Message

type sentMsg struct{}

type sendErrMsg struct {
	err error
}

func connectCmd() tea.Cmd {
	return func() tea.Msg {
		conn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			return connErrMsg{err: err}
		}
		return connectedMsg{
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
		return sentMsg{}
	}
}