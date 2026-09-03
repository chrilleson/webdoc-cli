package api

import (
	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

type CreateNoteRequest struct {
	PatientID        string `json:"patientId"`
	ClinicID         string `json:"clinicId"`
	PatientTypeID    int    `json:"patientTypeId"`
	UserID           string `json:"userId"`
	RecordTemplateID int    `json:"recordTemplateId"`
	CreatedAt        string `json:"createdAt"`
}

type Note struct {
	ID               string `json:"id"`
	PatientID        string `json:"patientId"`
	ClinicID         string `json:"clinicId"`
	PatientTypeID    int    `json:"patientTypeId"`
	UserID           string `json:"userId"`
	RecordTemplateID int    `json:"recordTemplateId"`
	CreatedAt        string `json:"createdAt"`
}

func CreateNote(c *httpclient.Client, req CreateNoteRequest) (Note, error) {
	return httpclient.Post[Note](c, "/v1/notes", req)
}
