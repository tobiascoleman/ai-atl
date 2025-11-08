package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Vote struct {
	ID     bson.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID bson.ObjectID `json:"user_id" bson:"user_id"`

	PlayerID       string  `json:"player_id" bson:"player_id"`
	PredictionType string  `json:"prediction_type" bson:"prediction_type"` // over, under, lock, fade
	StatLine       float64 `json:"stat_line" bson:"stat_line"`
	Week           int     `json:"week" bson:"week"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

