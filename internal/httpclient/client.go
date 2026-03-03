package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Body)
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func Get[T any](c *Client, path string, queryParams url.Values) (T, error) {
	var zero T

	fullURL := c.baseURL + path
	if len(queryParams) > 0 {
		fullURL += "?" + queryParams.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return zero, fmt.Errorf("building request: %w", err)
	}

	return do[T](c, req)
}

func Post[T any](c *Client, path string, body any) (T, error) {
	return sendJSON[T](c, http.MethodPost, path, body)
}

func Patch[T any](c *Client, path string, body any) (T, error) {
	return sendJSON[T](c, http.MethodPatch, path, body)
}

func PostMultipart[T any](c *Client, path string, contentType string, body io.Reader) (T, error) {
	var zero T

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, body)
	if err != nil {
		return zero, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)

	return do[T](c, req)
}

func sendJSON[T any](c *Client, method, path string, body any) (T, error) {
	var zero T

	data, err := json.Marshal(body)
	if err != nil {
		return zero, fmt.Errorf("serialising request body: %w", err)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return zero, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return do[T](c, req)
}

func do[T any](c *Client, req *http.Request) (T, error) {
	var zero T

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return zero, &APIError{
			StatusCode: resp.StatusCode,
			Body:       strings.TrimSpace(string(respBody)),
		}
	}

	if resp.StatusCode == http.StatusNoContent || len(respBody) == 0 {
		return zero, nil
	}

	var result T
	if err := json.Unmarshal(respBody, &result); err != nil {
		return zero, fmt.Errorf("decoding response: %w", err)
	}

	return result, nil
}
