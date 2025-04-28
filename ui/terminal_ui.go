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
		Bold(true)

	responseStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5F87FF"))
	
	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5F5F")).
		Bold(true)

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5F5F5F"))
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
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	return Model{
		textInput: ti,
		viewport:  vp,
		spinner:   s,
		handler:   handlers.NewHandler(cfg),
		history:   []string{},
	}
}

// welcomeMessage returns the initial welcome text
func welcomeMessage() string {
	return "Terminal Claude Assistant\n" +
		"------------------------\n" +
		"Type your requests or commands. Type 'exit' to quit.\n\n" +
		"Example commands:\n" +
		"- summarise my unread e-mails\n" +
		"- what's on this webpage? bbc.co.uk\n" +
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
func (m Model) sendRequest() tea.Cmd {
	input := m.textInput.Value()
	
	return func() tea.Msg {
		response, err := m.handler.ProcessCommand(input)
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
			
			// Update viewport with the new input
			content := strings.Join(m.history, "\n\n")
			m.viewport.SetContent(content)
			
			// Scroll to bottom
			m.viewport.GotoBottom()
			
			return m, tea.Batch(m.sendRequest(), m.spinner.Tick)
			
		case tea.KeyCtrlL:
			m.history = []string{welcomeMessage()}
			m.viewport.SetContent(strings.Join(m.history, "\n\n"))
			return m, nil
		}

	case responseMsg:
		m.loading = false
		m.history = append(m.history, responseStyle.Render(msg.response))
		m.viewport.SetContent(strings.Join(m.history, "\n\n"))
		m.viewport.GotoBottom()
		return m, nil

	case errMsg:
		m.loading = false
		m.err = msg
		m.history = append(m.history, errorStyle.Render(fmt.Sprintf("Error: %v", msg)))
		m.viewport.SetContent(strings.Join(m.history, "\n\n"))
		m.viewport.GotoBottom()
		return m, nil

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		headerHeight := 1
		footerHeight := 3
		verticalMargins := headerHeight + footerHeight
		
		if !m.loading {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMargins
		}
		
		m.textInput.Width = msg.Width - 2
		
		if m.viewport.Height >= 0 {
			m.viewport.SetContent(strings.Join(m.history, "\n\n"))
			m.viewport.GotoBottom()
		}

	case spinner.TickMsg:
		if m.loading {
			m.spinner, spCmd = m.spinner.Update(msg)
			return m, spCmd
		}
	}

	m.textInput, tiCmd = m.textInput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd, spCmd)
}

// View renders the UI
func (m Model) View() string {
	// Footer with input field
	input := promptStyle.Render("> ") + m.textInput.View()
	
	// Spinner when loading
	if m.loading {
		input = m.spinner.View() + " Processing..."
	}
	
	// Help text
	helpText := helpStyle.Render("Ctrl+C to quit, Ctrl+L to clear")
	
	// Put it all together
	return fmt.Sprintf(
		"%s\n%s\n\n%s\n%s",
		titleStyle.Render("Terminal Claude Assistant"),
		m.viewport.View(),
		input,
		helpText,
	)
}
