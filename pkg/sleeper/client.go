package sleeper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL = "https://api.sleeper.app/v1"
)

type Client struct {
	httpClient     *http.Client
	playerMappings map[string]string // NFL name -> Sleeper ID
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		playerMappings: make(map[string]string),
	}
}

// SleeperPlayer represents a player from Sleeper's players endpoint
type SleeperPlayer struct {
	PlayerID  string `json:"player_id"`
	FullName  string `json:"full_name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Position  string `json:"position"`
	Team      string `json:"team"`
	GSISID    string `json:"gsis_id"`
	ESPNID    int    `json:"espn_id"`
	Active    bool   `json:"active"`
}

// LoadPlayerMappings fetches all players and builds name->ID mapping
func (c *Client) LoadPlayerMappings(ctx context.Context) error {
	url := fmt.Sprintf("%s/players/nfl", baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch players: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var players map[string]SleeperPlayer
	if err := json.NewDecoder(resp.Body).Decode(&players); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Build mapping: normalized name -> sleeper ID
	for sleeperID, player := range players {
		if player.Active && player.FullName != "" {
			normalizedName := normalizeName(player.FullName)
			c.playerMappings[normalizedName] = sleeperID
		}
	}

	fmt.Printf("Loaded %d active player mappings from Sleeper\n", len(c.playerMappings))
	return nil
}

// GetWeeklyStats fetches weekly stats for all players
func (c *Client) GetWeeklyStats(ctx context.Context, season string, week int) (map[string]map[string]float64, error) {
	url := fmt.Sprintf("%s/stats/nfl/regular/%s/%d", baseURL, season, week)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stats map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return stats, nil
}

// GetPlayerSnapCount gets snap percentage for a specific player and week
func (c *Client) GetPlayerSnapCount(ctx context.Context, playerName string, season string, week int) (float64, error) {
	// Load player mappings if not already loaded
	if len(c.playerMappings) == 0 {
		fmt.Println("Loading Sleeper player mappings...")
		if err := c.LoadPlayerMappings(ctx); err != nil {
			return 0, err
		}
	}

	// Find Sleeper ID for this player
	normalizedName := normalizeName(playerName)
	sleeperID, ok := c.playerMappings[normalizedName]
	if !ok {
		// Try alternative formats (first name + last name)
		parts := strings.Fields(playerName)
		if len(parts) >= 2 {
			// Try "FirstLast" format
			altName := normalizeName(parts[0] + parts[len(parts)-1])
			if id, found := c.playerMappings[altName]; found {
				sleeperID = id
				ok = true
			}
		}

		if !ok {
			return 0, fmt.Errorf("player not found: %s (normalized: %s)", playerName, normalizedName)
		}
	}

	// Get weekly stats
	stats, err := c.GetWeeklyStats(ctx, season, week)
	if err != nil {
		return 0, err
	}

	// Get this player's stats
	playerStats, ok := stats[sleeperID]
	if !ok {
		// Player didn't play this week
		fmt.Printf("No stats for %s (ID: %s) in week %d\n", playerName, sleeperID, week)
		return 0, nil
	}

	// Calculate snap percentage
	offSnaps := playerStats["off_snp"]
	teamOffSnaps := playerStats["tm_off_snp"]

	if teamOffSnaps > 0 {
		snapPct := (offSnaps / teamOffSnaps) * 100
		fmt.Printf("%s: %.0f snaps / %.0f team snaps = %.1f%%\n", playerName, offSnaps, teamOffSnaps, snapPct)
		return snapPct, nil
	}

	fmt.Printf("%s: No snap data (off_snp=%.0f, tm_off_snp=%.0f)\n", playerName, offSnaps, teamOffSnaps)
	return 0, nil
}

// normalizeName converts player name to lowercase, removes punctuation
func normalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, ".", "")
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, "-", "")
	name = strings.TrimSpace(name)
	return name
}
