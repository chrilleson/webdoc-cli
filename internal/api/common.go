package api

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// Loose holds a JSON value that Webdoc returns inconsistently as a string on
// one endpoint and a number on another (prices, VAT, ids, fees). Always decodes,
// and prints as the plain text the API sent.
type Loose string

func (l *Loose) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		*l = ""
		return nil
	}
	if len(data) > 1 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*l = Loose(s)
		return nil
	}
	*l = Loose(data)
	return nil
}

func (l Loose) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(l))
}

func (l Loose) String() string {
	return string(l)
}

func (l Loose) Int() int {
	n, _ := strconv.Atoi(string(l))
	return n
}

// Ref is the {id, name} shape Webdoc uses for embedded lookups
// (bookingType, patientType, clinic, ...). The id may arrive as string or number.
type Ref struct {
	ID    Loose  `json:"id"`
	Name  string `json:"name,omitempty"`
	HsaID string `json:"hsaId,omitempty"`
}

// IDRef is the {id} request shape, e.g. booking PATCH `caregiver`.
type IDRef struct {
	ID string `json:"id"`
}

// Caregiver is the embedded caregiver on bookings, visits and invoices.
// Field presence varies per endpoint; unset fields stay empty.
type Caregiver struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Identifier string `json:"identifier"`
	HsaID      string `json:"hsaId"`
}

// Payment is the embedded payment line on bookings, visits and invoices.
type Payment struct {
	Price             Loose  `json:"price"`
	PatientFee        Loose  `json:"patientFee"`
	Amount            Loose  `json:"amount"`
	Comment           string `json:"comment"`
	Account           string `json:"account"`
	VATPercentage     Loose  `json:"VATPercentage"`
	VAT               Loose  `json:"VAT"`
	ExpiryDate        string `json:"expiryDate"`
	FreeByFreeCard    bool   `json:"freeByFreeCard"`
	ReducedByFreeCard bool   `json:"reducedByFreeCard"`
	ArticleName       string `json:"articleName"`
}

// FreeCard is the patient's high-cost protection card.
type FreeCard struct {
	CardNumber    string `json:"cardNumber"`
	ValidFrom     string `json:"validFrom"`
	ValidUntil    string `json:"validUntil"`
	AmountToLimit Loose  `json:"amountToLimit"`
}

// PatientListing is the {hsaId, countyName} listing block on embedded patients.
type PatientListing struct {
	HsaID      string `json:"hsaId"`
	CountyName string `json:"countyName"`
}

// EmbeddedPatient is the patient block nested inside bookings, visits and
// invoices. It is a superset - which fields the API fills depends on the endpoint.
type EmbeddedPatient struct {
	ID             string         `json:"id"`
	IsListed       bool           `json:"isListed"`
	PersonalNumber string         `json:"personalNumber"`
	FirstName      string         `json:"firstName"`
	LastName       string         `json:"lastName"`
	PatientType    string         `json:"patientType"`
	Address        PatientAddress `json:"address"`
	MobilePhone    string         `json:"mobilePhone"`
	PhoneNumber    string         `json:"phoneNumber"`
	Email          string         `json:"email"`
	Listing        PatientListing `json:"listing"`
}

// Keyword is the {id, value} pair used when writing records
// (PATCH /v1/visits/:visitId/records, POST /v1/externalRecords).
type Keyword struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
}
