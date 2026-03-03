package httpclient

import (
	"github.com/chrilleson/webdoc-cli/internal/auth"
	"github.com/chrilleson/webdoc-cli/internal/config"
)

func FromConfig(urlFlag string) (*Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	baseURL, err := config.ResolveBaseURL(urlFlag, cfg)
	if err != nil {
		return nil, err
	}

	token, err := auth.GetValidToken(cfg)
	if err != nil {
		return nil, err
	}

	return New(baseURL, token), nil
}
