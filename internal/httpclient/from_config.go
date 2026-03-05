package httpclient

import (
	"github.com/chrilleson/webdoc-cli/internal/auth"
	"github.com/chrilleson/webdoc-cli/internal/config"
)

func FromConfig(apiURLFlag string) (*Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	apiURL, err := config.ResolveAPIURL(apiURLFlag, cfg)
	if err != nil {
		return nil, err
	}

	token, err := auth.GetValidToken(cfg)
	if err != nil {
		return nil, err
	}

	return New(apiURL, token), nil
}
