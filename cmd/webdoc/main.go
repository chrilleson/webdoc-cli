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
	var apiURLFlag string

	rootCmd := &cobra.Command{
		Use:   "webdoc",
		Short: "CLI for the Webdoc EMR API",
	}

	// --url flag available on ALL subcommands
	rootCmd.PersistentFlags().StringVar(&apiURLFlag, "url", "", "Override API base URL")

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

	setAuthURLCmd := &cobra.Command{
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

	setAPIURLCmd := &cobra.Command{
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

	showConfigCmd := &cobra.Command{
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
			fmt.Printf("auth_url: %s\n", authURL)
			fmt.Printf("api_url:  %s\n", apiURL)
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
			client, err := httpclient.FromConfig(apiURLFlag)
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

	// -- users --------------------------------------------------------------`json:
	usersCmd := &cobra.Command{
		Use:   "users",
		Short: "Manage users",
	}

	usersSearchCmd := &cobra.Command{
		Use:   "search <personalNumber>",
		Short: "Search for a user by personalNumber",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := httpclient.FromConfig(apiURLFlag)
			if err != nil {
				return err
			}

			users, err := api.SearchUsers(client, args[0])
			if err != nil {
				return nil
			}

			if len(users) == 0 {
				fmt.Println("No users found")
				return nil
			}

			for _, u := range users {
				fmt.Printf("%s %s %s (%s)\n", u.ID, u.FirstName, u.LastName, u.PersonalNumber)
				for _, c := range u.Clinics {
					fmt.Printf("  └─ %s  %s\n", c.ID, c.Name)
				}
			}

			return nil
		},
	}

	// -- patients -----------------------------------------------------------
	patientsCmd := &cobra.Command{
		Use:   "patients",
		Short: "Manage patients",
	}

	patientsSearchCmd := &cobra.Command{
		Use:   "search <personalNumber>",
		Short: "Search for a patient by personal number",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := httpclient.FromConfig(apiURLFlag)
			if err != nil {
				return err
			}

			patients, err := api.GetPatients(client, args[0])
			if err != nil {
				return err
			}

			if len(patients) == 0 {
				fmt.Println("No patients found.")
				return nil
			}

			for _, p := range patients {
				fmt.Printf("%s  %s %s  (%s)\n", p.ID, p.FirstName, p.LastName, p.PersonalNumber)
				fmt.Printf("  Born: %s  Gender: %s  Nationality: %s\n", p.BirthDate, p.Gender, p.Nationality)
				fmt.Printf("  Address: %s, %s %s\n", p.Address.StreetName, p.Address.ZipCode, p.Address.City)
				if p.Email != "" {
					fmt.Printf("  Email: %s\n", p.Email)
				}
				if p.MobilePhoneNumber != "" {
					fmt.Printf("  Mobile: %s\n", p.MobilePhoneNumber)
				}
				fmt.Printf("  Patient type: %s (%s)\n", p.PatientType.Name, p.PatientType.Type)
				if p.Organization != nil {
					fmt.Printf("  Organization: %s\n", p.Organization.Name)
				}
			}

			return nil
		},
	}

	patientsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new patient",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := httpclient.FromConfig(apiURLFlag)
			if err != nil {
				return err
			}

			personalNumber, _ := cmd.Flags().GetString("personal-number")
			clinicName, _ := cmd.Flags().GetString("clinic-name")
			clinicHsaID, _ := cmd.Flags().GetString("clinic-hsa-id")
			countyName, _ := cmd.Flags().GetString("county-name")
			patientType, _ := cmd.Flags().GetString("patient-type")
			email, _ := cmd.Flags().GetString("email")
			mobile, _ := cmd.Flags().GetString("mobile")

			req := api.CreatePatientRequest{
				PersonalNumber: personalNumber,
				ListInfo: api.CreatePatientListInfo{
					Name:       clinicName,
					HsaID:      clinicHsaID,
					CountyName: countyName,
				},
				PatientType:  patientType,
				Email:        email,
				MobileNumber: mobile,
			}

			patient, err := api.CreatePatient(client, req)
			if err != nil {
				return err
			}

			fmt.Printf("Patient created: %s\n", patient.ID)
			fmt.Printf("  %s %s  (%s)\n", patient.FirstName, patient.LastName, patient.PersonalNumber)
			fmt.Printf("  Patient type: %s\n", patient.PatientType)

			return nil
		},
	}

	patientsCreateCmd.Flags().String("personal-number", "", "Personal number (required)")
	patientsCreateCmd.Flags().String("clinic-name", "", "Listed clinic name (required)")
	patientsCreateCmd.Flags().String("clinic-hsa-id", "", "Listed clinic HSA ID (required)")
	patientsCreateCmd.Flags().String("county-name", "", "County name (required)")
	patientsCreateCmd.Flags().String("patient-type", "", "Patient type name, e.g. Private (required)")
	patientsCreateCmd.Flags().String("email", "", "Email address")
	patientsCreateCmd.Flags().String("mobile", "", "Mobile number")
	patientsCreateCmd.MarkFlagRequired("personal-number")
	patientsCreateCmd.MarkFlagRequired("clinic-name")
	patientsCreateCmd.MarkFlagRequired("clinic-hsa-id")
	patientsCreateCmd.MarkFlagRequired("county-name")
	patientsCreateCmd.MarkFlagRequired("patient-type")

	// -- patient-types ------------------------------------------------------

	patientTypesCmd := &cobra.Command{
		Use:   "patient-types",
		Short: "List and inspect patient types",
	}

	patientTypesListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all patient types",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := httpclient.FromConfig(apiURLFlag)
			if err != nil {
				return err
			}

			name, _ := cmd.Flags().GetString("name")
			patientType, _ := cmd.Flags().GetString("type")

			types, err := api.ListPatientTypes(client, name, patientType)
			if err != nil {
				return err
			}

			if len(types) == 0 {
				fmt.Println("No patient types found.")
				return nil
			}

			for _, t := range types {
				fmt.Printf("%-4s  %-30s  %s\n", t.ID, t.Name, t.Type)
			}

			return nil
		},
	}
	patientTypesListCmd.Flags().String("name", "", "Filter by name")
	patientTypesListCmd.Flags().String("type", "", "Filter by type (e.g. private, insurance, healthcare_contract)")

	patientTypesGetCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a patient type by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := httpclient.FromConfig(apiURLFlag)
			if err != nil {
				return err
			}

			t, err := api.GetPatientType(client, args[0])
			if err != nil {
				return err
			}

			fmt.Printf("%-4s  %-30s  %s\n", t.ID, t.Name, t.Type)

			return nil
		},
	}

	// -- assemble tree ------------------------------------------------------
	authCmd.AddCommand(loginCmd, authStatusCmd)
	configCmd.AddCommand(setAuthURLCmd, setAPIURLCmd, showConfigCmd)
	bookingTypesCmd.AddCommand(bookingTypesList)
	usersCmd.AddCommand(usersSearchCmd)
	patientsCmd.AddCommand(patientsSearchCmd, patientsCreateCmd)
	patientTypesCmd.AddCommand(patientTypesListCmd, patientTypesGetCmd)

	rootCmd.AddCommand(
		authCmd,
		configCmd,
		bookingTypesCmd,
		usersCmd,
		patientsCmd,
		patientTypesCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
