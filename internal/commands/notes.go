package commands

import (
	"fmt"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewNotesCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notes",
		Short: "Create notes",
	}
	cmd.AddCommand(newNotesCreateCmd(o))
	return cmd
}

func newNotesCreateCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a note for a patient",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}
			clinicID, err := o.Clinic()
			if err != nil {
				return err
			}

			patientID, _ := cmd.Flags().GetString("patient-id")
			patientTypeID, _ := cmd.Flags().GetInt("patient-type-id")
			userID, _ := cmd.Flags().GetString("user-id")
			recordTemplateID, _ := cmd.Flags().GetInt("record-template-id")

			createdAt, _ := cmd.Flags().GetString("created-at")
			if createdAt, err = ValidateDateTime("created-at", createdAt); err != nil {
				return err
			}

			note, err := api.CreateNote(client, api.CreateNoteRequest{
				PatientID:        patientID,
				ClinicID:         clinicID,
				PatientTypeID:    patientTypeID,
				UserID:           userID,
				RecordTemplateID: recordTemplateID,
				CreatedAt:        createdAt,
			})
			if err != nil {
				return ScopeHint(err, "patient:write")
			}

			return Emit(o, note, func() {
				fmt.Printf("Note created: %s\n", note.ID)
				fmt.Printf("  Patient: %s\n", note.PatientID)
				fmt.Printf("  Template: %d\n", note.RecordTemplateID)
				fmt.Printf("  Created: %s\n", note.CreatedAt)
			})
		},
	}

	cmd.Flags().String("patient-id", "", "Patient ID (required)")
	cmd.Flags().Int("patient-type-id", 0, "Patient type ID (required)")
	cmd.Flags().String("user-id", "", "Authoring user ID (required)")
	cmd.Flags().Int("record-template-id", 0, "Record template ID (required)")
	cmd.Flags().String("created-at", "", "Creation time, YYYY-MM-DD HH:MM:SS")
	cmd.MarkFlagRequired("patient-id")
	cmd.MarkFlagRequired("patient-type-id")
	cmd.MarkFlagRequired("user-id")
	cmd.MarkFlagRequired("record-template-id")

	return cmd
}
