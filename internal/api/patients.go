package api

import (
	"fmt"
	"net/url"

	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

func GetPatients(c *httpclient.Client, personalNumber string) ([]Patient, error) {
	params := url.Values{}
	params.Set("personalNumber", personalNumber)
	return httpclient.Get[[]Patient](c, "/v2/patients", params)
}

func CreatePatient(c *httpclient.Client, req CreatePatientRequest) (CreatedPatient, error) {
	patients, err := httpclient.Post[[]CreatedPatient](c, "/v1/patients", req)
	if err != nil {
		return CreatedPatient{}, err
	}
	if len(patients) == 0 {
		return CreatedPatient{}, fmt.Errorf("patient created but no data returned")
	}
	return patients[0], nil
}
