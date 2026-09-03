package commands

import (
	"fmt"
	"path/filepath"

	"github.com/chrilleson/webdoc-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewDocumentsCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "documents",
		Short: "Upload clinic documents",
	}
	cmd.AddCommand(newDocumentsUploadCmd(o))
	return cmd
}

func newDocumentsUploadCmd(o *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a document to a clinic",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := o.Client()
			if err != nil {
				return err
			}
			clinicID, err := o.Clinic()
			if err != nil {
				return err
			}

			file, _ := cmd.Flags().GetString("file")
			fileName, _ := cmd.Flags().GetString("file-name")
			documentTypeID, _ := cmd.Flags().GetString("document-type-id")
			personalNumber, _ := cmd.Flags().GetString("personal-number")
			userID, _ := cmd.Flags().GetString("user-id")

			createdAt, _ := cmd.Flags().GetString("created-at")
			if createdAt, err = ValidateDateTime("created-at", createdAt); err != nil {
				return err
			}

			if fileName == "" {
				fileName = filepath.Base(file)
			}

			req := api.UploadDocumentRequest{
				FilePath:       file,
				FileName:       fileName,
				DocumentTypeID: documentTypeID,
				PersonalNumber: personalNumber,
				UserID:         userID,
				CreatedAt:      createdAt,
			}

			if err := api.UploadDocument(client, clinicID, req); err != nil {
				return ScopeHint(err, "documents:write")
			}

			result := map[string]string{"fileName": fileName, "clinicId": clinicID}
			return Emit(o, result, func() {
				fmt.Printf("Uploaded %s to clinic %s\n", fileName, clinicID)
			})
		},
	}

	cmd.Flags().String("file", "", "Path to the file to upload (required)")
	cmd.Flags().String("file-name", "", "Name to store the file as (default: the file's basename)")
	cmd.Flags().String("document-type-id", "", "Document type ID from document-types list (required)")
	cmd.Flags().String("personal-number", "", "Patient personal number (required)")
	cmd.Flags().String("user-id", "", "Uploading user ID (required)")
	cmd.Flags().String("created-at", "", "Creation time, YYYY-MM-DD HH:MM:SS")
	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("document-type-id")
	cmd.MarkFlagRequired("personal-number")
	cmd.MarkFlagRequired("user-id")

	return cmd
}
