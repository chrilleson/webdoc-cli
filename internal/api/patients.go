package api

import (
	"net/url"

	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

type PatientAddress struct {
	StreetName string `json:"streetName"`
	CoAddress  string `json:"coAddress"`
	City       string `json:"city"`
	ZipCode    string `json:"zipCode"`
}

type ListedClinic struct {
	HsaID string `json:"hsaId"`
}

type PatientOrganization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PatientDepartment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type County struct {
	Code string `json:"code"`
}

type PatientType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Patient struct {
	ID                  string               `json:"id"`
	PersonalNumber      string               `json:"personalNumber"`
	BirthDate           string               `json:"birthDate"`
	Gender              string               `json:"gender"`
	FirstName           string               `json:"firstName"`
	LastName            string               `json:"lastName"`
	Address             PatientAddress       `json:"address"`
	Nationality         string               `json:"nationality"`
	Email               string               `json:"email"`
	MobilePhoneNumber   string               `json:"mobilePhoneNumber"`
	WorkPhoneNumber     string               `json:"workPhoneNumber"`
	HomePhoneNumber     string               `json:"homePhoneNumber"`
	ListedClinic        ListedClinic         `json:"listedClinic"`
	InterpreterNeeded   bool                 `json:"interpreterNeeded"`
	InterpreterLanguage string               `json:"interpreterLanguage"`
	Organization        *PatientOrganization `json:"organization"`
	Department          *PatientDepartment   `json:"department"`
	County              County               `json:"county"`
	PatientType         PatientType          `json:"patientType"`
}

func GetPatients(c *httpclient.Client, personalNumber string) ([]Patient, error) {
	params := url.Values{}
	params.Set("personalNumber", personalNumber)
	return httpclient.Get[[]Patient](c, "/v2/patients", params)
}
