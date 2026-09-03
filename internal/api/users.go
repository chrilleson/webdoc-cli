package api

import (
	"fmt"
	"net/url"

	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

// InvoiceInfo is the billing block on a clinic. GET /v1/users only fills
// companyName; GET /v1/clinics fills the rest.
type InvoiceInfo struct {
	CompanyName        string `json:"companyName"`
	OrganisationNumber string `json:"organisationNumber,omitempty"`
	CompanyTelephone   string `json:"companyTelephone,omitempty"`
	CompanyAddress     string `json:"companyAddress,omitempty"`
	CompanyPost        string `json:"companyPost,omitempty"`
	CompanyPostArea    string `json:"companyPostArea,omitempty"`
	BankGiro           string `json:"bankGiro,omitempty"`
	PostGiro           string `json:"postGiro,omitempty"`
}

type UserClinic struct {
	ID                     string      `json:"id"`
	Name                   string      `json:"name"`
	HsaID                  string      `json:"hsaId"`
	JournalArchiveClinicID *string     `json:"journalArchiveClinicId"`
	InvoiceInfo            InvoiceInfo `json:"invoiceInfo"`
}

// UserSettings carries both spellings of the cost centre id: GET /v1/users
// returns the British "defaultCostCentreId", GET /v1/users/:userId the
// American "defaultCostCenterId". Use CostCentre() to read whichever is set.
type UserSettings struct {
	DefaultCostCentreID     string  `json:"defaultCostCentreId,omitempty"`
	DefaultCostCenterID     string  `json:"defaultCostCenterId,omitempty"`
	DefaultRecordTemplateID *string `json:"defaultRecordTemplateId"`
	DefaultPatientTypeID    string  `json:"defaultPatientTypeId,omitempty"`
}

func (s UserSettings) CostCentre() string {
	if s.DefaultCostCentreID != "" {
		return s.DefaultCostCentreID
	}
	return s.DefaultCostCenterID
}

// User is the shape of both GET /v1/users and GET /v1/users/:userId. The list
// endpoint returns "phoneNumber", the by-id endpoint "telephoneNumber"; both
// tags are present and Phone() picks whichever arrived.
type User struct {
	ID              string       `json:"id"`
	FirstName       string       `json:"firstName"`
	LastName        string       `json:"lastName"`
	PersonalNumber  string       `json:"personalNumber"`
	HsaID           string       `json:"hsaId"`
	Email           string       `json:"email"`
	PhoneNumber     string       `json:"phoneNumber,omitempty"`
	TelephoneNumber string       `json:"telephoneNumber,omitempty"`
	LastLoggedIn    string       `json:"lastLoggedIn"`
	Clinics         []UserClinic `json:"clinics"`
	Settings        UserSettings `json:"settings"`
}

func (u User) Phone() string {
	if u.PhoneNumber != "" {
		return u.PhoneNumber
	}
	return u.TelephoneNumber
}

// ClinicUser is the reduced shape of GET /v1/clinics/:clinicId/users.
type ClinicUser struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// Permissions maps clinic UUID -> permission name -> value ("full", "read",
// "read/write", "none", "no", "yes", ...).
type Permissions map[string]map[string]string

func SearchUsers(c *httpclient.Client, personalNumber string) ([]User, error) {
	params := url.Values{}
	params.Set("personalNumber", personalNumber)
	return httpclient.Get[[]User](c, "/v1/users", params)
}

func ListUsers(c *httpclient.Client, params url.Values) ([]User, error) {
	return httpclient.Get[[]User](c, "/v1/users", params)
}

func GetUser(c *httpclient.Client, userID string) (User, error) {
	return httpclient.Get[User](c, fmt.Sprintf("/v1/users/%s", userID), nil)
}

func GetUserPermissions(c *httpclient.Client, uuid string) (Permissions, error) {
	return httpclient.Get[Permissions](c, fmt.Sprintf("/v1/users/%s/permissions", uuid), nil)
}

func ListClinicUsers(c *httpclient.Client, clinicID string, params url.Values) ([]ClinicUser, error) {
	return httpclient.Get[[]ClinicUser](c, fmt.Sprintf("/v1/clinics/%s/users", clinicID), params)
}
