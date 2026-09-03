package api

import (
	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

// ActionCodeDetail is the GET /v1/actionCodes shape. It is deliberately NOT
// named ActionCode - bookingtypes.go already uses that name for the {id, name}
// reference embedded in a booking type.
type ActionCodeDetail struct {
	ID              int    `json:"id"`
	ArticleNo       Loose  `json:"articleNo"`
	CodeName        string `json:"codeName"`
	CodeDescription string `json:"codeDescription"`
	Account         int    `json:"account"`
	Fee             Loose  `json:"fee"`
	Compensation    Loose  `json:"compensation"`
	Active          int    `json:"active"`
}

func ListActionCodes(c *httpclient.Client) ([]ActionCodeDetail, error) {
	return httpclient.Get[[]ActionCodeDetail](c, "/v1/actionCodes", nil)
}
