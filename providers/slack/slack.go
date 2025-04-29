package slack

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"terminal-claude/mcp"
	"time"

	"github.com/slack-go/slack"
)

// Provider implements the Model Context Protocol for Slack
type Provider struct {
	client *slack.Client
}

// New creates a new Slack provider
func New() (*Provider, error) {
	// Get token from environment variable or file
	token, err := getSlackToken()
	if err != nil {
		return nil, fmt.Errorf("unable to get Slack token: %v", err)
	}

	// Create the Slack client
	client := slack.New(token)

	// Test the connection - but make this optional
	_, err = client.AuthTest()
	if err != nil {
		log.Printf("Warning: Slack authentication not successful, but continuing anyway: %v", err)
		// We'll continue even with authentication errors
		// This allows the provider to be registered but will return appropriate errors when used
	}

	return &Provider{
		client: client,
	}, nil
}

// Name returns the provider's name
func (p *Provider) Name() string {
	return "Slack"
}

// GetCapabilities returns the provider's capabilities
func (p *Provider) GetCapabilities() []mcp.Capability {
	return []mcp.Capability{
		{
			Name:        "messages",
			Description: "Access and summarize Slack messages",
			Commands:    []string{"list_channels", "recent_messages", "summarize_channel"},
		},
	}
}

// Execute runs a command with the given parameters
func (p *Provider) Execute(command string, params map[string]interface{}) (interface{}, error) {
	switch command {
	case "list_channels":
		return p.listChannels()
	case "recent_messages":
		channelID, ok := params["channel_id"].(string)
		if !ok {
			return nil, fmt.Errorf("channel_id parameter required")
		}
		count := 10 // Default count
		if c, ok := params["count"].(float64); ok {
			count = int(c)
		}
		return p.recentMessages(channelID, count)
	case "summarize_channel":
		channelID, ok := params["channel_id"].(string)
		if !ok {
			// Try to get channel by name if ID is not provided
			channelName, ok := params["channel"].(string)
			if !ok {
				return nil, fmt.Errorf("either channel_id or channel parameter required")
			}
			var err error
			channelID, err = p.getChannelIDByName(channelName)
			if err != nil {
				return nil, err
			}
		}
		count := 10 // Default count
		if c, ok := params["count"].(float64); ok {
			count = int(c)
		}
		return p.summarizeChannel(channelID, count)
	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// listChannels lists available Slack channels
func (p *Provider) listChannels() (map[string]interface{}, error) {
	channels, cursor, err := p.client.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list channels: %v", err)
	}

	var channelList []map[string]interface{}
	for _, channel := range channels {
		channelList = append(channelList, map[string]interface{}{
			"id":   channel.ID,
			"name": channel.Name,
			"is_private": channel.IsPrivate,
			"topic": channel.Topic.Value,
			"member_count": channel.NumMembers,
		})
	}

	return map[string]interface{}{
		"channels": channelList,
		"cursor":   cursor,
	}, nil
}

// recentMessages gets recent messages from a channel
func (p *Provider) recentMessages(channelID string, count int) (map[string]interface{}, error) {
	// Get messages from the channel
	history, err := p.client.GetConversationHistory(&slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     count,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get channel history: %v", err)
	}

	messages := []map[string]interface{}{}
	for _, msg := range history.Messages {
		// Get user info if available
		var username string
		if msg.User != "" {
			user, err := p.client.GetUserInfo(msg.User)
			if err == nil {
				username = user.RealName
				if username == "" {
					username = user.Name
				}
			} else {
				username = msg.User
			}
		} else {
			username = "Unknown"
		}

		// Format timestamp
		timestamp, _ := parseSlackTimestamp(msg.Timestamp)
		
		// Add the message to the list
		messages = append(messages, map[string]interface{}{
			"user":      username,
			"text":      msg.Text,
			"timestamp": timestamp.Format(time.RFC3339),
			"time_ago":  formatTimeAgo(timestamp),
		})
	}

	// Get channel info
	channel, err := p.client.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID: channelID,
	})
	if err != nil {
		return map[string]interface{}{
			"channel_id": channelID,
			"messages":   messages,
		}, nil
	}

	return map[string]interface{}{
		"channel_id":   channelID,
		"channel_name": channel.Name,
		"messages":     messages,
	}, nil
}

// summarizeChannel gets a summary of a channel
func (p *Provider) summarizeChannel(channelID string, count int) (map[string]interface{}, error) {
	result, err := p.recentMessages(channelID, count)
	if err != nil {
		return nil, err
	}

	// Get channel info
	channel, err := p.client.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID: channelID,
	})
	
	var channelName string
	if err == nil {
		channelName = channel.Name
	} else {
		channelName = channelID
	}

	// Add additional metadata
	result["count"] = count
	result["channel_name"] = channelName
	
	return result, nil
}

// getChannelIDByName gets a channel ID from a channel name
func (p *Provider) getChannelIDByName(channelName string) (string, error) {
	// Remove the # prefix if present
	channelName = strings.TrimPrefix(channelName, "#")
	
	channels, _, err := p.client.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
	})
	if err != nil {
		return "", fmt.Errorf("unable to list channels: %v", err)
	}

	for _, channel := range channels {
		if channel.Name == channelName {
			return channel.ID, nil
		}
	}

	return "", fmt.Errorf("channel not found: %s", channelName)
}

// getSlackToken gets the Slack API token
func getSlackToken() (string, error) {
	// Try to get from environment
	token := os.Getenv("SLACK_TOKEN")
	if token != "" {
		return token, nil
	}

	// Try to get from file
	tokenPath := getTokenPath()
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("unable to read token file: %v", err)
	}

	return strings.TrimSpace(string(data)), nil
}

// getTokenPath returns the path to the token file
func getTokenPath() string {
	tokenPath := os.Getenv("SLACK_TOKEN_PATH")
	if tokenPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Unable to get home directory: %v", err)
		}
		tokenPath = filepath.Join(homeDir, ".config", "terminal-claude", "slack_token.txt")
	}
	return tokenPath
}

// parseSlackTimestamp converts a Slack timestamp to a time.Time
func parseSlackTimestamp(timestamp string) (time.Time, error) {
	// Slack timestamps are in the format "1234567890.123456"
	parts := strings.Split(timestamp, ".")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid timestamp format: %s", timestamp)
	}

	// Parse the seconds
	seconds, err := fmt.Sscanf(parts[0], "%d", new(int64))
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse timestamp: %v", err)
	}

	return time.Unix(int64(seconds), 0), nil
}

// formatTimeAgo returns a human-readable string representing how long ago a time was
func formatTimeAgo(t time.Time) string {
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

// init registers the provider
func init() {
	provider, err := New()
	if err != nil {
		log.Printf("Warning: Unable to initialize Slack provider: %v", err)
		return
	}
	
	mcp.Register(provider)
}