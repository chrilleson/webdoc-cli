package api

import (
	"net/url"

	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

type InvoiceInfo struct {
	CompanyName string `json:"companyName"`
}

type UserClinic struct {
	ID                     string      `json:"id"`
	Name                   string      `json:"name"`
	HsaID                  string      `json:"hsaId"`
	JournalArchiveClinicID *string     `json:"journalArchiveClinicId"`
	InvoiceInfo            InvoiceInfo `json:"invoiceInfo"`
}

type UserSettings struct {
	DefaultCostCentreID     *string `json:"defaultCostCentreId"`
	DefaultRecordTemplateID *string `json:"defaultRecordTemplateId"`
	DefaultPatientTypeID    *string `json:"defaultPatientTypeId"`
}

type User struct {
	ID             string       `json:"id"`
	FirstName      string       `json:"firstName"`
	LastName       string       `json:"lastName"`
	PersonalNumber string       `json:"personalNumber"`
	HsaID          string       `json:"hsaId"`
	Email          string       `json:"email"`
	PhoneNumber    string       `json:"phoneNumber"`
	LastLoggedIn   string       `json:"lastLoggedIn"`
	Clinics        []UserClinic `json:"clinics"`
	Settings       UserSettings `json:"settings"`
}

func SearchUsers(c *httpclient.Client, personalNumber string) ([]User, error) {
	params := url.Values{}
	params.Set("personalNumber", personalNumber)
	return httpclient.Get[[]User](c, "/v1/users", params)
}
