package tui

import (
	"encoding/json"
	"net"
	"time"
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

type typingTickMsg time.Time

type sendErrMsg struct {
	err error
}

type reconnectMsg struct {
	attempt int
}
