package commands

import (
	"fmt"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewActionCodesCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "action-codes",
		Short: "List action codes",
	}
	cmd.AddCommand(newActionCodesListCmd(o))
	return cmd
}

func newActionCodesListCmd(o *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all action codes",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			codes, err := api.ListActionCodes(client)
			if err != nil {
				return ScopeHint(err, "actioncodes:read")
			}

			return EmitEmpty(o, codes, len(codes) == 0, "No action codes found.", func() {
				for _, c := range codes {
					fmt.Printf("%-6d  %-10s  %-40s  fee=%s", c.ID, c.CodeName, c.CodeDescription, c.Fee)
					if c.Active != 1 {
						fmt.Print(" [inactive]")
					}
					fmt.Println()
				}
			})
		},
	}
}
