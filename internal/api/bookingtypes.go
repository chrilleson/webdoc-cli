package api

import (
	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

type RecordType struct {
	LatestID int    `json:"latestId"`
	Name     string `json:"name"`
}

type ActionCode struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type PriceLevel struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	PriceListed   float64 `json:"priceListed"`
	PriceUnlisted float64 `json:"priceUnlisted"`
	PriceNormal   float64 `json:"priceNormal"`
	PriceReduced  float64 `json:"priceReduced"`
}

type BookingType struct {
	ID                    string       `json:"id"`
	Name                  string       `json:"name"`
	ExternallyVisibleName string       `json:"externallyVisibleName"`
	BackgroundColor       string       `json:"backgroundColor"`
	TextColor             string       `json:"textColor"`
	HasSelfService        bool         `json:"hasSelfService"`
	RecordType            *RecordType  `json:"recordType"`
	ActionCode            *ActionCode  `json:"actionCode"`
	PriceLevels           []PriceLevel `json:"priceLevels"`
}

func GetBookingTypes(c *httpclient.Client) ([]BookingType, error) {
	return httpclient.Get[[]BookingType](c, "/v1/bookingTypes", nil)
}
