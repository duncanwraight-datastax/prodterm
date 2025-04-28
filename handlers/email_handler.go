package handlers

import (
	"fmt"
	"terminal-claude/api"
	"terminal-claude/mcp"
	"time"
)

// HandleEmailSummary creates a summary of unread emails
func (h *Handler) HandleEmailSummary() (string, error) {
	// Get the Gmail provider
	result, err := mcp.ExecuteCommand("Gmail", "summarize_unread", map[string]interface{}{
		"count": 10,
	})
	
	if err != nil {
		return "", fmt.Errorf("failed to get unread emails: %v", err)
	}
	
	// Convert result to a format we can use
	summary, ok := result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected result type")
	}
	
	count, _ := summary["count"].(int)
	emails, _ := summary["emails"].([]map[string]interface{})
	
	if count == 0 {
		return "You have no unread emails.", nil
	}
	
	// Format the email data for Claude
	emailData := fmt.Sprintf("You have %d unread emails:\n", count)
	
	for i, email := range emails {
		from, _ := email["from"].(string)
		subject, _ := email["subject"].(string)
		dateStr, _ := email["date"].(string)
		
		// Parse the date
		date, err := parseEmailDate(dateStr)
		var timeAgo string
		if err == nil {
			timeAgo = timeAgo(date)
		} else {
			timeAgo = dateStr
		}
		
		emailData += fmt.Sprintf("%d. From: %s, Subject: %s, Received: %s\n", 
			i+1, from, subject, timeAgo)
		
		if snippet, ok := email["snippet"].(string); ok && snippet != "" {
			emailData += fmt.Sprintf("   Snippet: %s\n", snippet)
		}
	}
	
	prompt := "Here are my unread emails. Please provide a brief summary of each, including who they're from and what they appear to be about:\n\n" + emailData
	
	return h.claudeClient.Ask(prompt)
}

// parseEmailDate parses an email date string
func parseEmailDate(dateStr string) (time.Time, error) {
	// Try a few formats
	layouts := []string{
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 -0700 (MST)",
		time.RFC1123Z,
		time.RFC822Z,
		time.RFC822,
	}
	
	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("could not parse date: %s", dateStr)
}

// timeAgo returns a human-readable string representing how long ago a time was
func timeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)
	
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 48*time.Hour:
		return "yesterday"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2")
	}
}
