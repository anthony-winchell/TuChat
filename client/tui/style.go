package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorAccent   = lipgloss.Color("#7C3AED")
	colorMuted    = lipgloss.Color("#6B7280")
	colorError    = lipgloss.Color("#EF4444")
	colorNickname = lipgloss.Color("#22D3EE")
	colorPM       = lipgloss.Color("#F472B6")

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
			Width(24).
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
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)
)
