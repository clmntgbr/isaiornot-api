package sightengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type Client struct {
	apiURL    string
	apiUser   string
	apiSecret string
	http      *http.Client
}

func NewClient(apiURL, apiUser, apiSecret string) *Client {
	if apiURL == "" {
		apiURL = "https://api.sightengine.com/1.0/check.json"
	}
	return &Client{
		apiURL:    apiURL,
		apiUser:   apiUser,
		apiSecret: apiSecret,
		http: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) CheckGenAI(ctx context.Context, imageData []byte, filename string) (*CheckResponse, error) {
	if c.apiUser == "" || c.apiSecret == "" {
		return nil, fmt.Errorf("sightengine credentials are not configured")
	}
	if filename == "" {
		filename = "media.jpg"
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	mediaPart, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create media form field: %w", err)
	}
	if _, err := mediaPart.Write(imageData); err != nil {
		return nil, fmt.Errorf("failed to write media data: %w", err)
	}
	if err := writer.WriteField("models", "genai"); err != nil {
		return nil, fmt.Errorf("failed to write models field: %w", err)
	}
	if err := writer.WriteField("api_user", c.apiUser); err != nil {
		return nil, fmt.Errorf("failed to write api_user field: %w", err)
	}
	if err := writer.WriteField("api_secret", c.apiSecret); err != nil {
		return nil, fmt.Errorf("failed to write api_secret field: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())

	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("sightengine request failed: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read sightengine response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sightengine returned status %d: %s", response.StatusCode, string(responseBody))
	}

	var result CheckResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode sightengine response: %w", err)
	}
	if result.Status != "success" {
		return nil, fmt.Errorf("sightengine returned non-success status: %s", result.Status)
	}

	return &result, nil
}
