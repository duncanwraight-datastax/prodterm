package ui

import (
	"fmt"
	"strings"
	"terminal-claude/config"
	"terminal-claude/handlers"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF5F87")).
		Padding(0, 1)

	promptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87FF5F")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#87FF5F")).
		Padding(0, 1).
		Bold(true)

	responseStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5F87FF"))
	
	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5F5F")).
		Bold(true)

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5F5F5F"))
        
    spinnerStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("205"))
)

type errMsg error

// Model represents the UI state
type Model struct {
	viewport    viewport.Model
	textInput   textinput.Model
	spinner     spinner.Model
	handler     *handlers.Handler
	history     []string
	response    string
	err         error
	loading     bool
	windowWidth int
    windowHeight int
}

// InitialModel creates and initializes the UI model
func InitialModel(cfg config.Config) Model {
	ti := textinput.New()
	ti.Placeholder = "Type your request..."
	ti.Focus()
	ti.Width = 80
	
	vp := viewport.New(80, 20)
	vp.SetContent(welcomeMessage())
	
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle
	
	return Model{
		textInput: ti,
		viewport:  vp,
		spinner:   s,
		handler:   handlers.NewHandler(cfg),
		history:   []string{welcomeMessage()},
        windowWidth: 80,
        windowHeight: 24,
	}
}

// welcomeMessage returns the initial welcome text
// wrapText wraps the text to fit within width characters per line
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	
	var result strings.Builder
	lines := strings.Split(text, "\n")
	
	for i, line := range lines {
		if len(line) <= width {
			result.WriteString(line)
		} else {
			// Process the line by breaking it at word boundaries
			words := strings.Fields(line)
			lineLength := 0
			
			for j, word := range words {
				if lineLength+len(word) > width && lineLength > 0 {
					// Start a new line
					result.WriteString("\n")
					lineLength = 0
				}
				
				if j > 0 && lineLength > 0 {
					result.WriteString(" ")
					lineLength++
				}
				
				result.WriteString(word)
				lineLength += len(word)
			}
		}
		
		// Add newline unless it's the last line
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

func welcomeMessage() string {
	return `
 ______   ______     ______     _____     ______   ______     ______     __    __    
/\  == \ /\  == \   /\  __ \   /\  __-.  /\__  _\ /\  ___\   /\  == \   /\ "-./  \   
\ \  _-/ \ \  __<   \ \ \/\ \  \ \ \/\ \ \/_/\ \/ \ \  __\   \ \  __<   \ \ \-./\ \  
 \ \_\    \ \_\ \_\  \ \_____\  \ \____-    \ \_\  \ \_____\  \ \_\ \_\  \ \_\ \ \_\ 
  \/_/     \/_/ /_/   \/_____/   \/____/     \/_/   \/_____/   \/_/ /_/   \/_/  \/_/ 
                                                                                    
` +
		"type your requests or commands. type 'exit' to quit.\n\n" +
		"example commands:\n" +
		"- summarise my unread e-mails\n" +
		"- what's on this webpage? bbc.co.uk\n" +
		"- list slack channels\n" +
		"- summarise slack channel #general\n" +
		"- tell me about golang\n"
}

// Init initializes the UI
func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

// Update handles UI events
type responseMsg struct {
	response string
}

// sendRequest sends the request to the handler
func (m Model) sendRequest(input string) tea.Cmd {
	// Make a local copy of the input to ensure it doesn't change
	commandToProcess := input
	
	return func() tea.Msg {
		response, err := m.handler.ProcessCommand(commandToProcess)
		if err != nil {
			return errMsg(err)
		}
		return responseMsg{response: response}
	}
}

// Update handles UI events and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		spCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.textInput.Value() == "" {
				return m, nil
			}
			if m.textInput.Value() == "exit" {
				return m, tea.Quit
			}
			
			// Prepare for request
			userInput := m.textInput.Value()
			
			m.history = append(m.history, "> "+userInput)
			m.loading = true
			m.textInput.Reset()
			
			// Create a clean spinner
			s := spinner.New()
			s.Spinner = spinner.Dot
			s.Style = spinnerStyle
			m.spinner = s
			
			// Update viewport with the new input
			content := strings.Join(m.history, "\n")
			m.viewport.SetContent(content)
			
			// Scroll to bottom
			m.viewport.GotoBottom()
			
			return m, tea.Batch(m.sendRequest(userInput), m.spinner.Tick)
			
		case tea.KeyCtrlL:
			m.history = []string{welcomeMessage()}
			m.viewport.SetContent(strings.Join(m.history, "\n"))
			return m, nil
		}

	case responseMsg:
		m.loading = false
		// Completely reinitialize the spinner
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = spinnerStyle
		m.spinner = s
		
		// Make sure the response is wrapped to fit the width
		maxWidth := m.windowWidth - 4 // Account for margins
		if maxWidth <= 0 {
			maxWidth = 76 // Default width
		}
		
		wrappedResponse := wrapText(msg.response, maxWidth)
		m.history = append(m.history, responseStyle.Render(wrappedResponse))
		m.viewport.SetContent(strings.Join(m.history, "\n"))
		m.viewport.GotoBottom()
		return m, nil

	case errMsg:
		m.loading = false
		// Completely reinitialize the spinner
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = spinnerStyle
		m.spinner = s
		
		m.err = msg
		
		// Make sure the error message is wrapped to fit the width
		maxWidth := m.windowWidth - 4 // Account for margins
		if maxWidth <= 0 {
			maxWidth = 76 // Default width
		}
		
		errText := fmt.Sprintf("Error: %v", msg)
		wrappedError := wrapText(errText, maxWidth)
		m.history = append(m.history, errorStyle.Render(wrappedError))
		m.viewport.SetContent(strings.Join(m.history, "\n"))
		m.viewport.GotoBottom()
		return m, nil

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
        m.windowHeight = msg.Height
		headerHeight := 1
		footerHeight := 3
		verticalMargins := headerHeight + footerHeight
		
        // Adjust viewport dimensions to match terminal size
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMargins
		
		// Also adjust textinput width to match terminal width
		m.textInput.Width = msg.Width - 4 // Account for borders
		
		// Update content
		if m.viewport.Height >= 0 {
			m.viewport.SetContent(strings.Join(m.history, "\n"))
			m.viewport.GotoBottom()
		}
        
        return m, nil

	case spinner.TickMsg:
		if m.loading {
			// Only tick once
			m.spinner, spCmd = m.spinner.Update(msg)
			// Return the viewport alone without spinner commands to prevent multiple renders
			return m, spCmd
		}
	}

	m.textInput, tiCmd = m.textInput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd, spCmd)
}

// View renders the UI
func (m Model) View() string {
	var footerContent string
	
	// Calculate available width
	availWidth := m.windowWidth
	if availWidth <= 0 {
		availWidth = 80 // Default width
	}
	
	// Ensure the text input respects the available width
	m.textInput.Width = availWidth - 4 // Account for border and padding
	
	// Input field or spinner
	if m.loading {
		// Display a single spinner without duplication
		footerContent = m.spinner.View() + " Processing..."
	} else {
		// Box-style prompt with width constraint
		boxStyle := promptStyle.Copy().Width(availWidth - 2) // Apply width constraint to the box
		footerContent = boxStyle.Render(m.textInput.View())
	}
	
	// Help text
	helpText := helpStyle.Render("Ctrl+C to quit, Ctrl+L to clear")
	
	// Ensure the terminal width constraint is respected by all content
    maxWidth := m.windowWidth
    if maxWidth <= 0 {
        maxWidth = 80 // Default width
    }
    
    // Set maximum width for viewport
    m.viewport.Width = maxWidth
    
	// When processing, don't show help text to avoid duplication
	if m.loading {
		return fmt.Sprintf("%s\n\n%s", m.viewport.View(), footerContent)
	} else {
		return fmt.Sprintf("%s\n\n%s\n%s", m.viewport.View(), footerContent, helpText)
	}
}