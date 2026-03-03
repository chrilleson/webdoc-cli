package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	BaseURL     string    `json:"base_url"`
	AccessToken string    `json:"access_token,omitempty"`
	TokenExpiry time.Time `json:"token_expiry,omitempty"`
}

func configFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not find config directory: %w", err)
	}
	return filepath.Join(dir, "webdoc", "config.json"), nil
}

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

func (c *Config) Save() error {
	path, err := configFilePath()
	if err != nil {
		return err
	}

	// 0755 = rwxr-xr-x (owner can read/write/execute, others can read/execute)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("could not serialize config: %w", err)
	}

	// 0600 = rw------- (only owner can read/write — important for credentials)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	return nil
}

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

func (c *Config) IsTokenValid() bool {
	return c.AccessToken != "" && time.Now().Before(c.TokenExpiry.Add(-30*time.Second))
}
