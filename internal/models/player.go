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
	SeasonType string        `json:"season_type" bson:"season_type"` // REG, POST, REGPOST

	// Offensive Stats
	PassingYards   int `json:"passing_yards" bson:"passing_yards"`
	PassingTDs     int `json:"passing_tds" bson:"passing_tds"`
	Interceptions  int `json:"interceptions" bson:"interceptions"`
	RushingYards   int `json:"rushing_yards" bson:"rushing_yards"`
	RushingTDs     int `json:"rushing_tds" bson:"rushing_tds"`
	Receptions     int `json:"receptions" bson:"receptions"`
	ReceivingYards int `json:"receiving_yards" bson:"receiving_yards"`
	ReceivingTDs   int `json:"receiving_tds" bson:"receiving_tds"`
	Targets        int `json:"targets" bson:"targets"`

	// Defensive Stats
	Tackles          int     `json:"tackles" bson:"tackles"`
	TacklesSolo      int     `json:"tackles_solo" bson:"tackles_solo"`
	TacklesAssist    int     `json:"tackles_assist" bson:"tackles_assist"`
	TacklesForLoss   float64 `json:"tackles_for_loss" bson:"tackles_for_loss"`
	Sacks            float64 `json:"sacks" bson:"sacks"`
	SackYards        float64 `json:"sack_yards" bson:"sack_yards"`
	DefInterceptions int     `json:"def_interceptions" bson:"def_interceptions"` // Defensive INTs (different from QB INTs)
	PassDefended     int     `json:"pass_defended" bson:"pass_defended"`
	ForcedFumbles    int     `json:"forced_fumbles" bson:"forced_fumbles"`
	FumbleRecoveries int     `json:"fumble_recoveries" bson:"fumble_recoveries"`
	DefensiveTDs     int     `json:"defensive_tds" bson:"defensive_tds"`
	SafetyMD         int     `json:"safety_md" bson:"safety_md"` // Safeties

	// Performance Metrics (pre-calculated)
	EPA       float64 `json:"epa" bson:"epa"`               // Expected Points Added
	PlayCount int     `json:"play_count" bson:"play_count"` // Number of plays involved in

	// Fantasy Points
	FantasyPoints    float64 `json:"fantasy_points" bson:"fantasy_points"`         // Standard fantasy points
	FantasyPointsPPR float64 `json:"fantasy_points_ppr" bson:"fantasy_points_ppr"` // PPR fantasy points

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
	ID       bson.ObjectID `json:"id" bson:"_id,omitempty"`
	NFLID    string        `json:"nfl_id" bson:"nfl_id"`
	Week     int           `json:"week" bson:"week"`
	Season   int           `json:"season" bson:"season"`
	Opponent string        `json:"opponent" bson:"opponent"`

	// Passing Stats
	PassingYards  int `json:"passing_yards" bson:"passing_yards"`
	PassingTDs    int `json:"passing_tds" bson:"passing_tds"`
	Interceptions int `json:"interceptions" bson:"interceptions"`

	// Rushing Stats
	Carries      int `json:"carries" bson:"carries"`
	RushingYards int `json:"rushing_yards" bson:"rushing_yards"`
	RushingTDs   int `json:"rushing_tds" bson:"rushing_tds"`

	// Receiving Stats
	Receptions     int `json:"receptions" bson:"receptions"`
	Targets        int `json:"targets" bson:"targets"`
	ReceivingYards int `json:"receiving_yards" bson:"receiving_yards"`
	ReceivingTDs   int `json:"receiving_tds" bson:"receiving_tds"`

	// Performance Metrics
	EPA float64 `json:"epa" bson:"epa"`

	// Fantasy Points
	FantasyPoints    float64 `json:"fantasy_points" bson:"fantasy_points"`         // Standard fantasy points
	FantasyPointsPPR float64 `json:"fantasy_points_ppr" bson:"fantasy_points_ppr"` // PPR fantasy points

	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
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

// GetPlayerStatusDescription returns a human-readable status for any player status code
func GetPlayerStatusDescription(status, statusAbbr string) string {
	// Full status code mapping from NFLverse documentation
	statusMap := map[string]string{
		// Reserve statuses (injured)
		"R01": "Reserve/Injured",
		"R02": "Reserve/Retired",
		"R03": "Reserve/Left Squad",
		"R04": "Reserve/PUP",
		"R05": "Reserve/Military",
		"R06": "Reserve/Non-Football Injury",
		"R07": "Reserve/Suspended",
		"R08": "Reserve/Did Not Report",
		"R09": "Reserve/Commissioner Permission",
		"R48": "Reserve/Injured; DFR",

		// Practice Squad statuses
		"P01": "Practice Squad",
		"P02": "Practice Squad; Injured",
		"P03": "Practice Squad; Exempt",

		// Active statuses
		"A01": "Active",
		"A02": "Active/Physically Unable to Perform",
		"A03": "Active/Non-Football Injury",
		"A04": "Active/Commissioner Exempt",
		"A07": "Active/Suspended",

		// Waived statuses
		"W01": "Waived/Injured",
		"W03": "Waived/Injured; Settlement",

		// Other
		"E01": "Exempt/Left Squad",
	}

	if desc, ok := statusMap[statusAbbr]; ok {
		return desc
	}

	if status == "INA" {
		return "Inactive"
	} else if status == "ACT" {
		return "Active"
	}

	return "Active"
}
