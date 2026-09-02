package tui

import (
	"encoding/json"
	"net"
	"tuchat/protocol"
)

type connectedMsg struct {
	conn    net.Conn
	decoder *json.Decoder
	encoder *json.Encoder
}

type connErrMsg struct {
	err error
}

type serverMsg protocol.Message

type sendErrMsg struct {
	err error
}

type reconnectMsg struct {
	attempt int
}
