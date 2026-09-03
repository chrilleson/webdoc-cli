package commands

import (
	"fmt"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewBookingTypesCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "booking-types",
		Short: "Manage booking types",
	}
	cmd.AddCommand(newBookingTypesListCmd(o))
	return cmd
}

func newBookingTypesListCmd(o *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all booking types",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			bookingTypes, err := api.GetBookingTypes(client)
			if err != nil {
				return err
			}

			return EmitEmpty(o, bookingTypes, len(bookingTypes) == 0, "No booking types found", func() {
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
			})
		},
	}
}
