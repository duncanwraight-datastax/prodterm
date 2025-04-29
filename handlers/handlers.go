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
	
	// Check if it's an email command
	if strings.Contains(command, "unread emails") || strings.Contains(command, "unread e-mails") {
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
	} else if strings.HasPrefix(command, "list slack channels") || strings.HasPrefix(command, "show slack channels") {
		// List available Slack channels
		return h.HandleSlackChannels()
	} else if strings.Contains(command, "summarize slack channel") || strings.Contains(command, "summarise slack channel") {
		// Extract channel name or ID
		var channel string
		patterns := []string{"channel", "in", "#"}
		for _, pattern := range patterns {
			parts := strings.SplitN(command, pattern, 2)
			if len(parts) > 1 {
				channel = strings.TrimSpace(parts[1])
				if channel != "" {
					break
				}
			}
		}
		
		if channel == "" {
			// Try to find the channel name at the end
			words := strings.Fields(command)
			if len(words) > 0 {
				candidate := words[len(words)-1]
				if strings.HasPrefix(candidate, "#") || strings.HasPrefix(candidate, "C") {
					channel = candidate
				}
			}
		}
		
		if channel != "" {
			return h.HandleSlackSummary(channel)
		}
		return "Please specify a Slack channel name or ID to summarize.", nil
	} else {
		// For any other command, pass it directly to Claude
		response, err := h.claudeClient.Ask(command)
		if err != nil {
			return "", err
		}
		return response, nil
	}
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
