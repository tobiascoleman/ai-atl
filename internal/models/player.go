package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Player struct {
	ID       bson.ObjectID `json:"id" bson:"_id,omitempty"`
	NFLID    string             `json:"nfl_id" bson:"nfl_id"`
	Name     string             `json:"name" bson:"name"`
	Team     string             `json:"team" bson:"team"`
	Position string             `json:"position" bson:"position"`

	// Weekly stats from NFLverse
	WeeklyStats []WeeklyStat `json:"weekly_stats" bson:"weekly_stats"`

	// Advanced metrics from NFLverse
	EPAPerPlay   float64 `json:"epa_per_play" bson:"epa_per_play"`
	SuccessRate  float64 `json:"success_rate" bson:"success_rate"`
	SnapShare    float64 `json:"snap_share" bson:"snap_share"`
	TargetShare  float64 `json:"target_share" bson:"target_share"`

	// Injury data
	InjuryStatus  string                 `json:"injury_status" bson:"injury_status"`
	InjuryDetails map[string]interface{} `json:"injury_details" bson:"injury_details"`

	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type WeeklyStat struct {
	Week             int     `json:"week" bson:"week"`
	Season           int     `json:"season" bson:"season"`
	Opponent         string  `json:"opponent" bson:"opponent"`
	Yards            int     `json:"yards" bson:"yards"`
	Touchdowns       int     `json:"touchdowns" bson:"touchdowns"`
	Receptions       int     `json:"receptions" bson:"receptions"`
	Targets          int     `json:"targets" bson:"targets"`
	Carries          int     `json:"carries" bson:"carries"`
	PassingYards     int     `json:"passing_yards" bson:"passing_yards"`
	PassingTDs       int     `json:"passing_tds" bson:"passing_tds"`
	Interceptions    int     `json:"interceptions" bson:"interceptions"`
	EPA              float64 `json:"epa" bson:"epa"`
	ProjectedPoints  float64 `json:"projected_points" bson:"projected_points"`
	ActualPoints     float64 `json:"actual_points" bson:"actual_points"`
}

