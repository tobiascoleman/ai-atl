package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	baseURL = "https://generativelanguage.googleapis.com/v1"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

type GenerateRequest struct {
	Contents []Content `json:"contents"`
	GenerationConfig GenerationConfig `json:"generationConfig,omitempty"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type GenerationConfig struct {
	Temperature float64 `json:"temperature,omitempty"`
	TopK        int     `json:"topK,omitempty"`
	TopP        float64 `json:"topP,omitempty"`
}

type GenerateResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

// NewClient creates a new Gemini API client
func NewClient() *Client {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = "demo-key" // For development
	}

	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		model: "gemini-2.5-flash-lite",
	}
}

// Generate sends a prompt to Gemini and returns the response
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", baseURL, c.model, c.apiKey)

	reqBody := GenerateRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: GenerationConfig{
			Temperature: 0.7,
			TopK:        40,
			TopP:        0.95,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var genResp GenerateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(genResp.Candidates) == 0 || len(genResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return genResp.Candidates[0].Content.Parts[0].Text, nil
}

// GenerateWithRetry generates with automatic retry on failure
func (c *Client) GenerateWithRetry(ctx context.Context, prompt string, retries int) (string, error) {
	var lastErr error
	for i := 0; i < retries; i++ {
		result, err := c.Generate(ctx, prompt)
		if err == nil {
			return result, nil
		}
		lastErr = err
		time.Sleep(time.Second * time.Duration(i+1))
	}
	return "", fmt.Errorf("failed after %d retries: %w", retries, lastErr)
}

