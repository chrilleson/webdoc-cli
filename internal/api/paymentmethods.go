package api

import (
	"fmt"

	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

type PaymentMethod struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DirectPayment bool   `json:"directPayment"`
	Active        bool   `json:"active"`
	AssetsAccount string `json:"assetsAccount"`
}

func ListPaymentMethods(c *httpclient.Client, clinicID string) ([]PaymentMethod, error) {
	return httpclient.Get[[]PaymentMethod](c, fmt.Sprintf("/v1/clinics/%s/paymentMethods", clinicID), nil)
}
