package models

// AnthropicResponse from Anthropic API
type AnthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
	ID string `json:"id"`
}

// MessageContent represents a content item in a message
type MessageContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Message represents a message in the conversation with structured content
type Message struct {
	Role    string          `json:"role"`
	Content []MessageContent `json:"content"`
}

// AnthropicRequest to Anthropic API
type AnthropicRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
}

// ChatMessage represents an older format message in the conversation
// Kept for backward compatibility
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
