// Package tui is responsible for rendering the user interface
package tui

import (
	"encoding/json"
	"net"
	"time"
	"tuchat/protocol"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	connect    connectState
	connection connection
	auth       authState
	chat       chatState
	ui         uiState

	serverName string

	screen screen
	err    error
}

type connection struct {
	addr string
	conn net.Conn
	dec  *json.Decoder
	enc  *json.Encoder

	state connectionState

	reconnectAttempt int
	creds            reconnectCreds
	preserveHistory  bool
}

type reconnectCreds struct {
	choice         string
	username       string
	password       string
	serverPassword string
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

	input  textarea.Model
	secret textinput.Model

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

	layout layout
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
	connectionReconnecting
)

type connectState struct {
	input textinput.Model
	error string
}

const maxInputRows = 5

type screen int

const (
	screenConnect screen = iota
	screenAuth
	screenChat
)

func (m Model) Init() tea.Cmd {
	if m.connection.addr == "" {
		return m.connect.input.Focus()
	}
	return connectCmd(m.connection.addr)
}

func New(addr string) Model {
	connectInput := newConnectInput()

	screen := screenAuth
	if addr == "" {
		connectInput.SetValue("localhost:8080")
		screen = screenConnect
	}

	authInput := newAuthInput()
	authMenu := newAuthMenu()
	chatViewport := newChatViewport()
	chatInput := newChatInput()
	passwordInput := newPasswordInput()

	return Model{
		screen: screen,

		connect: connectState{
			input: connectInput,
		},

		connection: connection{
			addr:  addr,
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
			secret:   passwordInput,
		},
	}
}

func newConnectInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "host:port"
	ti.Prompt = "🔗 "
	ti.Focus()

	return ti
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

func newChatInput() textarea.Model {
	ta := textarea.New()

	ta.Placeholder = "message or /command"
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.SetHeight(1)
	ta.CharLimit = 2000

	ta.KeyMap.InsertNewline.SetKeys("alt+enter")

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	return ta
}

func newPasswordInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "password"
	ti.Prompt = "🔒 "
	ti.EchoMode = textinput.EchoPassword
	return ti
}
