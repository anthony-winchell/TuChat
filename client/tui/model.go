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

type Model struct {
	connection connection
	auth       authState
	chat       chatState
	ui         uiState

	serverName string

	screen screen
	err    error
}

type connection struct {
	conn  net.Conn
	dec   *json.Decoder
	enc   *json.Encoder
	state connectionState
}

type authState struct {
	error       string
	stage       authStage
	choice      string
	pendingUser string
	input       textinput.Model
	menu        list.Model
}

type chatState struct {
	viewport    viewport.Model
	entries     []chatEntry
	newMessages int

	input textinput.Model

	users []protocol.UserSummary
	rooms []protocol.RoomSummary

	selfNickname string

	roomName             string
	roomTopic            string
	selectedRoom         int
	awaitingRoomPassword string

	activeSidebar sidebarTab
}

type uiState struct {
	width  int
	height int
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
	stageServerPassword
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

type connectionState int

const (
	connectionConnecting connectionState = iota
	connectionConnected
	connectionDisconnected
)

type screen int

const (
	screenAuth screen = iota
	screenChat
)

func (m Model) Init() tea.Cmd {
	return connectCmd()
}

func New() Model {
	authInput := newAuthInput()
	authMenu := newAuthMenu()
	chatViewport := newChatViewport()
	chatInput := newChatInput()

	return Model{
		screen: screenAuth,

		connection: connection{
			state: connectionConnecting,
		},

		auth: authState{
			stage: stageMenu,
			input: authInput,
			menu:  authMenu,
		},

		chat: chatState{
			viewport: chatViewport,
			input:    chatInput,
		},
	}
}

func newAuthInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "username"

	return ti
}

func newAuthMenu() list.Model {
	items := []list.Item{
		authOption("Login"),
		authOption("Register"),
	}

	menu := list.New(
		items,
		list.NewDefaultDelegate(),
		0,
		0,
	)

	menu.Title = "TuChat"
	menu.SetSize(30, 10)
	menu.SetShowStatusBar(false)
	menu.SetFilteringEnabled(false)

	return menu
}

func newChatViewport() viewport.Model {
	vp := viewport.New(0, 0)
	vp.Width = 60
	vp.Height = 15

	return vp
}

func newChatInput() textinput.Model {
	ti := textinput.New()

	ti.Placeholder = "message or /command"
	ti.EchoMode = textinput.EchoNormal
	ti.Prompt = "> "
	ti.Focus()

	return ti
}
