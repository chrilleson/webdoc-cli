package main

import (
	"fmt"
	"os"
	"time"

	"github.com/chrilleson/webdoc-cli/internal/auth"
	"github.com/chrilleson/webdoc-cli/internal/config"
	"github.com/spf13/cobra"
)

func main() {
	var urlFlag string

	rootCmd := &cobra.Command{
		Use:   "webdoc",
		Short: "CLI for the Webdoc EMR API",
	}

	// --url flag available on ALL subcommands
	rootCmd.PersistentFlags().StringVar(&urlFlag, "url", "", "Override base URL (e.g. https://test.clinic.webdoc.com)")

	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with the Webdoc API",
	}

	var clientID, clientSecret string
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Obtain and cache an OAuth2 access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			baseURL, err := config.ResolveBaseURL(urlFlag, cfg)
			if err != nil {
				return err
			}
			if err := auth.Login(baseURL, clientID, clientSecret); err != nil {
				return nil
			}
			fmt.Println("Login successful")
			return nil
		},
	}
	loginCmd.Flags().StringVar(&clientID, "client-id", "", "OAuth2 client ID (required)")
	loginCmd.Flags().StringVar(&clientSecret, "client-secret", "", "OAuth2 client secret (required)")
	loginCmd.MarkFlagRequired("client-id")
	loginCmd.MarkFlagRequired("client-secret")

	authStatusCmd := &cobra.Command{
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

	// `webdoc config` group
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}

	// `webdoc config set-url <url>`
	setURLCmd := &cobra.Command{
		Use:   "set-url <url>",
		Short: "Set and persist the Webdoc base URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cfg.BaseURL = args[0]
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Printf("Base URL saved: %s\n", args[0])
			return nil
		},
	}

	// `webdoc config show`
	showConfigCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if cfg.BaseURL == "" {
				fmt.Println("base_url: (not set)")
			} else {
				fmt.Printf("base_url: %s\n", cfg.BaseURL)
			}
			return nil
		},
	}

	// Placeholder commands — we'll flesh these out in later steps
	patientsCmd := &cobra.Command{
		Use:   "patients",
		Short: "Manage patients",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			url, err := config.ResolveBaseURL(urlFlag, cfg)
			if err != nil {
				return err
			}
			fmt.Printf("patients command — base URL: %s\n", url)
			return nil
		},
	}

	bookingsCmd := &cobra.Command{
		Use:   "bookings",
		Short: "Manage bookings and calendar",
	}

	documentsCmd := &cobra.Command{
		Use:   "documents",
		Short: "Manage documents",
	}

	// Assemble the command tree
	authCmd.AddCommand(loginCmd, authStatusCmd)
	configCmd.AddCommand(setURLCmd, showConfigCmd)
	rootCmd.AddCommand(authCmd, configCmd, patientsCmd, bookingsCmd, documentsCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
