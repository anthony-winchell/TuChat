package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorAccent       = lipgloss.Color("#4078c0")
	colorMuted        = lipgloss.Color("#546377")
	colorError        = lipgloss.Color("#EF4444")
	colorAnnouncement = lipgloss.Color("#EAB308")
	colorNickname     = lipgloss.Color("#6e5494")
	colorPM           = lipgloss.Color("#72f478")

	timestampStyle      = lipgloss.NewStyle().Foreground(colorMuted)
	nicknameStyle       = lipgloss.NewStyle().Bold(true).Foreground(colorNickname)
	selfNickStyle       = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	systemStyle         = lipgloss.NewStyle().Foreground(colorMuted).Italic(true)
	pmStyle             = lipgloss.NewStyle().Foreground(colorPM).Italic(true)
	joinLeaveStyle      = lipgloss.NewStyle().Foreground(colorMuted)
	nicknameColumnStyle = lipgloss.NewStyle().Width(8)
	statusStyle         = lipgloss.NewStyle().
				Foreground(colorMuted)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(colorAccent).
			Padding(0, 1)

	viewportStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(0, 1)

	sidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)

	sidebarTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent)

	sidebarInactiveStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

	sidebarItemStyle = lipgloss.NewStyle().
				PaddingLeft(1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			BorderForeground(colorAccent)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	newMessagesStyle = lipgloss.NewStyle().
				Align(lipgloss.Center).
				Foreground(colorAccent).
				Italic(true)

	authTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)

	authInputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Width(35)

	announcementStyle = lipgloss.NewStyle().
				Bold(true).
				Border(lipgloss.DoubleBorder()).
				Padding(3, 5).
				BorderForeground(colorAnnouncement)

	announcementTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAnnouncement)

	typingStyle = lipgloss.NewStyle().Foreground(colorMuted).Italic(true)
)
