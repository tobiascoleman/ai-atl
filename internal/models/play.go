package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Play represents a single play from NFLverse historical data
type Play struct {
	ID     bson.ObjectID `json:"id" bson:"_id,omitempty"`
	PlayID string             `json:"play_id" bson:"play_id"`
	GameID string             `json:"game_id" bson:"game_id"`
	Season int                `json:"season" bson:"season"`
	Week   int                `json:"week" bson:"week"`

	// Play details
	PlayType    string  `json:"play_type" bson:"play_type"`
	YardsGained int     `json:"yards_gained" bson:"yards_gained"`
	EPA         float64 `json:"epa" bson:"epa"`
	WPA         float64 `json:"wpa" bson:"wpa"`
	Success     bool    `json:"success" bson:"success"`

	// Game script context
	ScoreDifferential int `json:"score_differential" bson:"score_differential"`
	Quarter           int `json:"quarter" bson:"quarter"`
	TimeRemaining     int `json:"time_remaining" bson:"time_remaining"`
	Down              int `json:"down" bson:"down"`
	Distance          int `json:"distance" bson:"distance"`

	// Team and player info
	PossessionTeam string `json:"possession_team" bson:"possession_team"`
	DefenseTeam    string `json:"defense_team" bson:"defense_team"`
}

