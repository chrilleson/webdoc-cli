package main

import (
	"os"

	"github.com/chrilleson/webdoc-cli/internal/commands"
	"github.com/spf13/cobra"
)

func main() {
	opts := &commands.Options{}

	rootCmd := &cobra.Command{
		Use:           "webdoc",
		Short:         "CLI for the Webdoc EMR API",
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	// Persistent flags available on ALL subcommands.
	rootCmd.PersistentFlags().StringVar(&opts.APIURL, "url", "", "Override API base URL")
	rootCmd.PersistentFlags().StringVar(&opts.ClinicID, "clinic-id", "", "Clinic ID for clinic-scoped commands (default: config clinic_id)")
	rootCmd.PersistentFlags().BoolVar(&opts.JSON, "json", false, "Print the raw API response as JSON")

	// One entry per resource. Keep alphabetical; each constructor lives in its
	// own file under internal/commands/.
	rootCmd.AddCommand(
		commands.NewActionCodesCmd(opts),
		commands.NewAuthCmd(opts),
		commands.NewBookingTypesCmd(opts),
		commands.NewBookingsCmd(opts),
		commands.NewClinicsCmd(opts),
		commands.NewConfigCmd(opts),
		commands.NewDocumentTypesCmd(opts),
		commands.NewDocumentsCmd(opts),
		commands.NewHealthCmd(opts),
		commands.NewInvoicesCmd(opts),
		commands.NewNotesCmd(opts),
		commands.NewOrganizationsCmd(opts),
		commands.NewPatientTypesCmd(opts),
		commands.NewPatientsCmd(opts),
		commands.NewPaymentMethodsCmd(opts),
		commands.NewRecordTemplatesCmd(opts),
		commands.NewRecordsCmd(opts),
		commands.NewUsersCmd(opts),
		commands.NewVisitsCmd(opts),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
