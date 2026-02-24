package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all persisted CLI settings.
// Think of this as a plain DTO / record in C#.
type Config struct {
	BaseURL string `json:"base_url"`
	// We'll add token fields here later (Step 4)
}

// configFilePath returns the path to the config file.
// On Linux/macOS: ~/.config/webdoc/config.json
// On Windows:     %APPDATA%\webdoc\config.json
func configFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not find config directory: %w", err)
	}
	return filepath.Join(dir, "webdoc", "config.json"), nil
}

// Load reads the config file from disk.
// Returns an empty Config (not an error) if the file doesn't exist yet.
func Load() (*Config, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Config{}, nil // first run, no config yet
	}
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	return &cfg, nil
}

// Save writes the config to disk, creating directories if needed.
func (c *Config) Save() error {
	path, err := configFilePath()
	if err != nil {
		return err
	}

	// Create ~/.config/webdoc/ if it doesn't exist
	// 0755 = rwxr-xr-x (owner can read/write/execute, others can read/execute)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("could not serialize config: %w", err)
	}

	// 0600 = rw------- (only owner can read/write — important for credentials)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	return nil
}

// ResolveBaseURL returns the flag value if set, otherwise falls back to config.
// This is the "flag wins" part of your design decision.
func ResolveBaseURL(flagValue string, cfg *Config) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if cfg.BaseURL != "" {
		return cfg.BaseURL, nil
	}
	return "", errors.New(
		"no base URL configured — run `webdoc config set-url <url>` or pass --url",
	)
}
