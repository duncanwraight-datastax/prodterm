package mcp

import (
	"fmt"
	"sync"
)

var (
	registry = make(map[string]Provider)
	mutex    sync.RWMutex
)

// Register adds a provider to the registry
func Register(provider Provider) {
	mutex.Lock()
	defer mutex.Unlock()
	registry[provider.Name()] = provider
}

// Get returns a provider by name
func Get(name string) (Provider, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	
	if provider, ok := registry[name]; ok {
		return provider, nil
	}
	
	return nil, fmt.Errorf("provider not found: %s", name)
}

// ListProviders returns a list of registered providers
func ListProviders() []string {
	mutex.RLock()
	defer mutex.RUnlock()
	
	var providers []string
	for name := range registry {
		providers = append(providers, name)
	}
	
	return providers
}

// ExecuteCommand executes a command on a provider
func ExecuteCommand(provider string, command string, params map[string]interface{}) (interface{}, error) {
	p, err := Get(provider)
	if err != nil {
		return nil, err
	}
	
	return p.Execute(command, params)
}
