package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Player represents a player's roster entry for a specific season
// This is the base roster data from NFLverse - stats come from separate sources
type Player struct {
	ID       bson.ObjectID `json:"id" bson:"_id,omitempty"`
	NFLID    string        `json:"nfl_id" bson:"nfl_id"`
	Season   int           `json:"season" bson:"season"` // Year of this roster entry
	Name     string        `json:"name" bson:"name"`
	Team     string        `json:"team" bson:"team"` // Current team for this season
	Position string        `json:"position" bson:"position"`

	// Injury status from weekly rosters
	Status                string `json:"status" bson:"status"`                                   // ACT or INA (injured)
	StatusDescriptionAbbr string `json:"status_description_abbr" bson:"status_description_abbr"` // R01 (R/Injured), P02 (Prac Sq.; Inj), etc.
	Week                  int    `json:"week" bson:"week"`                                       // Latest week this status was updated

	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// PlayerStats represents season-level stats for a player
// This would be loaded from player_stats Parquet files
type PlayerStats struct {
	ID         bson.ObjectID `json:"id" bson:"_id,omitempty"`
	NFLID      string        `json:"nfl_id" bson:"nfl_id"`
	Season     int           `json:"season" bson:"season"`
	SeasonType string        `json:"season_type" bson:"season_type"` // REG, POST

	// Stats will be populated when we load player_stats files
	PassingYards   int `json:"passing_yards" bson:"passing_yards"`
	PassingTDs     int `json:"passing_tds" bson:"passing_tds"`
	Interceptions  int `json:"interceptions" bson:"interceptions"`
	RushingYards   int `json:"rushing_yards" bson:"rushing_yards"`
	RushingTDs     int `json:"rushing_tds" bson:"rushing_tds"`
	Receptions     int `json:"receptions" bson:"receptions"`
	ReceivingYards int `json:"receiving_yards" bson:"receiving_yards"`
	ReceivingTDs   int `json:"receiving_tds" bson:"receiving_tds"`
	Targets        int `json:"targets" bson:"targets"`

	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// WeeklyRosterEntry represents a player's weekly status from roster_weekly
// Used primarily for extracting injury status
type WeeklyRosterEntry struct {
	NFLID                 string `json:"nfl_id"`
	Season                int    `json:"season"`
	Week                  int    `json:"week"`
	Team                  string `json:"team"`
	Status                string `json:"status"`                  // ACT or INA
	StatusDescriptionAbbr string `json:"status_description_abbr"` // R01, P02, etc.
}

type WeeklyStat struct {
	Week            int     `json:"week" bson:"week"`
	Season          int     `json:"season" bson:"season"`
	Opponent        string  `json:"opponent" bson:"opponent"`
	Yards           int     `json:"yards" bson:"yards"`
	Touchdowns      int     `json:"touchdowns" bson:"touchdowns"`
	Receptions      int     `json:"receptions" bson:"receptions"`
	Targets         int     `json:"targets" bson:"targets"`
	Carries         int     `json:"carries" bson:"carries"`
	PassingYards    int     `json:"passing_yards" bson:"passing_yards"`
	PassingTDs      int     `json:"passing_tds" bson:"passing_tds"`
	Interceptions   int     `json:"interceptions" bson:"interceptions"`
	EPA             float64 `json:"epa" bson:"epa"`
	ProjectedPoints float64 `json:"projected_points" bson:"projected_points"`
	ActualPoints    float64 `json:"actual_points" bson:"actual_points"`
}

// IsInjured returns true if the player has an injury status
func (w *WeeklyRosterEntry) IsInjured() bool {
	if w.Status == "INA" {
		return true
	}
	// Check for injury-related status codes
	injuryStatuses := []string{"R01", "R04", "R48", "P02"} // R/Injured, R/PUP, R/Injured; DFR, Prac Sq.; Inj
	for _, status := range injuryStatuses {
		if w.StatusDescriptionAbbr == status {
			return true
		}
	}
	return false
}

// GetInjuryDescription returns a human-readable injury status
func (w *WeeklyRosterEntry) GetInjuryDescription() string {
	statusMap := map[string]string{
		"R01": "Reserve/Injured",
		"R04": "Reserve/PUP",
		"R48": "Reserve/Injured; DFR",
		"P02": "Practice Squad; Injured",
	}
	if desc, ok := statusMap[w.StatusDescriptionAbbr]; ok {
		return desc
	}
	if w.Status == "INA" {
		return "Inactive"
	}
	return "Active"
}
