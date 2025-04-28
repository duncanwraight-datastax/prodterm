package main

import (
	"fmt"
	"log"
	"os"

	"terminal-claude/config"
	"terminal-claude/mcp"
	"terminal-claude/providers/gmail"
	"terminal-claude/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize providers
	initializeProviders()

	// Start the UI
	if err := tea.NewProgram(ui.InitialModel(cfg), tea.WithAltScreen()).Start(); err != nil {
		fmt.Printf("Error starting application: %v\n", err)
		os.Exit(1)
	}
}

// initializeProviders registers all MCP providers
func initializeProviders() {
	// Initialize Gmail provider
	gmailProvider, err := gmail.New()
	if err != nil {
		log.Printf("Warning: Failed to initialize Gmail provider: %v", err)
	} else {
		mcp.Register(gmailProvider)
		log.Println("Registered Gmail provider")
	}
	
	// Add other providers here as needed
}
