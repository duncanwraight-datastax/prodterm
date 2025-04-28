package config

import (
	"errors"
	"os"
)

// Config holds application configuration
type Config struct {
	AnthropicAPIKey string
	Model           string
}

// Load configuration from environment variables
func Load() (Config, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return Config{}, errors.New("ANTHROPIC_API_KEY environment variable not set")
	}
	
	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-3-sonnet-20240229" // Default model
	}
	
	return Config{
		AnthropicAPIKey: apiKey,
		Model:           model,
	}, nil
}
