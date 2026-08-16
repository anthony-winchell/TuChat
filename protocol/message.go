package protocol

import "time"

type Message struct {
	Type string `json:"type"`

	ServerName string `json:"server_name,omitempty"`

	Username string `json:"username,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Password string `json:"password,omitempty"`
	Target   string `json:"target,omitempty"`

	Message string `json:"message,omitempty"`

	Users []UserSummary `json:"users,omitempty"`

	Rooms         []RoomSummary `json:"rooms,omitempty"`
	RoomName      string        `json:"room_name,omitempty"`
	RoomOwner     string        `json:"room_owner,omitempty"`
	RoomAdmins    []string      `json:"room_admins,omitempty"`
	RoomTopic     string        `json:"room_topic,omitempty"`
	RoomUserCount int           `json:"room_user_count,omitempty"`

	Timestamp time.Time `json:"timestamp,omitempty"`
}

type RoomSummary struct {
	Name        string `json:"name"`
	Users       int    `json:"users"`
	HasPassword bool   `json:"has_password"`
}

type UserSummary struct {
	Nickname string `json:"nickname"`
	Admin    bool   `json:"admin"`
	Owner    bool   `json:"owner"`
}
