package mcp

// Provider defines the interface for a Model Context Protocol provider
type Provider interface {
	// Name returns the provider's name
	Name() string
	
	// GetCapabilities returns the provider's capabilities
	GetCapabilities() []Capability
	
	// Execute runs a command with the given parameters
	Execute(command string, params map[string]interface{}) (interface{}, error)
}

// Capability represents a provider capability
type Capability struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
}
