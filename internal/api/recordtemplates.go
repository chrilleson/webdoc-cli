package api

import (
	"encoding/json"
	"fmt"

	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

type KeywordVisual struct {
	Order int `json:"order"`
	SizeX int `json:"sizeX"`
	SizeY int `json:"sizeY"`
}

// TemplateKeyword is one keyword slot on a record template. maxLength and
// defaultValue arrive as JSON strings ("64", "70"), not numbers.
type TemplateKeyword struct {
	ID             int           `json:"id"`
	Title          string        `json:"title"`
	Type           int           `json:"type"`
	MaxLength      string        `json:"maxLength"`
	DefaultValue   string        `json:"defaultValue"`
	ParentID       int           `json:"parentId"`
	WarnIfEmpty    bool          `json:"warnIfEmpty"`
	Information    string        `json:"information"`
	RepetitionTime string        `json:"repetitionTime"`
	WarningTime    string        `json:"warningTime"`
	Visual         KeywordVisual `json:"visual"`
}

type RecordTemplate struct {
	ID               int               `json:"id"`
	Title            string            `json:"title"`
	TemplateType     int               `json:"templateType"`
	PreviousVersions []json.RawMessage `json:"previousVersions"`
	Keywords         []TemplateKeyword `json:"keywords"`
}

func ListRecordTemplates(c *httpclient.Client) ([]RecordTemplate, error) {
	return httpclient.Get[[]RecordTemplate](c, "/v1/recordTemplates", nil)
}

// GetRecordTemplate fetches one template. The endpoint returns a single-element
// JSON array rather than an object.
func GetRecordTemplate(c *httpclient.Client, templateID string) (RecordTemplate, error) {
	templates, err := httpclient.Get[[]RecordTemplate](c, "/v1/recordTemplates/"+templateID, nil)
	if err != nil {
		return RecordTemplate{}, err
	}
	if len(templates) == 0 {
		return RecordTemplate{}, fmt.Errorf("record template %s not found", templateID)
	}
	return templates[0], nil
}
