package commands

import (
	"fmt"
	"time"

	"github.com/chrilleson/webdoc-cli/internal/auth"
	"github.com/chrilleson/webdoc-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewAuthCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with the Webdoc API",
	}
	cmd.AddCommand(newAuthLoginCmd(), newAuthStatusCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var clientID, clientSecret, scope string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Obtain and cache an OAuth2 access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			authURL, err := config.ResolveAuthURL("", cfg)
			if err != nil {
				return err
			}
			if err := auth.Login(authURL, clientID, clientSecret, scope); err != nil {
				return err
			}
			fmt.Println("Login successful")
			return nil
		},
	}
	cmd.Flags().StringVar(&clientID, "client-id", "", "OAuth2 client ID (required)")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "OAuth2 client secret (required)")
	cmd.Flags().StringVar(&scope, "scope", "self-service", "OAuth2 scopes (space-separated)")
	cmd.MarkFlagRequired("client-id")
	cmd.MarkFlagRequired("client-secret")

	return cmd
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check if the current token is valid",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if cfg.IsTokenValid() {
				fmt.Printf("Authenticated ✓  (token expires %s)\n", cfg.TokenExpiry.Format(time.RFC1123))
			} else {
				fmt.Println("Not authenticated — run `webdoc auth login`")
			}
			return nil
		},
	}
}
