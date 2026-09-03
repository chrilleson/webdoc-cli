package api

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/chrilleson/webdoc-cli/internal/httpclient"
)

type DocumentType struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	InReferral   bool   `json:"inreferral"`
	Active       bool   `json:"active"`
	SignRequired bool   `json:"signRequired"`
}

type UploadDocumentRequest struct {
	FilePath       string
	FileName       string
	DocumentTypeID string
	PersonalNumber string
	UserID         string
	CreatedAt      string
}

func ListDocumentTypes(c *httpclient.Client) ([]DocumentType, error) {
	return httpclient.Get[[]DocumentType](c, "/v1/documentTypes", nil)
}

// UploadDocument posts the file as multipart/form-data. The endpoint returns no
// meaningful body.
func UploadDocument(c *httpclient.Client, clinicID string, req UploadDocumentRequest) error {
	f, err := os.Open(req.FilePath)
	if err != nil {
		return fmt.Errorf("open %s: %w", req.FilePath, err)
	}
	defer f.Close()

	fileName := req.FileName
	if fileName == "" {
		fileName = filepath.Base(req.FilePath)
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fields := []struct{ key, value string }{
		{"documentTypeId", req.DocumentTypeID},
		{"personalNumber", req.PersonalNumber},
		{"userId", req.UserID},
		{"createdAt", req.CreatedAt},
	}
	for _, field := range fields {
		if field.value == "" {
			continue
		}
		if err := w.WriteField(field.key, field.value); err != nil {
			return fmt.Errorf("write field %s: %w", field.key, err)
		}
	}

	part, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return fmt.Errorf("create file part: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return fmt.Errorf("copy %s: %w", req.FilePath, err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	_, err = httpclient.PostMultipart[struct{}](c, "/v1/clinics/"+clinicID+"/documents", w.FormDataContentType(), &buf)
	return err
}
