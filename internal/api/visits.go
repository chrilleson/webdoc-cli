package api

import (
	"net/url"

	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

type CreateVisitRequest struct {
	BookingID string `json:"bookingId"`
	VisitDate string `json:"visitDate"`
}

type SignVisitRequest struct {
	UserID string `json:"userId"`
}

type Visit struct {
	ID               string          `json:"id"`
	ClinicID         string          `json:"clinicId"`
	VisitDate        string          `json:"visitDate"`
	BookingID        string          `json:"bookingId"`
	RecordTemplateID Loose           `json:"recordTemplateId"`
	InjuryNumber     string          `json:"injuryNumber"`
	Payments         []Payment       `json:"payments"`
	Patient          EmbeddedPatient `json:"patient"`
	Caregiver        Caregiver       `json:"caregiver"`
}

func CreateVisit(c *httpclient.Client, req CreateVisitRequest) (Visit, error) {
	return httpclient.Post[Visit](c, "/v2/visits", req)
}

func ListClinicVisits(c *httpclient.Client, clinicID string, params url.Values) ([]Visit, error) {
	return httpclient.Get[[]Visit](c, "/v1/clinics/"+clinicID+"/visits", params)
}

// SignVisit signs the visit's record. The endpoint returns no body.
func SignVisit(c *httpclient.Client, visitID string, req SignVisitRequest) error {
	_, err := httpclient.Post[struct{}](c, "/v1/visits/"+visitID+"/records/signatures", req)
	return err
}
