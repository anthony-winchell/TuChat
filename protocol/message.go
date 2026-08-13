package protocol

import "time"

type Message struct {
	Type string `json:"type"`

	Username string `json:"username,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Password string `json:"password,omitempty"`
	Target   string `json:"target,omitempty"`

	Message string `json:"message,omitempty"`

	Users []string `json:"users,omitempty"`

	Rooms []RoomSummary `json:"rooms,omitempty"`
	RoomOwner string `json:"room_owner,omitempty"`
	RoomAdmins []string `json:"room_admins,omitempty"`
	RoomTopic string `json:"room_topic,omitempty"`
	RoomUserCount int `json:"room_user_count,omitempty"`

	Timestamp time.Time `json:"timestamp,omitempty"`
}


type RoomSummary struct {
	Name string `json:"name"`
	Users int `json:"users"`
}