package api

import (
	"fmt"

	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

type Department struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Parent string `json:"parent"`
}

type Organization struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Active      int          `json:"active"`
	Departments []Department `json:"departments"`
}

// OrganizationRow is one row of the create batch. level_1 is the company and
// level_2..7 are nested sub-departments; unused levels are sent as null, so the
// pointers must not carry omitempty.
type OrganizationRow struct {
	Level1               *string `json:"level_1"`
	Level2               *string `json:"level_2"`
	Level3               *string `json:"level_3"`
	Level4               *string `json:"level_4"`
	Level5               *string `json:"level_5"`
	Level6               *string `json:"level_6"`
	Level7               *string `json:"level_7"`
	OrganizationNumber   string  `json:"organizationNumber"`
	CompanyText          string  `json:"companyText"`
	Email                string  `json:"email"`
	StreetAddress        string  `json:"streetAddress"`
	ZipCode              string  `json:"zipCode"`
	City                 string  `json:"city"`
	InvoiceEmail         string  `json:"invoiceEmail"`
	InvoiceStreetAddress string  `json:"invoiceStreetAddress"`
	InvoiceZipCode       string  `json:"invoiceZipCode"`
	InvoiceCity          string  `json:"invoiceCity"`
}

type CreateOrganizationsRequest struct {
	Rows []OrganizationRow `json:"rows"`
}

type CreateOrganizationsResponse struct {
	UUID string `json:"uuid"`
}

type AddPatientToOrganizationRequest struct {
	PatientID    string `json:"patientId"`
	DepartmentID string `json:"departmentId"`
}

func ListOrganizations(c *httpclient.Client) ([]Organization, error) {
	return httpclient.Get[[]Organization](c, "/v1/organizations", nil)
}

// CreateOrganizations posts the batch. Note the singular /v1/organization/
// prefix - that is the real path, not a typo.
func CreateOrganizations(c *httpclient.Client, req CreateOrganizationsRequest) (CreateOrganizationsResponse, error) {
	return httpclient.Post[CreateOrganizationsResponse](c, "/v1/organization/organizations", req)
}

func AddPatientToOrganization(c *httpclient.Client, organizationID string, req AddPatientToOrganizationRequest) error {
	_, err := httpclient.Put[struct{}](c, fmt.Sprintf("/v1/organizations/%s/patient", organizationID), req)
	return err
}
