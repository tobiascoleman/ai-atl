package espn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"github.com/ai-atl/nfl-platform/internal/models"
)

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const (
	baseURL = "https://fantasy.espn.com/apis/v3/games/ffl"
)

// Client handles ESPN Fantasy Football API requests
type Client struct {
	httpClient *http.Client
	leagueID   string
	seasonYear int
	swid       string
	espnS2     string
}

// NewClient creates a new ESPN Fantasy client
func NewClient(leagueID string, seasonYear int, swid, espnS2 string) *Client {
	jar, _ := cookiejar.New(nil)

	return &Client{
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
		leagueID:   leagueID,
		seasonYear: seasonYear,
		swid:       swid,
		espnS2:     espnS2,
	}
}

// GetLeague fetches complete league information including settings and all teams
func (c *Client) GetLeague(ctx context.Context) (*models.ESPNLeague, error) {
	endpoint := fmt.Sprintf("%s/seasons/%d/segments/0/leagues/%s?view=mTeam&view=mRoster&view=mSettings&view=mStandings",
		baseURL, c.seasonYear, c.leagueID)

	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Debug: Print first chars of response
	fmt.Printf("[ESPN Client DEBUG] GetLeague response length: %d bytes\n", len(data))
	if len(data) > 0 {
		preview := string(data[:min(500, len(data))])
		fmt.Printf("[ESPN Client DEBUG] First 500 chars: %s\n", preview)
		// Write to file for inspection
		os.WriteFile("/tmp/espn_getleague_response.txt", data, 0644)
		fmt.Printf("[ESPN Client DEBUG] Full response saved to /tmp/espn_getleague_response.txt\n")
	}

	var response struct {
		ID       int `json:"id"`
		SeasonID int `json:"seasonId"`
		Status   struct {
			CurrentMatchupPeriod int `json:"currentMatchupPeriod"`
			LatestScoringPeriod  int `json:"latestScoringPeriod"`
			FinalScoringPeriod   int `json:"finalScoringPeriod"`
		} `json:"status"`
		Settings struct {
			Name             string `json:"name"`
			Size             int    `json:"size"`
			ScoringType      string `json:"scoringType"`
			SchedulePeriods  int    `json:"schedulePeriods"`
			PlayoffTeamCount int    `json:"playoffTeamCount"`
			TradeSettings    struct {
				Deadline          int `json:"deadlineDate"`
				VetoVotesRequired int `json:"vetoVotesRequired"`
			} `json:"tradeSettings"`
			AcquisitionSettings struct {
				WaiverProcessHour *int `json:"waiverProcessHour"`
			} `json:"acquisitionSettings"`
			ScoringSettings struct {
				ScoringType string `json:"scoringType"`
			} `json:"scoringSettings"`
		} `json:"settings"`
		Teams []struct {
			ID         int    `json:"id"`
			Abbrev     string `json:"abbrev"`
			Location   string `json:"location"`
			Nickname   string `json:"nickname"`
			Logo       string `json:"logo"`
			DivisionID int    `json:"divisionId"`
			Record     struct {
				Overall struct {
					Wins   int `json:"wins"`
					Losses int `json:"losses"`
					Ties   int `json:"ties"`
				} `json:"overall"`
				PointsFor     float64 `json:"pointsFor"`
				PointsAgainst float64 `json:"pointsAgainst"`
			} `json:"record"`
			Roster struct {
				Entries []struct {
					PlayerPoolEntry struct {
						Player struct {
							ID        int    `json:"id"`
							FullName  string `json:"fullName"`
							ProTeam   int    `json:"proTeamId"`
							Position  string `json:"defaultPositionId"`
							InjStatus string `json:"injuryStatus"`
							Ownership struct {
								PercentOwned   float64 `json:"percentOwned"`
								PercentStarted float64 `json:"percentStarted"`
							} `json:"ownership"`
						} `json:"player"`
					} `json:"playerPoolEntry"`
					LineupSlotID int `json:"lineupSlotId"`
				} `json:"entries"`
			} `json:"roster"`
			PlayoffSeed int      `json:"playoffSeed"`
			Owners      []string `json:"owners"`
		} `json:"teams"`
		Members []struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"members"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse league response: %w", err)
	}

	// Build owner map
	ownerMap := make(map[string]string)
	for _, member := range response.Members {
		ownerMap[member.ID] = member.DisplayName
	}

	league := &models.ESPNLeague{
		Settings: models.ESPNLeagueSettings{
			LeagueID:           c.leagueID,
			SeasonYear:         c.seasonYear,
			Name:               response.Settings.Name,
			Size:               response.Settings.Size,
			CurrentWeek:        response.Status.LatestScoringPeriod,
			ScoringPeriodID:    response.Status.LatestScoringPeriod,
			FinalScoringPeriod: response.Status.FinalScoringPeriod,
			ScoringType:        response.Settings.ScoringType,
			PlayoffTeamCount:   response.Settings.PlayoffTeamCount,
			TradeDeadline:      response.Settings.TradeSettings.Deadline,
			VetoVotesRequired:  response.Settings.TradeSettings.VetoVotesRequired,
			WaiverProcessHour:  response.Settings.AcquisitionSettings.WaiverProcessHour,
			RegSeasonCount:     response.Settings.SchedulePeriods,
		},
		Teams: []models.ESPNTeam{},
	}

	// Parse teams
	for _, t := range response.Teams {
		team := models.ESPNTeam{
			TeamID:        t.ID,
			Abbrev:        t.Abbrev,
			TeamName:      fmt.Sprintf("%s %s", t.Location, t.Nickname),
			Wins:          t.Record.Overall.Wins,
			Losses:        t.Record.Overall.Losses,
			Ties:          t.Record.Overall.Ties,
			PointsFor:     t.Record.PointsFor,
			PointsAgainst: t.Record.PointsAgainst,
			Standing:      t.PlayoffSeed,
			DivisionID:    t.DivisionID,
			LogoURL:       t.Logo,
			Roster:        []models.ESPNPlayer{},
		}

		// Get owner name if available
		if len(t.Owners) > 0 {
			if ownerName, ok := ownerMap[t.Owners[0]]; ok {
				team.Owner = ownerName
			}
		}

		// Parse roster
		for _, entry := range t.Roster.Entries {
			player := models.ESPNPlayer{
				PlayerID:       entry.PlayerPoolEntry.Player.ID,
				Name:           entry.PlayerPoolEntry.Player.FullName,
				Position:       c.mapPosition(entry.PlayerPoolEntry.Player.Position),
				Team:           c.mapTeam(entry.PlayerPoolEntry.Player.ProTeam),
				SlotPosition:   c.mapSlotPosition(entry.LineupSlotID),
				InjuryStatus:   entry.PlayerPoolEntry.Player.InjStatus,
				PercentOwned:   entry.PlayerPoolEntry.Player.Ownership.PercentOwned,
				PercentStarted: entry.PlayerPoolEntry.Player.Ownership.PercentStarted,
			}
			team.Roster = append(team.Roster, player)
		}

		league.Teams = append(league.Teams, team)
	}

	return league, nil
}

// GetTeam fetches team info including roster and record
func (c *Client) GetTeam(ctx context.Context, teamID int) (*models.ESPNTeam, error) {
	// First, try a simple league settings request to verify auth works
	settingsEndpoint := fmt.Sprintf("%s/seasons/%d/segments/0/leagues/%s",
		baseURL, c.seasonYear, c.leagueID)

	fmt.Printf("[ESPN Client] Testing auth with settings endpoint: %s\n", settingsEndpoint)
	testData, err := c.doRequest(ctx, "GET", settingsEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("auth test failed: %w", err)
	}

	// Check if we got JSON back
	var testJSON map[string]interface{}
	if err := json.Unmarshal(testData, &testJSON); err != nil {
		return nil, fmt.Errorf("auth test returned non-JSON: %w", err)
	}
	fmt.Printf("[ESPN Client] Auth test successful! League data received.\n")

	// Now get the full team data
	endpoint := fmt.Sprintf("%s/seasons/%d/segments/0/leagues/%s?view=mTeam&view=mRoster",
		baseURL, c.seasonYear, c.leagueID)

	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Debug: Log first part of response
	preview := string(data)
	if len(preview) > 200 {
		preview = preview[:200]
	}
	fmt.Printf("[ESPN Client] Response preview: %s...\n", preview)

	var response struct {
		Teams []struct {
			ID       int    `json:"id"`
			Location string `json:"location"`
			Nickname string `json:"nickname"`
			Record   struct {
				Overall struct {
					Wins   int `json:"wins"`
					Losses int `json:"losses"`
					Ties   int `json:"ties"`
				} `json:"overall"`
				PointsFor     float64 `json:"pointsFor"`
				PointsAgainst float64 `json:"pointsAgainst"`
			} `json:"record"`
			Roster struct {
				Entries []struct {
					PlayerPoolEntry struct {
						Player struct {
							ID        int    `json:"id"`
							FullName  string `json:"fullName"`
							ProTeam   int    `json:"proTeamId"`
							Position  string `json:"defaultPositionId"`
							InjStatus string `json:"injuryStatus"`
						} `json:"player"`
					} `json:"playerPoolEntry"`
					LineupSlotID int `json:"lineupSlotId"`
				} `json:"entries"`
			} `json:"roster"`
			PlayoffSeed int `json:"playoffSeed"`
		} `json:"teams"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Find the specific team
	for _, t := range response.Teams {
		if t.ID == teamID {
			team := &models.ESPNTeam{
				TeamID:        t.ID,
				TeamName:      fmt.Sprintf("%s %s", t.Location, t.Nickname),
				Wins:          t.Record.Overall.Wins,
				Losses:        t.Record.Overall.Losses,
				Ties:          t.Record.Overall.Ties,
				PointsFor:     t.Record.PointsFor,
				PointsAgainst: t.Record.PointsAgainst,
				Standing:      t.PlayoffSeed,
				Roster:        []models.ESPNPlayer{},
			}

			// Parse roster
			for _, entry := range t.Roster.Entries {
				player := models.ESPNPlayer{
					PlayerID:     entry.PlayerPoolEntry.Player.ID,
					Name:         entry.PlayerPoolEntry.Player.FullName,
					Position:     c.mapPosition(entry.PlayerPoolEntry.Player.Position),
					Team:         c.mapTeam(entry.PlayerPoolEntry.Player.ProTeam),
					SlotPosition: c.mapSlotPosition(entry.LineupSlotID),
					InjuryStatus: entry.PlayerPoolEntry.Player.InjStatus,
				}
				team.Roster = append(team.Roster, player)
			}

			return team, nil
		}
	}

	return nil, fmt.Errorf("team not found")
}

// GetRoster fetches the current roster for a team
func (c *Client) GetRoster(ctx context.Context, teamID int) ([]models.ESPNPlayer, error) {
	team, err := c.GetTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return team.Roster, nil
}

// GetMatchup fetches matchup information for a specific week
func (c *Client) GetMatchup(ctx context.Context, teamID int, week int) (*models.ESPNMatchup, error) {
	endpoint := fmt.Sprintf("%s/seasons/%d/segments/0/leagues/%s?view=mMatchup&scoringPeriodId=%d",
		baseURL, c.seasonYear, c.leagueID, week)

	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Schedule []struct {
			MatchupPeriodId int `json:"matchupPeriodId"`
			Home            struct {
				TeamID      int     `json:"teamId"`
				TotalPoints float64 `json:"totalPoints"`
			} `json:"home"`
			Away struct {
				TeamID      int     `json:"teamId"`
				TotalPoints float64 `json:"totalPoints"`
			} `json:"away"`
		} `json:"schedule"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Find matchup for this team
	for _, m := range response.Schedule {
		if m.MatchupPeriodId == week && (m.Home.TeamID == teamID || m.Away.TeamID == teamID) {
			matchup := &models.ESPNMatchup{
				Week:       week,
				HomeTeamID: m.Home.TeamID,
				AwayTeamID: m.Away.TeamID,
				HomeScore:  m.Home.TotalPoints,
				AwayScore:  m.Away.TotalPoints,
			}

			if m.Home.TotalPoints > m.Away.TotalPoints {
				matchup.Winner = "home"
			} else if m.Away.TotalPoints > m.Home.TotalPoints {
				matchup.Winner = "away"
			} else {
				matchup.Winner = "tie"
			}

			return matchup, nil
		}
	}

	return nil, fmt.Errorf("matchup not found for week %d", week)
}

// GetFreeAgents fetches available free agents
func (c *Client) GetFreeAgents(ctx context.Context, position string, limit int) ([]models.ESPNFreeAgent, error) {
	endpoint := fmt.Sprintf("%s/seasons/%d/segments/0/leagues/%s?view=kona_player_info",
		baseURL, c.seasonYear, c.leagueID)

	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Players []struct {
			Player struct {
				ID        int     `json:"id"`
				FullName  string  `json:"fullName"`
				ProTeam   int     `json:"proTeamId"`
				Position  string  `json:"defaultPositionId"`
				Ownership float64 `json:"ownership"`
				InjStatus string  `json:"injuryStatus"`
			} `json:"player"`
		} `json:"players"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	freeAgents := []models.ESPNFreeAgent{}
	for _, p := range response.Players {
		if len(freeAgents) >= limit {
			break
		}

		playerPos := c.mapPosition(p.Player.Position)
		if position == "" || playerPos == position {
			agent := models.ESPNFreeAgent{
				PlayerID:     p.Player.ID,
				Name:         p.Player.FullName,
				Position:     playerPos,
				Team:         c.mapTeam(p.Player.ProTeam),
				PercentOwned: p.Player.Ownership,
				InjuryStatus: p.Player.InjStatus,
			}
			freeAgents = append(freeAgents, agent)
		}
	}

	return freeAgents, nil
}

// GetStandings fetches league standings
func (c *Client) GetStandings(ctx context.Context) ([]models.ESPNTeam, error) {
	endpoint := fmt.Sprintf("%s/seasons/%d/segments/0/leagues/%s?view=mTeam",
		baseURL, c.seasonYear, c.leagueID)

	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Teams []struct {
			ID       int    `json:"id"`
			Location string `json:"location"`
			Nickname string `json:"nickname"`
			Record   struct {
				Overall struct {
					Wins   int `json:"wins"`
					Losses int `json:"losses"`
					Ties   int `json:"ties"`
				} `json:"overall"`
				PointsFor     float64 `json:"pointsFor"`
				PointsAgainst float64 `json:"pointsAgainst"`
			} `json:"record"`
			PlayoffSeed int `json:"playoffSeed"`
		} `json:"teams"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	standings := []models.ESPNTeam{}
	for _, t := range response.Teams {
		team := models.ESPNTeam{
			TeamID:        t.ID,
			TeamName:      fmt.Sprintf("%s %s", t.Location, t.Nickname),
			Wins:          t.Record.Overall.Wins,
			Losses:        t.Record.Overall.Losses,
			Ties:          t.Record.Overall.Ties,
			PointsFor:     t.Record.PointsFor,
			PointsAgainst: t.Record.PointsAgainst,
			Standing:      t.PlayoffSeed,
		}
		standings = append(standings, team)
	}

	return standings, nil
}

// GetBoxScore fetches detailed box score for a specific week's matchup
func (c *Client) GetBoxScore(ctx context.Context, week int) ([]models.ESPNBoxScore, error) {
	endpoint := fmt.Sprintf("%s/seasons/%d/segments/0/leagues/%s?view=mMatchupScore&view=mScoreboard&scoringPeriodId=%d",
		baseURL, c.seasonYear, c.leagueID, week)

	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Schedule []struct {
			MatchupPeriodID int `json:"matchupPeriodId"`
			Home            struct {
				TeamID                        int     `json:"teamId"`
				TotalPoints                   float64 `json:"totalPoints"`
				TotalProjectedPoints          float64 `json:"totalProjectedPoints"`
				RosterForCurrentScoringPeriod struct {
					Entries []struct {
						PlayerPoolEntry struct {
							Player struct {
								ID       int    `json:"id"`
								FullName string `json:"fullName"`
								ProTeam  int    `json:"proTeamId"`
								Position string `json:"defaultPositionId"`
								Stats    []struct {
									ScoringPeriodID int                `json:"scoringPeriodId"`
									AppliedTotal    float64            `json:"appliedTotal"`
									Stats           map[string]float64 `json:"stats"`
								} `json:"stats"`
							} `json:"player"`
						} `json:"playerPoolEntry"`
						LineupSlotID int `json:"lineupSlotId"`
					} `json:"entries"`
				} `json:"rosterForCurrentScoringPeriod"`
			} `json:"home"`
			Away struct {
				TeamID                        int     `json:"teamId"`
				TotalPoints                   float64 `json:"totalPoints"`
				TotalProjectedPoints          float64 `json:"totalProjectedPoints"`
				RosterForCurrentScoringPeriod struct {
					Entries []struct {
						PlayerPoolEntry struct {
							Player struct {
								ID       int    `json:"id"`
								FullName string `json:"fullName"`
								ProTeam  int    `json:"proTeamId"`
								Position string `json:"defaultPositionId"`
								Stats    []struct {
									ScoringPeriodID int                `json:"scoringPeriodId"`
									AppliedTotal    float64            `json:"appliedTotal"`
									Stats           map[string]float64 `json:"stats"`
								} `json:"stats"`
							} `json:"player"`
						} `json:"playerPoolEntry"`
						LineupSlotID int `json:"lineupSlotId"`
					} `json:"entries"`
				} `json:"rosterForCurrentScoringPeriod"`
			} `json:"away"`
		} `json:"schedule"`
		Teams []struct {
			ID       int    `json:"id"`
			Location string `json:"location"`
			Nickname string `json:"nickname"`
		} `json:"teams"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse box score response: %w", err)
	}

	// Build team name map
	teamNames := make(map[int]string)
	for _, team := range response.Teams {
		teamNames[team.ID] = fmt.Sprintf("%s %s", team.Location, team.Nickname)
	}

	boxScores := []models.ESPNBoxScore{}
	for _, matchup := range response.Schedule {
		if matchup.MatchupPeriodID != week {
			continue
		}

		boxScore := models.ESPNBoxScore{
			Week: week,
			HomeTeam: models.ESPNBoxTeam{
				TeamID:   matchup.Home.TeamID,
				TeamName: teamNames[matchup.Home.TeamID],
				Score:    matchup.Home.TotalPoints,
			},
			AwayTeam: models.ESPNBoxTeam{
				TeamID:   matchup.Away.TeamID,
				TeamName: teamNames[matchup.Away.TeamID],
				Score:    matchup.Away.TotalPoints,
			},
			HomeLineup: []models.ESPNPlayerBox{},
			AwayLineup: []models.ESPNPlayerBox{},
		}

		// Parse home lineup
		for _, entry := range matchup.Home.RosterForCurrentScoringPeriod.Entries {
			player := c.parseBoxPlayer(entry, week)
			boxScore.HomeLineup = append(boxScore.HomeLineup, player)
		}

		// Parse away lineup
		for _, entry := range matchup.Away.RosterForCurrentScoringPeriod.Entries {
			player := c.parseBoxPlayer(entry, week)
			boxScore.AwayLineup = append(boxScore.AwayLineup, player)
		}

		boxScores = append(boxScores, boxScore)
	}

	return boxScores, nil
}

// GetRecentActivity fetches recent league transactions and activity
func (c *Client) GetRecentActivity(ctx context.Context, size int) ([]models.ESPNActivity, error) {
	endpoint := fmt.Sprintf("%s/seasons/%d/segments/0/leagues/%s?view=kona_league_communication",
		baseURL, c.seasonYear, c.leagueID)

	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Communication struct {
			Topics []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Date     int64  `json:"date"`
				TeamID   int    `json:"teamId,omitempty"`
				Messages []struct {
					MessageTypeID int      `json:"messageTypeId"`
					TargetID      int      `json:"targetId"`
					To            []string `json:"to"`
					From          int      `json:"from"`
					MessageText   string   `json:"messageText"`
				} `json:"messages"`
			} `json:"topics"`
		} `json:"communication"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		// Activity endpoint might not be available for all leagues
		// Return empty array instead of error
		return []models.ESPNActivity{}, nil
	}

	activities := []models.ESPNActivity{}
	count := 0
	for _, topic := range response.Communication.Topics {
		if count >= size {
			break
		}

		activity := models.ESPNActivity{
			ID:     topic.ID,
			Type:   topic.Type,
			Date:   topic.Date,
			TeamID: topic.TeamID,
		}

		// Parse messages to get player names and description
		if len(topic.Messages) > 0 {
			activity.Description = topic.Messages[0].MessageText
		}

		activities = append(activities, activity)
		count++
	}

	return activities, nil
}

// parseBoxPlayer is a helper to parse player data from box score
func (c *Client) parseBoxPlayer(entry interface{}, week int) models.ESPNPlayerBox {
	// Type assertion for the nested structure
	type entryType struct {
		PlayerPoolEntry struct {
			Player struct {
				ID       int    `json:"id"`
				FullName string `json:"fullName"`
				ProTeam  int    `json:"proTeamId"`
				Position string `json:"defaultPositionId"`
				Stats    []struct {
					ScoringPeriodID int                `json:"scoringPeriodId"`
					AppliedTotal    float64            `json:"appliedTotal"`
					Stats           map[string]float64 `json:"stats"`
				} `json:"stats"`
			} `json:"player"`
		} `json:"playerPoolEntry"`
		LineupSlotID int `json:"lineupSlotId"`
	}

	// Marshal and unmarshal to convert interface to struct
	jsonData, _ := json.Marshal(entry)
	var e entryType
	json.Unmarshal(jsonData, &e)

	player := models.ESPNPlayerBox{
		PlayerID:     e.PlayerPoolEntry.Player.ID,
		Name:         e.PlayerPoolEntry.Player.FullName,
		Position:     c.mapPosition(e.PlayerPoolEntry.Player.Position),
		Team:         c.mapTeam(e.PlayerPoolEntry.Player.ProTeam),
		SlotPosition: c.mapSlotPosition(e.LineupSlotID),
		Stats:        make(map[string]float64),
	}

	// Find stats for the specific week
	for _, stat := range e.PlayerPoolEntry.Player.Stats {
		if stat.ScoringPeriodID == week {
			player.Points = stat.AppliedTotal
			player.Stats = stat.Stats
			break
		}
	}

	return player
}

// doRequest performs HTTP request with ESPN authentication
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication cookies
	req.AddCookie(&http.Cookie{Name: "swid", Value: c.swid})
	req.AddCookie(&http.Cookie{Name: "espn_s2", Value: c.espnS2})

	// Set headers to mimic a browser request
	// Note: NOT setting Accept-Encoding to avoid gzip compression issues
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://fantasy.espn.com/football/")
	req.Header.Set("Origin", "https://fantasy.espn.com")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	// Debug logging
	fmt.Printf("[ESPN Client] Request to: %s\n", endpoint)
	fmt.Printf("[ESPN Client] SWID length: %d, S2 length: %d\n", len(c.swid), len(c.espnS2))
	fmt.Printf("[ESPN Client] SWID value: %s\n", c.swid)
	fmt.Printf("[ESPN Client] S2 preview: %s...\n", c.espnS2[:min(50, len(c.espnS2))])
	fmt.Printf("[ESPN Client] Cookie header: %s\n", req.Header.Get("Cookie"))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("ESPN authentication failed - cookies may be expired")
	}

	if resp.StatusCode != http.StatusOK {
		// Log first 500 chars of response for debugging
		preview := string(data)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		return nil, fmt.Errorf("ESPN API returned status %d. Response: %s", resp.StatusCode, preview)
	}

	// Check if response looks like HTML instead of JSON
	if len(data) > 0 && data[0] == '<' {
		// Write full HTML response to a debug file
		debugFile := "/tmp/espn_error_response.html"
		if err := os.WriteFile(debugFile, data, 0644); err == nil {
			fmt.Printf("\n[ESPN Client] ==================== HTML ERROR ====================\n")
			fmt.Printf("Received HTML instead of JSON (likely auth/access issue)\n")
			fmt.Printf("Full HTML response saved to: %s\n", debugFile)
			fmt.Printf("First 500 chars: %s\n", string(data[:min(500, len(data))]))
			fmt.Printf("==========================================================\n\n")
		}
		return nil, fmt.Errorf("ESPN returned HTML instead of JSON (likely auth issue)")
	}

	return data, nil
}

// Helper functions to map ESPN IDs to readable values

func (c *Client) mapPosition(posID string) string {
	positions := map[string]string{
		"1": "QB", "2": "RB", "3": "WR", "4": "TE",
		"5": "K", "16": "D/ST",
	}
	if pos, ok := positions[posID]; ok {
		return pos
	}
	return posID
}

func (c *Client) mapSlotPosition(slotID int) string {
	slots := map[int]string{
		0: "QB", 2: "RB", 4: "WR", 6: "TE",
		16: "D/ST", 17: "K", 20: "BENCH", 23: "FLEX",
	}
	if slot, ok := slots[slotID]; ok {
		return slot
	}
	return "BENCH"
}

func (c *Client) mapTeam(teamID int) string {
	teams := map[int]string{
		1: "ATL", 2: "BUF", 3: "CHI", 4: "CIN", 5: "CLE", 6: "DAL",
		7: "DEN", 8: "DET", 9: "GB", 10: "TEN", 11: "IND", 12: "KC",
		13: "LV", 14: "LAR", 15: "MIA", 16: "MIN", 17: "NE", 18: "NO",
		19: "NYG", 20: "NYJ", 21: "PHI", 22: "ARI", 23: "PIT", 24: "LAC",
		25: "SF", 26: "SEA", 27: "TB", 28: "WSH", 29: "CAR", 30: "JAX",
		33: "BAL", 34: "HOU",
	}
	if team, ok := teams[teamID]; ok {
		return team
	}
	return "FA"
}
