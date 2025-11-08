package nflverse

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL = "https://github.com/nflverse/nflverse-data/releases/download"
)

// Client for fetching NFLverse data
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new NFLverse client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// FetchPlayerStats downloads player stats for a given season
func (c *Client) FetchPlayerStats(ctx context.Context, season int) ([]byte, error) {
	url := fmt.Sprintf("%s/player_stats/player_stats_%d.parquet", baseURL, season)
	return c.downloadFile(ctx, url)
}

// FetchPlayByPlay downloads play-by-play data for a given season
func (c *Client) FetchPlayByPlay(ctx context.Context, season int) ([]byte, error) {
	url := fmt.Sprintf("%s/pbp/play_by_play_%d.parquet", baseURL, season)
	return c.downloadFile(ctx, url)
}

// FetchRosters downloads roster data for a given season
func (c *Client) FetchRosters(ctx context.Context, season int) ([]byte, error) {
	url := fmt.Sprintf("%s/rosters/roster_%d.parquet", baseURL, season)
	return c.downloadFile(ctx, url)
}

// FetchInjuries downloads injury data
func (c *Client) FetchInjuries(ctx context.Context, season int) ([]byte, error) {
	url := fmt.Sprintf("%s/injuries/injuries_%d.parquet", baseURL, season)
	return c.downloadFile(ctx, url)
}

// FetchNextGenStats downloads Next Gen Stats for a given stat type and season
func (c *Client) FetchNextGenStats(ctx context.Context, season int, statType string) ([]byte, error) {
	// statType: receiving, rushing, passing
	url := fmt.Sprintf("%s/nextgen_stats/ngs_%s_%d.parquet", baseURL, statType, season)
	return c.downloadFile(ctx, url)
}

// downloadFile downloads a file from the given URL
func (c *Client) downloadFile(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// ParseParquetData is a placeholder for parquet parsing
// In a real implementation, you would use a library like github.com/xitongsys/parquet-go
func ParseParquetData(data []byte) ([]map[string]interface{}, error) {
	// TODO: Implement parquet parsing
	// For hackathon, this would return mock data or use a parquet library
	return []map[string]interface{}{}, nil
}
