package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewRecordsCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "records",
		Short: "Manage visit records",
	}
	cmd.AddCommand(
		newRecordsUpdateCmd(o),
		newRecordsCreateExternalCmd(o),
	)
	return cmd
}

// parseKeywords turns repeated --keyword id=value flags into the payload shape.
// Only the first '=' separates the pair, so values may contain '='. Get valid
// ids from `webdoc record-templates keywords <templateId>`.
func parseKeywords(values []string) ([]api.Keyword, error) {
	keywords := []api.Keyword{}
	for _, raw := range values {
		id, value, found := strings.Cut(raw, "=")
		if !found {
			return nil, fmt.Errorf("--keyword must be id=value, got %q", raw)
		}
		n, err := strconv.Atoi(strings.TrimSpace(id))
		if err != nil {
			return nil, fmt.Errorf("--keyword id must be an integer, got %q", id)
		}
		keywords = append(keywords, api.Keyword{ID: n, Value: value})
	}
	return keywords, nil
}

func codeList(codes []api.CodeRef) string {
	out := make([]string, 0, len(codes))
	for _, c := range codes {
		out = append(out, c.Code)
	}
	return strings.Join(out, ", ")
}

func newRecordsUpdateCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <visitId>",
		Short: "Write keyword values onto a visit's record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			userID, _ := cmd.Flags().GetString("user-id")
			rawKeywords, _ := cmd.Flags().GetStringArray("keyword")
			keywords, err := parseKeywords(rawKeywords)
			if err != nil {
				return err
			}

			diagnosisCodes, _ := cmd.Flags().GetStringArray("diagnosis-code")
			if diagnosisCodes == nil {
				diagnosisCodes = []string{}
			}
			actionCodes, _ := cmd.Flags().GetStringArray("action-code")
			if actionCodes == nil {
				actionCodes = []string{}
			}

			record, err := api.UpdateRecord(client, args[0], api.UpdateRecordRequest{
				UserID:         userID,
				Keywords:       keywords,
				DiagnosisCodes: diagnosisCodes,
				ActionCodes:    actionCodes,
			})
			if err != nil {
				return ScopeHint(err, "patient:write")
			}

			return Emit(o, record, func() {
				fmt.Printf("Record updated for visit %s\n", args[0])
				fmt.Printf("  Keywords: %d\n", len(record.Keywords))
				fmt.Printf("  Diagnoses: %s\n", codeList(record.Diagnoses))
				fmt.Printf("  Actions: %s\n", codeList(record.Actions))
			})
		},
	}

	cmd.Flags().String("user-id", "", "User ID authoring the record (required)")
	cmd.Flags().StringArray("keyword", nil, "Keyword value as id=value (repeatable)")
	cmd.Flags().StringArray("diagnosis-code", nil, "Diagnosis code, e.g. W02.03 (repeatable)")
	cmd.Flags().StringArray("action-code", nil, "Action code, e.g. PB003 (repeatable)")
	cmd.MarkFlagRequired("user-id")

	return cmd
}

func newRecordsCreateExternalCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-external",
		Short: "Create an external record for a patient",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}

			patientIdentity, _ := cmd.Flags().GetString("patient-identity")
			authorIdentity, _ := cmd.Flags().GetInt("author-identity")
			recordTemplateID, _ := cmd.Flags().GetString("record-template-id")
			source, _ := cmd.Flags().GetString("source")
			questionnaireName, _ := cmd.Flags().GetString("questionnaire-name")

			rawKeywords, _ := cmd.Flags().GetStringArray("keyword")
			keywords, err := parseKeywords(rawKeywords)
			if err != nil {
				return err
			}

			diagnosisCodes, _ := cmd.Flags().GetStringArray("diagnosis-code")
			if diagnosisCodes == nil {
				diagnosisCodes = []string{}
			}
			actionCodes, _ := cmd.Flags().GetStringArray("action-code")
			if actionCodes == nil {
				actionCodes = []string{}
			}

			record, err := api.CreateExternalRecord(client, api.CreateExternalRecordRequest{
				PatientIdentity:   patientIdentity,
				AuthorIdentity:    authorIdentity,
				RecordTemplateID:  recordTemplateID,
				Source:            source,
				QuestionnaireName: questionnaireName,
				Keywords:          keywords,
				ActionCodes:       actionCodes,
				DiagnosisCodes:    diagnosisCodes,
			})
			if err != nil {
				return ScopeHint(err, "system-admin")
			}

			return Emit(o, record, func() {
				fmt.Printf("External record created: %s\n", record.UUID)
				fmt.Printf("  Patient: %s\n", record.PatientIdentity)
				fmt.Printf("  Template: %s\n", record.RecordTemplateID)
				if record.QuestionnaireName != "" {
					fmt.Printf("  Questionnaire: %s\n", record.QuestionnaireName)
				}
				if record.Source != "" {
					fmt.Printf("  Source: %s\n", record.Source)
				}
				fmt.Printf("  Keywords: %d\n", len(record.Keywords))
				fmt.Printf("  Diagnoses: %s\n", strings.Join(record.DiagnosisCodes, ", "))
				fmt.Printf("  Actions: %s\n", strings.Join(record.ActionCodes, ", "))
			})
		},
	}

	cmd.Flags().String("patient-identity", "", "Patient UUID (required)")
	cmd.Flags().Int("author-identity", 0, "Author identity (required)")
	cmd.Flags().String("record-template-id", "", "Record template ID (required)")
	cmd.Flags().String("source", "", "Source system name")
	cmd.Flags().String("questionnaire-name", "", "Questionnaire name")
	cmd.Flags().StringArray("keyword", nil, "Keyword value as id=value (repeatable)")
	cmd.Flags().StringArray("diagnosis-code", nil, "Diagnosis code, e.g. W00.00 (repeatable)")
	cmd.Flags().StringArray("action-code", nil, "Action code, e.g. AA016 (repeatable)")
	cmd.MarkFlagRequired("patient-identity")
	cmd.MarkFlagRequired("author-identity")
	cmd.MarkFlagRequired("record-template-id")

	return cmd
}
