package tui

import (
	"encoding/json"
	"time"
	"tuchat/protocol"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

const serverAddr = "localhost:8080"

type Model struct {
	screen    screen
	authStage authStage

	authChoice  string
	authError   string
	pendingUser string

	authMenu list.Model

	chatLog   []chatEntry
	viewport  viewport.Model
	width     int
	height    int
	chatInput textinput.Model

	activeSidebar sidebarTab
	users         []protocol.UserSummary
	rooms         []protocol.RoomSummary
	selfNickname string

	roomName  string
	roomTopic string

	decoder *json.Decoder
	encoder *json.Encoder
	input   textinput.Model
	err     error
}

type chatEntry struct {
	kind      string
	timestamp time.Time
	nickname  string
	target    string
	text      string
	self      bool
}

type authStage int

const (
	stageMenu authStage = iota
	stageUsername
	stagePassword
)

type authOption string

func (a authOption) Title() string {
	return string(a)
}

func (a authOption) Description() string {
	return ""
}

func (a authOption) FilterValue() string {
	return string(a)
}

func (m Model) Init() tea.Cmd {
	return connectCmd()
}

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "username"
	ti.Focus()

	items := []list.Item{
		authOption("Login"),
		authOption("Register"),
	}

	authMenu := list.New(items, list.NewDefaultDelegate(), 0, 0)
	authMenu.Title = "TuChat"
	authMenu.SetSize(30, 10)
	authMenu.SetShowStatusBar(false)
	authMenu.SetFilteringEnabled(false)

	vp := viewport.New(0, 0)
	vp.Width = 60
	vp.Height = 15

	ci := textinput.New()
	ci.Placeholder = "message or /command"
	ci.Focus()

	return Model{
		input:     ti,
		authMenu:  authMenu,
		viewport:  vp,
		chatInput: ci,
	}
}
