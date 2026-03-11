package api

import (
	"fmt"
	"net/url"

	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

type PatientTypeResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func ListPatientTypes(c *httpclient.Client, name, patientType string) ([]PatientTypeResult, error) {
	params := url.Values{}
	if name != "" {
		params.Set("name", name)
	}
	if patientType != "" {
		params.Set("type", patientType)
	}
	return httpclient.Get[[]PatientTypeResult](c, "/v1/patientTypes", params)
}

func GetPatientType(c *httpclient.Client, id string) (PatientTypeResult, error) {
	return httpclient.Get[PatientTypeResult](c, fmt.Sprintf("/v1/patientTypes/%s", id), nil)
}
