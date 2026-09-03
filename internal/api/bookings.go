package api

import (
	"net/url"

	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

type CreateBookingRequest struct {
	UserID                    string   `json:"userId"`
	ClinicID                  string   `json:"clinicId"`
	PatientID                 string   `json:"patientId"`
	BookingTypeID             int      `json:"bookingTypeId"`
	PatientTypeID             int      `json:"patientTypeId"`
	Date                      string   `json:"date"`
	StartTime                 string   `json:"startTime"`
	EndTime                   string   `json:"endTime"`
	InjuryNumber              string   `json:"injuryNumber,omitempty"`
	Note                      string   `json:"note,omitempty"`
	PreviousVisitID           string   `json:"previousVisitId,omitempty"`
	Fee                       *float64 `json:"fee,omitempty"`
	SMSReminder               *bool    `json:"smsReminder,omitempty"`
	SMSConfirmation           *bool    `json:"smsConfirmation,omitempty"`
	ActionCodes               []int    `json:"actionCodes,omitempty"`
	AllowOverlap              *bool    `json:"allowOverlap,omitempty"`
	ReservedForOrganizationID int      `json:"reservedForOrganizationId,omitempty"`
	ReservedForDepartmentID   int      `json:"reservedForDepartmentId,omitempty"`
}

type UpdateBookingRequest struct {
	RequestedByUserID         string `json:"requestedByUserId,omitempty"`
	BookingTypeID             int    `json:"bookingTypeId,omitempty"`
	Caregiver                 *IDRef `json:"caregiver,omitempty"`
	ReservedForOrganizationID int    `json:"reservedForOrganizationId,omitempty"`
	ReservedForDepartmentID   int    `json:"reservedForDepartmentId,omitempty"`
	PatientID                 string `json:"patientId,omitempty"`
	Date                      string `json:"date,omitempty"`
	StartTime                 string `json:"startTime,omitempty"`
	EndTime                   string `json:"endTime,omitempty"`
	Note                      string `json:"note,omitempty"`
	InjuryNumber              string `json:"injuryNumber,omitempty"`
	ActionCodes               []int  `json:"actionCodes,omitempty"`
}

type DeleteBookingRequest struct {
	UserID string `json:"userId"`
}

// BookingActionRef is the embedded action code on a booking. The id arrives as
// a number on GET /v1/bookings and as a string on the POST response.
type BookingActionRef struct {
	ID       Loose  `json:"id"`
	CodeName string `json:"codeName"`
}

type BookingClinic struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	HsaID       string      `json:"hsaId"`
	InvoiceInfo InvoiceInfo `json:"invoiceInfo"`
}

// BookingUser is the `user` block on GET /v1/bookings, where the same person is
// called `caregiver` on the other booking endpoints.
type BookingUser struct {
	ID             string `json:"id"`
	PersonalNumber string `json:"personalNumber"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Title          string `json:"title"`
}

// Booking is the shape returned by POST /v1/bookings and GET /v1/bookings.
type Booking struct {
	ID                        string             `json:"id"`
	Date                      string             `json:"date"`
	StartTime                 string             `json:"startTime"`
	EndTime                   string             `json:"endTime"`
	Note                      string             `json:"note"`
	InjuryNumber              string             `json:"injuryNumber"`
	SMSReminder               bool               `json:"smsReminder"`
	ReservedForOrganizationID int                `json:"reservedForOrganizationId"`
	ReservedForDepartmentID   int                `json:"reservedForDepartmentId"`
	Phase                     int                `json:"phase"`
	Caregiver                 Caregiver          `json:"caregiver"`
	User                      BookingUser        `json:"user"`
	ActionCodes               []BookingActionRef `json:"actionCodes"`
	BookingType               Ref                `json:"bookingType"`
	PatientType               Ref                `json:"patientType"`
	Clinic                    BookingClinic      `json:"clinic"`
	Patient                   EmbeddedPatient    `json:"patient"`
}

// ClinicBooking is the shape returned by GET /v1/clinics/:clinicId/bookings and
// PATCH /v1/bookings/:bookingId, where bookingType and bookedPatientType are
// plain strings rather than objects.
type ClinicBooking struct {
	ID                 string             `json:"id"`
	ClinicID           string             `json:"clinicId"`
	Date               string             `json:"date"`
	StartTime          string             `json:"startTime"`
	EndTime            string             `json:"endTime"`
	Arrived            int                `json:"arrived"`
	ActionCodes        []BookingActionRef `json:"actionCodes"`
	BookingDescription string             `json:"bookingDescription"`
	BookingType        string             `json:"bookingType"`
	BookedPatientType  string             `json:"bookedPatientType"`
	InjuryNumber       string             `json:"injuryNumber"`
	Payments           []Payment          `json:"payments"`
	Patient            EmbeddedPatient    `json:"patient"`
	Caregiver          Caregiver          `json:"caregiver"`
}

func CreateBooking(c *httpclient.Client, req CreateBookingRequest) (Booking, error) {
	return httpclient.Post[Booking](c, "/v1/bookings", req)
}

func ListBookings(c *httpclient.Client, params url.Values) ([]Booking, error) {
	return httpclient.Get[[]Booking](c, "/v1/bookings", params)
}

func ListClinicBookings(c *httpclient.Client, clinicID string, params url.Values) ([]ClinicBooking, error) {
	return httpclient.Get[[]ClinicBooking](c, "/v1/clinics/"+clinicID+"/bookings", params)
}

func UpdateBooking(c *httpclient.Client, bookingID string, req UpdateBookingRequest) (ClinicBooking, error) {
	return httpclient.Patch[ClinicBooking](c, "/v1/bookings/"+bookingID, req)
}

func CancelBooking(c *httpclient.Client, bookingID string, req DeleteBookingRequest) error {
	_, err := httpclient.Delete[struct{}](c, "/v1/bookings/"+bookingID, req)
	return err
}
