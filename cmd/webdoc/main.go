package main

import (
	"fmt"
	"os"
	"time"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/chrilleson/webdoc-cli/internal/auth"
	"github.com/chrilleson/webdoc-cli/internal/config"
	"github.com/chrilleson/webdoc-cli/internal/httpclient"
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

	// -- auth --------------------------------------------------
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with the Webdoc API",
	}

	var clientID, clientSecret, scope string
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
			if err := auth.Login(baseURL, clientID, clientSecret, scope); err != nil {
				return err
			}
			fmt.Println("Login successful")
			return nil
		},
	}
	loginCmd.Flags().StringVar(&clientID, "client-id", "", "OAuth2 client ID (required)")
	loginCmd.Flags().StringVar(&clientSecret, "client-secret", "", "OAuth2 client secret (required)")
	loginCmd.Flags().StringVar(&scope, "scope", "self-service", "OAuth2 scopes (space-separated)")
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

	// -- config ------------------------------------------------------

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}

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

	// -- booking types -----------------------------------------------------
	bookingTypesCmd := &cobra.Command{
		Use:   "booking-types",
		Short: "Manage booking types",
	}

	bookingTypesList := &cobra.Command{
		Use:   "list",
		Short: "List all booking types",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := httpclient.FromConfig(urlFlag)
			if err != nil {
				return err
			}

			bookingTypes, err := api.GetBookingTypes(client)
			if err != nil {
				return err
			}

			if len(bookingTypes) == 0 {
				fmt.Println("No booking types found")
				return nil
			}

			for _, bt := range bookingTypes {
				fmt.Printf("%-4s %s", bt.ID, bt.Name)
				if bt.ExternallyVisibleName != bt.Name {
					fmt.Printf(" (%s)", bt.ExternallyVisibleName)
				}
				if bt.HasSelfService {
					fmt.Print(" [self-service]")
				}

				fmt.Println()
			}

			return nil
		},
	}

	// -- assemble tree ------------------------------------------------------
	authCmd.AddCommand(loginCmd, authStatusCmd)
	configCmd.AddCommand(setURLCmd, showConfigCmd)
	bookingTypesCmd.AddCommand(bookingTypesList)

	rootCmd.AddCommand(
		authCmd,
		configCmd,
		bookingTypesCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
