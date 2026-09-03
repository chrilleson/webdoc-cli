package api

import (
	"net/url"

	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

type Clinic struct {
	ID                 string      `json:"id"`
	Name               string      `json:"name"`
	HsaID              string      `json:"hsaId"`
	CompanyHsaID       string      `json:"companyHsaId"`
	OrganisationNumber string      `json:"organisationNumber"`
	Address            string      `json:"address"`
	PostNumber         string      `json:"postNumber"`
	PostAddress        string      `json:"postAddress"`
	InvoiceInfo        InvoiceInfo `json:"invoiceInfo"`
}

func ListClinics(c *httpclient.Client, params url.Values) ([]Clinic, error) {
	return httpclient.Get[[]Clinic](c, "/v1/clinics", params)
}
