package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"terminal-claude/api"
	"terminal-claude/config"
)

// Handler processes user commands
type Handler struct {
	claudeClient *api.Client
}

// NewHandler creates a new command handler
func NewHandler(cfg config.Config) *Handler {
	return &Handler{
		claudeClient: api.NewClient(cfg),
	}
}

// ProcessCommand handles different types of user commands
func (h *Handler) ProcessCommand(command string) (string, error) {
	command = strings.TrimSpace(command)
	
	if command == "exit" {
		return "Exiting...", nil
	}
	
	// Parse the command to determine what to do
	if strings.HasPrefix(command, "summarise my unread e-mails") || 
	   strings.HasPrefix(command, "summarize my unread e-mails") {
		return h.HandleEmailSummary()
	} else if strings.HasPrefix(command, "what's on this webpage?") || 
	          strings.HasPrefix(command, "what's on this webpage") {
		parts := strings.SplitN(command, "?", 2)
		var url string
		if len(parts) > 1 {
			url = strings.TrimSpace(parts[1])
		} else {
			// Try to extract URL from the original command
			fields := strings.Fields(command)
			if len(fields) > 4 {
				url = fields[len(fields)-1]
			}
		}
		
		if url != "" {
			return h.HandleWebpageSummary(url)
		}
		return "Please provide a URL to summarize.", nil
	} else {
		// For any other command, pass it directly to Claude
		response, err := h.claudeClient.Ask(command)
		if err != nil {
			return "", err
		}
		return response, nil
	}
}

// HandleEmailSummary creates a summary of unread emails
func (h *Handler) HandleEmailSummary() (string, error) {
	// This is a placeholder - in a real implementation, you would:
	// 1. Connect to the email server
	// 2. Fetch unread emails
	// 3. Format them for Claude
	// 4. Send to Claude for summarization
	
	emailData := "You have 3 unread emails:\n" +
		"1. From: boss@company.com, Subject: Project Update Meeting\n" +
		"2. From: newsletter@tech.com, Subject: Weekly Tech Digest\n" +
		"3. From: friend@gmail.com, Subject: Weekend Plans"
	
	prompt := "Here are my unread emails. Please provide a brief summary of each:\n\n" + emailData
	
	return h.claudeClient.Ask(prompt)
}

// HandleWebpageSummary creates a summary of a webpage
func (h *Handler) HandleWebpageSummary(url string) (string, error) {
	// Fetch webpage content
	content, err := h.fetchWebpage(url)
	if err != nil {
		return "", fmt.Errorf("error fetching webpage: %v", err)
	}
	
	// Truncate content if it's too long
	if len(content) > 8000 {
		content = content[:8000] + "... (content truncated)"
	}
	
	prompt := fmt.Sprintf("Please summarize the content of this webpage from %s:\n\n%s", url, content)
	
	return h.claudeClient.Ask(prompt)
}

// fetchWebpage retrieves content from a URL
func (h *Handler) fetchWebpage(url string) (string, error) {
	// Add http:// prefix if not present
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	return string(body), nil
}
