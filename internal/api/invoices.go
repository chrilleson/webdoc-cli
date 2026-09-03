package api

import (
	"fmt"
	"net/url"

	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

// Invoice is the shape returned by all four invoice endpoints. Every money
// field arrives as a string on invoices and as a number on bookings/visits,
// hence Loose throughout.
type Invoice struct {
	InvoiceID     string           `json:"invoiceId"`
	InvoiceNumber string           `json:"invoiceNumber"`
	InvoiceStatus string           `json:"invoiceStatus"`
	Currency      string           `json:"currency"`
	NetAmount     Loose            `json:"netAmount"`
	TotalAmount   Loose            `json:"totalAmount"`
	VAT           Loose            `json:"VAT"`
	VATPercentage Loose            `json:"VATPercentage"`
	ClinicName    string           `json:"clinicName"`
	PaidDate      *string          `json:"paidDate"`
	ExpiryDate    string           `json:"expiryDate"`
	CreatedAt     string           `json:"createdAt"`
	Payments      []Payment        `json:"payments"`
	Caregiver     Caregiver        `json:"caregiver"`
	Patient       EmbeddedPatient  `json:"patient"`
	Recipient     InvoiceRecipient `json:"recipient"`
}

type InvoiceRecipient struct {
	Type           string `json:"type"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	BillingAddress string `json:"billingAddress"`
	BillingZipCode string `json:"billingZipCode"`
	BillingCity    string `json:"billingCity"`
}

type CreateInvoiceRequest struct {
	VisitID      string `json:"visitId"`
	TotalPayment string `json:"totalPayment"`
}

type UpdateInvoiceRequest struct {
	TotalPayment  string `json:"totalPayment,omitempty"`
	PaymentMethod string `json:"paymentMethod,omitempty"`
}

func ListInvoices(c *httpclient.Client, clinicID string, params url.Values) ([]Invoice, error) {
	return httpclient.Get[[]Invoice](c, fmt.Sprintf("/v1/clinics/%s/invoices", clinicID), params)
}

func ListCreditInvoices(c *httpclient.Client, clinicID string, params url.Values) ([]Invoice, error) {
	return httpclient.Get[[]Invoice](c, fmt.Sprintf("/v1/clinics/%s/creditInvoices", clinicID), params)
}

// CreateInvoice posts a new invoice. The response is a single-element array.
func CreateInvoice(c *httpclient.Client, clinicID string, req CreateInvoiceRequest) (Invoice, error) {
	invoices, err := httpclient.Post[[]Invoice](c, fmt.Sprintf("/v1/clinics/%s/invoices", clinicID), req)
	if err != nil {
		return Invoice{}, err
	}
	if len(invoices) == 0 {
		return Invoice{}, fmt.Errorf("invoice created but no data returned")
	}
	return invoices[0], nil
}

func UpdateInvoice(c *httpclient.Client, clinicID, invoiceID string, req UpdateInvoiceRequest) (Invoice, error) {
	return httpclient.Patch[Invoice](c, fmt.Sprintf("/v1/clinics/%s/invoices/%s", clinicID, invoiceID), req)
}
