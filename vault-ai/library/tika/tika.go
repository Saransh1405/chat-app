package tika

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// Client represents a Tika server client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Tika client
func NewClient() *Client {
	tikaURL := os.Getenv("TIKA_SERVER_URL")
	if tikaURL == "" {
		tikaURL = "http://localhost:9998" // fallback for local development
	}


	log.Printf("TIKA_SERVER_URL: %s", tikaURL)

	return &Client{
		BaseURL: tikaURL,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second, // Tika can be slow for large files
		},
	}
}

// ExtractText extracts text from a file using Tika server
func (c *Client) ExtractText(fileHeader *multipart.FileHeader) (string, error) {
	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Create request to Tika server
	req, err := http.NewRequest("PUT", c.BaseURL+"/tika", bytes.NewReader(fileContent))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("Content-Type", "application/octet-stream")

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Tika: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("tika server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read extracted text
	text, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(text), nil
}

// ExtractMetadata extracts metadata from a file using Tika server
func (c *Client) ExtractMetadata(fileHeader *multipart.FileHeader) (map[string]interface{}, error) {
	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Create request to Tika server
	req, err := http.NewRequest("PUT", c.BaseURL+"/meta", bytes.NewReader(fileContent))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/octet-stream")

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Tika: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tika server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse metadata (simplified - you might want to use encoding/json for full parsing)
	metadata := make(map[string]interface{})

	return metadata, nil
}

// DetectContentType detects the content type of a file
func (c *Client) DetectContentType(fileHeader *multipart.FileHeader) (string, error) {
	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Create request to Tika server
	req, err := http.NewRequest("PUT", c.BaseURL+"/detect/stream", bytes.NewReader(fileContent))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Tika: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tika server returned status %d", resp.StatusCode)
	}

	// Read content type
	contentType, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(contentType), nil
}

// HealthCheck checks if Tika server is available
func (c *Client) HealthCheck() error {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/tika")
	if err != nil {
		return fmt.Errorf("tika server is not reachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tika server returned status %d", resp.StatusCode)
	}

	return nil
}
