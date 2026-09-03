package commands

import (
	"fmt"

	"github.com/chrilleson/webdoc-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewConfigCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}
	cmd.AddCommand(
		newConfigSetAuthURLCmd(),
		newConfigSetAPIURLCmd(),
		newConfigSetClinicIDCmd(),
		newConfigShowCmd(),
	)
	return cmd
}

func newConfigSetAuthURLCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-auth-url <url>",
		Short: "Set and persist the Webdoc auth URL (e.g. https://auth-integration.carasent.net)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cfg.AuthURL = args[0]
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Printf("Auth URL saved: %s\n", args[0])
			return nil
		},
	}
}

func newConfigSetAPIURLCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-api-url <url>",
		Short: "Set and persist the Webdoc API URL (e.g. https://api.atlan.se)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cfg.APIURL = args[0]
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Printf("API URL saved: %s\n", args[0])
			return nil
		},
	}
}

func newConfigSetClinicIDCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-clinic-id <id>",
		Short: "Set and persist the default clinic ID used by clinic-scoped commands",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cfg.ClinicID = args[0]
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Printf("Clinic ID saved: %s\n", args[0])
			return nil
		},
	}
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			authURL := cfg.AuthURL
			if authURL == "" {
				authURL = "(not set)"
			}
			apiURL := cfg.APIURL
			if apiURL == "" {
				apiURL = "(not set)"
			}
			clinicID := cfg.ClinicID
			if clinicID == "" {
				clinicID = "(not set)"
			}
			fmt.Printf("auth_url:  %s\n", authURL)
			fmt.Printf("api_url:   %s\n", apiURL)
			fmt.Printf("clinic_id: %s\n", clinicID)
			return nil
		},
	}
}
