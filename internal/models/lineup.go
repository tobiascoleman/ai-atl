package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type FantasyLineup struct {
	ID     bson.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID bson.ObjectID `json:"user_id" bson:"user_id"`

	Week   int `json:"week" bson:"week"`
	Season int `json:"season" bson:"season"`

	// Positions map: QB, RB1, RB2, WR1, WR2, WR3, TE, FLEX, K, DEF
	Positions map[string]string `json:"positions" bson:"positions"` // position -> player_id

	ProjectedPoints float64 `json:"projected_points" bson:"projected_points"`
	ActualPoints    float64 `json:"actual_points" bson:"actual_points"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

