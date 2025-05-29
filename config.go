package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	// Application name for keyring
	appName = "OpenRouterProxy"
	// Username for keyring (we use a fixed value)
	userName = "openrouter-proxy-user"
	// Key for API key in keyring
	apiKeyName = "openrouter-api-key"
)

// Config holds the application configuration
type Config struct {
	// ServerEnabled indicates if the proxy server is running
	ServerEnabled bool `json:"server_enabled"`
	// LastUsedModelFilter is the path to the last used model filter file
	LastUsedModelFilter string `json:"last_used_model_filter"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		ServerEnabled:       false,
		LastUsedModelFilter: "models-filter",
	}
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Create .openrouter-proxy directory if it doesn't exist
	configDir := filepath.Join(homeDir, ".openrouter-proxy")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}

// LoadConfig loads the configuration from disk
func LoadConfig() (Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), err
	}

	// Parse config
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultConfig(), err
	}

	return config, nil
}

// SaveConfig saves the configuration to disk
func SaveConfig(config Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Write config file
	return os.WriteFile(configPath, data, 0644)
}

// GetAPIKey retrieves the API key from the keyring
func GetAPIKey() (string, error) {
	return keyring.Get(appName, userName)
}

// SetAPIKey stores the API key in the keyring
func SetAPIKey(apiKey string) error {
	return keyring.Set(appName, userName, apiKey)
}

// HasAPIKey checks if an API key is stored
func HasAPIKey() bool {
	_, err := GetAPIKey()
	return err == nil
}