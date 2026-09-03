package commands

import (
	"fmt"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewPaymentMethodsCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment-methods",
		Short: "List clinic payment methods",
	}
	cmd.AddCommand(newPaymentMethodsListCmd(o))
	return cmd
}

func newPaymentMethodsListCmd(o *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the payment methods of a clinic",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}
			clinicID, err := o.Clinic()
			if err != nil {
				return err
			}

			methods, err := api.ListPaymentMethods(client, clinicID)
			if err != nil {
				return err
			}

			return EmitEmpty(o, methods, len(methods) == 0, "No payment methods found.", func() {
				for _, m := range methods {
					fmt.Printf("%-4s  %-30s  %s", m.ID, m.Name, m.AssetsAccount)
					if m.DirectPayment {
						fmt.Print(" [direct]")
					}
					if !m.Active {
						fmt.Print(" [inactive]")
					}
					fmt.Println()
				}
			})
		},
	}
}
