package commands

import (
	"fmt"
	"net/url"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewClinicsCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clinics",
		Short: "Look up clinics",
	}
	cmd.AddCommand(newClinicsListCmd(o))
	return cmd
}

func newClinicsListCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List clinics",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			hsaID, _ := cmd.Flags().GetString("hsa-id")
			companyHsaID, _ := cmd.Flags().GetString("company-hsa-id")
			organisationNumber, _ := cmd.Flags().GetString("organisation-number")

			q := url.Values{}
			SetStr(q, "hsaId", hsaID)
			SetStr(q, "companyHsaId", companyHsaID)
			SetStr(q, "organisationNumber", organisationNumber)

			clinics, err := api.ListClinics(client, q)
			if err != nil {
				return ScopeHint(err, "clinics:read")
			}

			return EmitEmpty(o, clinics, len(clinics) == 0, "No clinics found.", func() {
				for _, c := range clinics {
					fmt.Printf("%s  %-40s  %s\n", c.ID, c.Name, c.HsaID)
					fmt.Printf("  └─ %s, %s %s\n", c.Address, c.PostNumber, c.PostAddress)
				}
			})
		},
	}
	cmd.Flags().String("hsa-id", "", "Filter by clinic HSA ID")
	cmd.Flags().String("company-hsa-id", "", "Filter by company HSA ID")
	cmd.Flags().String("organisation-number", "", "Filter by organisation number")

	return cmd
}
