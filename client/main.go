package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"tuchat/client/tui"
)

func main() {
	p := tea.NewProgram(tui.New(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}