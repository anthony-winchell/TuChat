package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"tuchat/client/tui"
)

func main() {
	p := tea.NewProgram(
		tui.New(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()

	if err != nil {
		os.Exit(1)
	}
}
