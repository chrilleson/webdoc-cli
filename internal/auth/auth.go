package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chrilleson/webdoc-cli/internal/config"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   int    `json:"token_type"`
}

func Login(baseURL, clientID, clientSecret string) error {
	tokenUrl := strings.TrimRight(baseURL, "/") + "/oauth/token"

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	resp, err := http.Post(
		tokenUrl,
		"application/x-www-form-urlencoded",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed with status %d. Check your client_id and client_secrets", resp.StatusCode)
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("could not parse token response: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	cfg.AccessToken = tokenResp.AccessToken
	cfg.TokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("could not save token: %w", err)
	}

	return nil
}

func GetValidToken(cfg *config.Config) (string, error) {
	if !cfg.IsTokenValid() {
		return "", fmt.Errorf("not authenticated. run `webdoc auth login` first")
	}

	return cfg.AccessToken, nil
}
