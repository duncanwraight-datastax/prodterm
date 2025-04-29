package handlers

import (
	"fmt"
	"strings"
	"terminal-claude/mcp"
)

// HandleSlackSummary summarizes recent messages from a Slack channel
func (h *Handler) HandleSlackSummary(channel string) (string, error) {
	// Determine if input is a channel ID or name
	var params map[string]interface{}
	if strings.HasPrefix(channel, "C") && len(channel) == 9 {
		// Looks like a channel ID
		params = map[string]interface{}{
			"channel_id": channel,
			"count":      20,
		}
	} else {
		// Assume it's a channel name
		params = map[string]interface{}{
			"channel": strings.TrimPrefix(channel, "#"),
			"count":   20,
		}
	}

	// Get the Slack provider to summarize the channel
	result, err := mcp.ExecuteCommand("Slack", "summarize_channel", params)
	if err != nil {
		// Check if this is a provider not found error and provide helpful instructions
		if strings.Contains(err.Error(), "provider not found") {
			return "", fmt.Errorf("Slack integration is not configured. Please see docs/slack_setup.md for setup instructions")
		}
		
		// For authentication errors
		if strings.Contains(err.Error(), "authentication") || strings.Contains(err.Error(), "token") {
			return "", fmt.Errorf("Slack authentication failed. Please check your token in ~/.config/terminal-claude/slack_token.txt")
		}
		
		return "", fmt.Errorf("failed to summarize Slack channel: %v", err)
	}

	// Convert result to a format we can use
	summary, ok := result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected result type")
	}

	channelName, _ := summary["channel_name"].(string)
	messages, _ := summary["messages"].([]map[string]interface{})

	if len(messages) == 0 {
		return fmt.Sprintf("No recent messages found in #%s", channelName), nil
	}

	// Format the channel data for Claude
	channelData := fmt.Sprintf("Recent messages from #%s (newest first):\n\n", channelName)

	for i, message := range messages {
		user, _ := message["user"].(string)
		text, _ := message["text"].(string)
		timeAgo, _ := message["time_ago"].(string)

		channelData += fmt.Sprintf("%d. %s (%s): %s\n", 
			i+1, user, timeAgo, text)
	}

	prompt := fmt.Sprintf("Here are recent messages from a Slack channel. Please provide:\n"+
		"1. A concise summary of the main topics and discussions\n"+
		"2. Any important decisions or action items\n"+
		"3. Any questions that appear to need answers\n\n%s", channelData)

	return h.claudeClient.Ask(prompt)
}

// HandleSlackChannels lists available Slack channels
func (h *Handler) HandleSlackChannels() (string, error) {
	// Get the Slack provider to list channels
	result, err := mcp.ExecuteCommand("Slack", "list_channels", map[string]interface{}{})
	if err != nil {
		// Check if this is a provider not found error and provide helpful instructions
		if strings.Contains(err.Error(), "provider not found") {
			return "", fmt.Errorf("Slack integration is not configured. Please see docs/slack_setup.md for setup instructions")
		}
		
		// For authentication errors
		if strings.Contains(err.Error(), "authentication") || strings.Contains(err.Error(), "token") {
			return "", fmt.Errorf("Slack authentication failed. Please check your token in ~/.config/terminal-claude/slack_token.txt")
		}
		
		return "", fmt.Errorf("failed to list Slack channels: %v", err)
	}

	// Convert result to a format we can use
	channelList, ok := result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected result type")
	}

	channels, _ := channelList["channels"].([]map[string]interface{})

	if len(channels) == 0 {
		return "No Slack channels found.", nil
	}

	// Format the channel list
	response := "Available Slack channels:\n\n"

	for _, channel := range channels {
		name, _ := channel["name"].(string)
		topic, _ := channel["topic"].(string)
		memberCount, _ := channel["member_count"].(float64)

		if topic != "" {
			response += fmt.Sprintf("#%s (%d members) - %s\n", 
				name, int(memberCount), topic)
		} else {
			response += fmt.Sprintf("#%s (%d members)\n", 
				name, int(memberCount))
		}
	}

	return response, nil
}