package commands

import (
	"fmt"
	"sort"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

// NewRecordTemplatesCmd owns the record-templates group. Both endpoints need the
// system-admin scope, so their errors are wrapped in ScopeHint.
func NewRecordTemplatesCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record-templates",
		Short: "List and inspect record templates",
	}
	cmd.AddCommand(
		newRecordTemplatesListCmd(o),
		newRecordTemplatesGetCmd(o),
		newRecordTemplatesKeywordsCmd(o),
	)
	return cmd
}

// sortedKeywords returns a copy ordered by visual.order, leaving the API's own
// order intact for --json.
func sortedKeywords(keywords []api.TemplateKeyword) []api.TemplateKeyword {
	sorted := make([]api.TemplateKeyword, len(keywords))
	copy(sorted, keywords)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Visual.Order < sorted[j].Visual.Order
	})
	return sorted
}

func newRecordTemplatesListCmd(o *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all record templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			templates, err := api.ListRecordTemplates(client)
			if err != nil {
				return ScopeHint(err, "system-admin")
			}

			return EmitEmpty(o, templates, len(templates) == 0, "No record templates found.", func() {
				for _, t := range templates {
					fmt.Printf("%-6d  %-40s  type=%d  %d keywords\n", t.ID, t.Title, t.TemplateType, len(t.Keywords))
				}
			})
		},
	}
}

func newRecordTemplatesGetCmd(o *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <templateId>",
		Short: "Show a record template and its keywords",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			template, err := api.GetRecordTemplate(client, args[0])
			if err != nil {
				return ScopeHint(err, "system-admin")
			}

			return Emit(o, template, func() {
				fmt.Printf("%d  %s\n", template.ID, template.Title)
				fmt.Printf("  Template type: %d\n", template.TemplateType)
				fmt.Printf("  Keywords: %d\n", len(template.Keywords))
				for _, k := range sortedKeywords(template.Keywords) {
					fmt.Printf("  └─ %-5d %-30s type=%d\n", k.ID, k.Title, k.Type)
				}
			})
		},
	}
}

func newRecordTemplatesKeywordsCmd(o *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "keywords <templateId>",
		Short: "List a template's keyword IDs for building record payloads",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			template, err := api.GetRecordTemplate(client, args[0])
			if err != nil {
				return ScopeHint(err, "system-admin")
			}

			keywords := template.Keywords
			return EmitEmpty(o, keywords, len(keywords) == 0, "No keywords found.", func() {
				fmt.Printf("%-6s  %-30s  %-6s  %-10s  %-12s\n", "ID", "TITLE", "TYPE", "MAXLENGTH", "DEFAULT")
				for _, k := range sortedKeywords(keywords) {
					marker := ""
					if k.WarnIfEmpty {
						marker = "*"
					}
					fmt.Printf("%-6d  %-30s  %-6d  %-10s  %-12s %s\n", k.ID, k.Title, k.Type, k.MaxLength, k.DefaultValue, marker)
				}
				fmt.Println("* = warn if empty")
			})
		},
	}
}
