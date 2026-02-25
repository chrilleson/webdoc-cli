package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chrilleson/webdoc-cli/internal/config"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func Login(baseURL, clientID, clientSecret, scope string) error {
	if !strings.Contains(scope, "self-service") {
		scope = "self-service " + scope
	}
	tokenURL := strings.TrimRight(baseURL, "/") + "/oauth/token"

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("scope", scope)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	resp, err := http.Post(
		tokenURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed with status %d. Check your client_id and client_secrets", resp.StatusCode)
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
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
