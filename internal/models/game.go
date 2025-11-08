package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Game struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	GameID    string             `json:"game_id" bson:"game_id"`
	Season    int                `json:"season" bson:"season"`
	Week      int                `json:"week" bson:"week"`
	HomeTeam  string             `json:"home_team" bson:"home_team"`
	AwayTeam  string             `json:"away_team" bson:"away_team"`
	StartTime time.Time          `json:"start_time" bson:"start_time"`
	Status    string             `json:"status" bson:"status"` // scheduled, live, final

	// Betting data from NFLverse
	VegasLine   float64 `json:"vegas_line" bson:"vegas_line"`
	OverUnder   float64 `json:"over_under" bson:"over_under"`

	// Scores
	HomeScore int `json:"home_score" bson:"home_score"`
	AwayScore int `json:"away_score" bson:"away_score"`

	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

