package protocol

type Message struct {
	Type string `json:"type"`

	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Target string `json:"target,omitempty"`

	Message string `json:"message,omitempty"`

	Users []string `json:"users,omitempty"`
	
}