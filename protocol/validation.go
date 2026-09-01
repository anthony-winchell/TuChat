package protocol

import (
	"encoding/json"
	"errors"
	"strconv"
)

const (
	MaxMessageSize = 1 << 16 // 64 KiB
	MaxMessageLen  = 4096
	MaxNicknameLen = 13
	MaxRoomNameLen = 32
	MaxTargetLen   = MaxNicknameLen
)

func ValidateField(name, value string, maxLen int) error {
	if len(value) > maxLen {
		return errors.New(name + " exceeds max length of " + strconv.Itoa(maxLen))
	}
	return nil
}

func ValidateMessage(msg *Message) error {
	if err := ValidateField("message", msg.Message, MaxMessageLen); err != nil {
		return err
	}

	if err := ValidateField("nickname", msg.Nickname, MaxNicknameLen); err != nil {
		return err
	}

	if err := ValidateField("room_name", msg.RoomName, MaxRoomNameLen); err != nil {
		return err
	}

	if err := ValidateField("target", msg.Target, MaxTargetLen); err != nil {
		return err
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if len(b) > MaxMessageSize {
		return errors.New("message exceeds maximum size")
	}

	return nil
}
