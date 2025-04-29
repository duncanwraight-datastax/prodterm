package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"terminal-claude/config"
	"terminal-claude/models"
)

// Client represents a Claude API client
type Client struct {
	Config config.Config
}

// NewClient creates a new Claude API client
func NewClient(cfg config.Config) *Client {
	return &Client{
		Config: cfg,
	}
}

// Ask sends a prompt to Claude AI and returns the response
func (c *Client) Ask(prompt string) (string, error) {
	url := "https://api.anthropic.com/v1/messages"
	
	// Force the correct model name
	modelName := "claude-3-haiku-20240307"
	
	// Create a simpler message structure
	requestBody := models.AnthropicRequest{
		Model:     modelName, // Use the hardcoded model name for now
		MaxTokens: 1024,
		System:    "You are Claude, an AI assistant by Anthropic. You're helpful, harmless, and honest.",
	}
	
	// Check if the prompt might be too long or has formatting issues
	if len(prompt) > 100000 {
		// Truncate if needed
		prompt = prompt[:100000]
	}
	
	// Ensure we have non-empty content
	if prompt == "" {
		prompt = "Hello"
	}
	
	// Create proper content structure for the user message
	requestBody.Messages = []models.Message{
		{
			Role: "user",
			Content: []models.MessageContent{
				{
					Type: "text",
					Text: prompt,
				},
			},
		},
	}
	
	// Verify content is not empty
	if len(requestBody.Messages) == 0 || len(requestBody.Messages[0].Content) == 0 {
		return "", fmt.Errorf("cannot create empty message content")
	}
	
	if requestBody.Messages[0].Content[0].Text == "" {
		return "", fmt.Errorf("message text cannot be empty")
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.Config.AnthropicAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01") 
	// Use a newer API version
	req.Header.Set("anthropic-beta", "messages-2023-12-15")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request to Claude: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errorMsg := string(bodyBytes)
		return "", fmt.Errorf("error from Claude API (Status %d): %s", resp.StatusCode, errorMsg)
	}
	
	var result models.AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}
	
	// Extract the text from the response
	var responseText string
	for _, content := range result.Content {
		if content.Type == "text" {
			responseText += content.Text
		}
	}
	
	return responseText, nil
}
