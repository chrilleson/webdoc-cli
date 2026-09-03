package commands

import (
	"fmt"
	"strings"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewDocumentTypesCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "document-types",
		Short: "List document types",
	}
	cmd.AddCommand(newDocumentTypesListCmd(o))
	return cmd
}

func newDocumentTypesListCmd(o *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all document types",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			documentTypes, err := api.ListDocumentTypes(client)
			if err != nil {
				return ScopeHint(err, "document-types:read")
			}

			return EmitEmpty(o, documentTypes, len(documentTypes) == 0, "No document types found.", func() {
				for _, dt := range documentTypes {
					var flags []string
					if !dt.Active {
						flags = append(flags, "[inactive]")
					}
					if dt.SignRequired {
						flags = append(flags, "[sign required]")
					}
					if dt.InReferral {
						flags = append(flags, "[in-referral]")
					}
					fmt.Printf("%-4d  %-40s  %s\n", dt.ID, dt.Name, strings.Join(flags, " "))
				}
			})
		},
	}
}
