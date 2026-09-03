package commands

import (
	"fmt"
	"strings"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

// NewHealthCmd is the top-level `webdoc health` command. It has no
// subcommands and exits non-zero when the API does not report "ok", so it can
// be used directly in a shell check.
func NewHealthCmd(o *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check API health",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			health, err := api.GetHealth(client)
			if err != nil {
				return err
			}

			if err := Emit(o, health, func() {
				fmt.Printf("status: %s\n", health.Status)
			}); err != nil {
				return err
			}

			if !strings.EqualFold(health.Status, "ok") {
				return fmt.Errorf("API is not healthy: status %q", health.Status)
			}
			return nil
		},
	}
}
