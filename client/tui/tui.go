package tui

import (
	"encoding/json"
	"net"
	"time"
	"tuchat/protocol"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

const serverAddr = "localhost:8080"

type Model struct {
	//connection
	conn    net.Conn
	decoder *json.Decoder
	encoder *json.Encoder

	//current screen
	screen screen

	//auth
	authFocus   int
	authError   string
	authStage   authStage
	authChoice  string
	pendingUser string
	authMenu    list.Model

	//chat
	viewport    viewport.Model
	chatLog     []chatEntry
	newMessages int

	width     int
	height    int
	chatInput textinput.Model

	activeSidebar sidebarTab
	users         []protocol.UserSummary
	rooms         []protocol.RoomSummary
	selfNickname  string

	roomName             string
	awaitingRoomPassword string
	roomTopic            string
	selectedRoom         int

	authInput textinput.Model
	err       error
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
	stageAuthenticating
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
	ci.EchoMode = textinput.EchoNormal
	ci.Prompt = "> "
	ci.Focus()

	return Model{
		authInput: ti,
		authMenu:  authMenu,
		viewport:  vp,
		chatInput: ci,
	}
}
