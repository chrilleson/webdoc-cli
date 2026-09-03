package commands

import (
	"fmt"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewPatientTypesCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "patient-types",
		Short: "List and inspect patient types",
	}
	cmd.AddCommand(newPatientTypesListCmd(o), newPatientTypesGetCmd(o))
	return cmd
}

func newPatientTypesListCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all patient types",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			name, _ := cmd.Flags().GetString("name")
			patientType, _ := cmd.Flags().GetString("type")

			types, err := api.ListPatientTypes(client, name, patientType)
			if err != nil {
				return err
			}

			return EmitEmpty(o, types, len(types) == 0, "No patient types found.", func() {
				for _, t := range types {
					fmt.Printf("%-4s  %-30s  %s\n", t.ID, t.Name, t.Type)
				}
			})
		},
	}
	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("type", "", "Filter by type (e.g. private, insurance, healthcare_contract)")

	return cmd
}

func newPatientTypesGetCmd(o *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a patient type by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			t, err := api.GetPatientType(client, args[0])
			if err != nil {
				return err
			}

			return Emit(o, t, func() {
				fmt.Printf("%-4s  %-30s  %s\n", t.ID, t.Name, t.Type)
			})
		},
	}
}
