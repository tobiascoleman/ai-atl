package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// NextGenStat represents Next Gen Stats for a player
// These are advanced metrics from NFL's tracking data
type NextGenStat struct {
	ID       bson.ObjectID `json:"id" bson:"_id,omitempty"`
	PlayerID string        `json:"player_id" bson:"player_id"` // gsis_id
	Season   int           `json:"season" bson:"season"`
	Week     int           `json:"week" bson:"week"` // 0 for season totals
	StatType string        `json:"stat_type" bson:"stat_type"` // passing, rushing, receiving

	// Common fields
	PlayerName       string `json:"player_name" bson:"player_name"`
	Team             string `json:"team" bson:"team"`
	Position         string `json:"position" bson:"position"`
	PlayerGamesPlayed int   `json:"player_game_played" bson:"player_game_played"`

	// Passing NGS (stat_type: "passing")
	PassAttempts                   int     `json:"pass_attempts" bson:"pass_attempts"`
	PassCompletions                int     `json:"pass_completions" bson:"pass_completions"`
	PassYards                      int     `json:"pass_yards" bson:"pass_yards"`
	PassTouchdowns                 int     `json:"pass_touchdowns" bson:"pass_touchdowns"`
	Interceptions                  int     `json:"interceptions" bson:"interceptions"`
	CompletionPercentageAboveExpectation float64 `json:"completion_percentage_above_expectation" bson:"completion_percentage_above_expectation"`
	AvgTimeToThrow                 float64 `json:"avg_time_to_throw" bson:"avg_time_to_throw"`
	AvgCompletedAirYards           float64 `json:"avg_completed_air_yards" bson:"avg_completed_air_yards"`
	AvgIntendedAirYards            float64 `json:"avg_intended_air_yards" bson:"avg_intended_air_yards"`
	AvgAirYardsDifferential        float64 `json:"avg_air_yards_differential" bson:"avg_air_yards_differential"`
	MaxCompletedAirDistance        float64 `json:"max_completed_air_distance" bson:"max_completed_air_distance"`
	MaxAirDistance                 float64 `json:"max_air_distance" bson:"max_air_distance"`

	// Rushing NGS (stat_type: "rushing")
	Carries              int     `json:"carries" bson:"carries"`
	RushYards            int     `json:"rush_yards" bson:"rush_yards"`
	RushTouchdowns       int     `json:"rush_touchdowns" bson:"rush_touchdowns"`
	ExpectedRushYards    float64 `json:"expected_rush_yards" bson:"expected_rush_yards"`
	RushYardsOverExpected float64 `json:"rush_yards_over_expected" bson:"rush_yards_over_expected"`
	AvgTimeToLOS         float64 `json:"avg_time_to_los" bson:"avg_time_to_los"` // Line of scrimmage
	RushPct8Defenders    float64 `json:"rush_pct_8_defenders" bson:"rush_pct_8_defenders"`
	Efficiency           float64 `json:"efficiency" bson:"efficiency"`

	// Receiving NGS (stat_type: "receiving")
	Receptions             int     `json:"receptions" bson:"receptions"`
	Targets                int     `json:"targets" bson:"targets"`
	ReceivingYards         int     `json:"receiving_yards" bson:"receiving_yards"`
	ReceivingTouchdowns    int     `json:"receiving_touchdowns" bson:"receiving_touchdowns"`
	AvgCushion             float64 `json:"avg_cushion" bson:"avg_cushion"`
	AvgSeparation          float64 `json:"avg_separation" bson:"avg_separation"`
	AvgIntendedAirYardsRec float64 `json:"avg_intended_air_yards_rec" bson:"avg_intended_air_yards_rec"`
	CatchPercentage        float64 `json:"catch_percentage" bson:"catch_percentage"`
	ShareOfTeamTargets     float64 `json:"share_of_team_targets" bson:"share_of_team_targets"`
	AvgYAC                 float64 `json:"avg_yac" bson:"avg_yac"` // Yards after catch
	AvgExpectedYAC         float64 `json:"avg_expected_yac" bson:"avg_expected_yac"`
	AvgYACAboveExpectation float64 `json:"avg_yac_above_expectation" bson:"avg_yac_above_expectation"`

	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

