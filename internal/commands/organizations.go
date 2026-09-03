package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewOrganizationsCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "organizations",
		Short: "Manage organizations",
	}
	cmd.AddCommand(
		newOrganizationsListCmd(o),
		newOrganizationsCreateCmd(o),
		newOrganizationsAddPatientCmd(o),
	)
	return cmd
}

func newOrganizationsListCmd(o *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List organizations and their departments",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			orgs, err := api.ListOrganizations(client)
			if err != nil {
				return ScopeHint(err, "organization:read")
			}

			return EmitEmpty(o, orgs, len(orgs) == 0, "No organizations found.", func() {
				for _, org := range orgs {
					fmt.Printf("%s  %-40s", org.ID, org.Name)
					if org.Active != 1 {
						fmt.Print(" [inactive]")
					}
					fmt.Println()
					for _, d := range org.Departments {
						fmt.Printf("  └─ %s  %s\n", d.ID, d.Name)
					}
				}
			})
		},
	}
}

func newOrganizationsCreateCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create organizations from a batch of rows",
		Long: "Create one or more organizations.\n\n" +
			"The API takes a batch of rows, each naming a path through the org hierarchy\n" +
			"in level_1 (the company) down to level_7 (the deepest sub-department), plus\n" +
			"address and invoicing details. Because that is far too many values for flags,\n" +
			"pass them in a JSON file with --file. The file may be either a bare array of\n" +
			"row objects or a full {\"rows\": [...]} object.\n\n" +
			"Example row:\n" +
			"  {\n" +
			"    \"level_1\": \"Foretag AB\",\n" +
			"    \"level_2\": \"Underavdelning\",\n" +
			"    \"organizationNumber\": \"556896-8001\",\n" +
			"    \"streetAddress\": \"Gatan 1\", \"zipCode\": \"123 45\", \"city\": \"Goteborg\",\n" +
			"    \"invoiceStreetAddress\": \"Box 123\", \"invoiceZipCode\": \"123 45\",\n" +
			"    \"invoiceCity\": \"Goteborg\", \"invoiceEmail\": \"billing@example.com\"\n" +
			"  }\n\n" +
			"For the trivial single-row case use --level-1 instead of a file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			file, _ := cmd.Flags().GetString("file")
			level1, _ := cmd.Flags().GetString("level-1")

			var req api.CreateOrganizationsRequest
			switch {
			case file != "":
				data, err := os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("reading %s: %w", file, err)
				}
				if err := json.Unmarshal(data, &req); err != nil {
					// Fall back to a bare array of rows.
					var rows []api.OrganizationRow
					if err2 := json.Unmarshal(data, &rows); err2 != nil {
						return fmt.Errorf("parsing %s: expected a JSON array of rows or {\"rows\": [...]}: %w", file, err)
					}
					req.Rows = rows
				}
				if len(req.Rows) == 0 {
					return fmt.Errorf("%s contains no rows", file)
				}
			case level1 != "":
				name := level1
				req.Rows = []api.OrganizationRow{{Level1: &name}}
			default:
				return fmt.Errorf("provide --file <path> or --level-1 <name>")
			}

			client, err := o.Client()
			if err != nil {
				return err
			}

			created, err := api.CreateOrganizations(client, req)
			if err != nil {
				return ScopeHint(err, "organization:write")
			}

			return Emit(o, created, func() {
				fmt.Printf("Organizations created: %s\n", created.UUID)
				fmt.Printf("  %d rows\n", len(req.Rows))
			})
		},
	}
	cmd.Flags().String("file", "", "JSON file of organization rows")
	cmd.Flags().String("level-1", "", "Company name, for the single-row case")

	return cmd
}

func newOrganizationsAddPatientCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-patient <organizationId>",
		Short: "Add a patient to an organization department",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			patientID, _ := cmd.Flags().GetString("patient-id")
			departmentID, _ := cmd.Flags().GetString("department-id")

			if err := api.AddPatientToOrganization(client, args[0], api.AddPatientToOrganizationRequest{
				PatientID:    patientID,
				DepartmentID: departmentID,
			}); err != nil {
				return ScopeHint(err, "organization:write")
			}

			result := map[string]string{
				"organizationId": args[0],
				"patientId":      patientID,
				"departmentId":   departmentID,
			}
			return Emit(o, result, func() {
				fmt.Printf("Patient %s added to department %s\n", patientID, departmentID)
			})
		},
	}
	cmd.Flags().String("patient-id", "", "Patient ID (required)")
	cmd.Flags().String("department-id", "", "Department ID (required)")
	cmd.MarkFlagRequired("patient-id")
	cmd.MarkFlagRequired("department-id")

	return cmd
}
