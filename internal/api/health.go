package api

import (
	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

type Health struct {
	Status string `json:"status"`
}

func GetHealth(c *httpclient.Client) (Health, error) {
	return httpclient.Get[Health](c, "/v1/health-check", nil)
}
