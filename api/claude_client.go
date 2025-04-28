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
	
	messages := []models.ChatMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}
	
	requestBody := models.AnthropicRequest{
		Model:    c.Config.Model,
		Messages: messages,
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
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request to Claude: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error from Claude API (Status %d): %s", resp.StatusCode, string(bodyBytes))
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
