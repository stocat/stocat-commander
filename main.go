package main

import (
	"fmt"
	"os"

	"stocat-commander/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Root workspace directory where all projects live
	// The user explicitly wants commands to run and clone inside the current directory (stocat-commander)
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Can't get current dir: %v\n", err)
		os.Exit(1)
	}

	workspaceDir := cwd // Use the current directory, NOT the parent

	m := ui.InitialModel(workspaceDir)

	// Create the program
	p := tea.NewProgram(&m, tea.WithAltScreen())

	// Pass the program reference to the model for async updates
	m.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
