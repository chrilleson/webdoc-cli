// Package commands holds one file per API resource. Each file exposes a single
// exported constructor - func NewXxxCmd(o *Options) *cobra.Command - which builds
// that resource's command group and all of its subcommands. cmd/webdoc/main.go
// only wires the constructors onto the root command.
package commands

import (
	"github.com/chrilleson/webdoc-cli/internal/config"
	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

// Options carries the root persistent flags. It is filled by cobra before any
// RunE runs, so read the fields inside RunE, never at construction time.
type Options struct {
	APIURL   string // --url
	ClinicID string // --clinic-id
	JSON     bool   // --json
}

// Client builds an authenticated API client honouring --url.
func (o *Options) Client() (*httpclient.Client, error) {
	return httpclient.FromConfig(o.APIURL)
}

// Clinic resolves the clinic ID for clinic-scoped endpoints:
// --clinic-id first, then the persisted config value.
func (o *Options) Clinic() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}
	return config.ResolveClinicID(o.ClinicID, cfg)
}
