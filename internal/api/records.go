package api

import (
	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

// UpdateRecordRequest writes keyword values onto a visit's record. Keywords,
// DiagnosisCodes and ActionCodes must serialise as [] rather than null.
type UpdateRecordRequest struct {
	UserID         string    `json:"userId"`
	Keywords       []Keyword `json:"keywords"`
	DiagnosisCodes []string  `json:"diagnosisCodes"`
	ActionCodes    []string  `json:"actionCodes"`
}

type CodeRef struct {
	Code string `json:"code"`
}

// UpdatedRecord is the PATCH response. Note the asymmetry with the request:
// diagnosisCodes/actionCodes go out as strings, diagnoses/actions come back as
// objects with a `code` field.
type UpdatedRecord struct {
	UserID    string    `json:"userId"`
	Keywords  []Keyword `json:"keywords"`
	Diagnoses []CodeRef `json:"diagnoses"`
	Actions   []CodeRef `json:"actions"`
}

type CreateExternalRecordRequest struct {
	PatientIdentity   string    `json:"patientIdentity"`
	AuthorIdentity    int       `json:"authorIdentity"`
	RecordTemplateID  string    `json:"recordTemplateId"`
	Source            string    `json:"source"`
	QuestionnaireName string    `json:"questionnaireName"`
	Keywords          []Keyword `json:"keywords"`
	ActionCodes       []string  `json:"actionCodes"`
	DiagnosisCodes    []string  `json:"diagnosisCodes"`
}

type ExternalRecordKeyword struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Value string `json:"value"`
	Type  int    `json:"type"`
}

// ExternalRecord is the POST /v1/externalRecords response. recordTemplateId is
// a string going out and a number coming back.
type ExternalRecord struct {
	UUID              string                  `json:"uuid"`
	PatientIdentity   string                  `json:"patientIdentity"`
	AuthorIdentity    int                     `json:"authorIdentity"`
	RecordTemplateID  Loose                   `json:"recordTemplateId"`
	QuestionnaireName string                  `json:"questionnaireName"`
	Source            string                  `json:"source"`
	Keywords          []ExternalRecordKeyword `json:"keywords"`
	ActionCodes       []string                `json:"actionCodes"`
	DiagnosisCodes    []string                `json:"diagnosisCodes"`
}

func UpdateRecord(c *httpclient.Client, visitID string, req UpdateRecordRequest) (UpdatedRecord, error) {
	return httpclient.Patch[UpdatedRecord](c, "/v1/visits/"+visitID+"/records", req)
}

func CreateExternalRecord(c *httpclient.Client, req CreateExternalRecordRequest) (ExternalRecord, error) {
	return httpclient.Post[ExternalRecord](c, "/v1/externalRecords", req)
}
