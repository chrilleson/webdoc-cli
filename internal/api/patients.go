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

// ListPatientsV1 calls the deprecated GET /v1/patients. Despite the plural
// path it fetches one patient and returns a single object.
func ListPatientsV1(c *httpclient.Client, personalNumber string) (PatientV1, error) {
	params := url.Values{}
	if personalNumber != "" {
		params.Set("personalNumber", personalNumber)
	}
	return httpclient.Get[PatientV1](c, "/v1/patients", params)
}

// UpdatePatient patches a patient. The endpoint returns no body.
func UpdatePatient(c *httpclient.Client, personalNumber string, req UpdatePatientRequest) error {
	_, err := httpclient.Patch[struct{}](c, fmt.Sprintf("/v1/patients/%s", personalNumber), req)
	return err
}

// CreatePatientWithoutSSN posts a reserve-number/foreign patient. The response
// is a single-element array.
func CreatePatientWithoutSSN(c *httpclient.Client, req CreatePatientWithoutSSNRequest) (PatientWithoutSSN, error) {
	patients, err := httpclient.Post[[]PatientWithoutSSN](c, "/v1/patient/withoutSocialSecurityNumber", req)
	if err != nil {
		return PatientWithoutSSN{}, err
	}
	if len(patients) == 0 {
		return PatientWithoutSSN{}, fmt.Errorf("patient created but no data returned")
	}
	return patients[0], nil
}

func GetNationalRegistryPerson(c *httpclient.Client, personalNumber string) (NationalRegistryPerson, error) {
	return httpclient.Get[NationalRegistryPerson](c, fmt.Sprintf("/v1/nationalRegistry/%s", personalNumber), nil)
}

// SetCurrentPatient switches the token user's current patient. No response body.
func SetCurrentPatient(c *httpclient.Client, req SetCurrentPatientRequest) error {
	_, err := httpclient.Put[struct{}](c, "/v1/currentUser/currentPatient", req)
	return err
}
