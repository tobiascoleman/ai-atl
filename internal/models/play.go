package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Play represents a single play from NFLverse play-by-play data
type Play struct {
	ID bson.ObjectID `json:"id" bson:"_id,omitempty"`

	// Game identifiers
	GameID    string `json:"game_id" bson:"game_id"`
	PlayID    string `json:"play_id" bson:"play_id"`
	Season    int    `json:"season" bson:"season"`
	Week      int    `json:"week" bson:"week"`
	
	// Play details
	Quarter        int    `json:"quarter" bson:"quarter"`
	Down           int    `json:"down" bson:"down"`
	YardsToGo      int    `json:"yards_to_go" bson:"yards_to_go"`
	YardLine       int    `json:"yard_line" bson:"yard_line"`
	GameSeconds    int    `json:"game_seconds" bson:"game_seconds"`
	Description    string `json:"description" bson:"description"`
	PlayType       string `json:"play_type" bson:"play_type"` // pass, run, punt, kickoff, etc.
	
	// Team data
	PossessionTeam string `json:"possession_team" bson:"possession_team"`
	DefenseTeam    string `json:"defense_team" bson:"defense_team"`
	
	// Player data
	PasserPlayerID   string `json:"passer_player_id" bson:"passer_player_id"`
	PasserPlayerName string `json:"passer_player_name" bson:"passer_player_name"`
	ReceiverPlayerID string `json:"receiver_player_id" bson:"receiver_player_id"`
	RusherPlayerID   string `json:"rusher_player_id" bson:"rusher_player_id"`
	
	// Outcome
	Yards         int     `json:"yards" bson:"yards"`
	Touchdown     bool    `json:"touchdown" bson:"touchdown"`
	Interception  bool    `json:"interception" bson:"interception"`
	Fumble        bool    `json:"fumble" bson:"fumble"`
	Sack          bool    `json:"sack" bson:"sack"`
	
	// Advanced metrics from NFLverse
	EPA           float64 `json:"epa" bson:"epa"`            // Expected Points Added
	WPA           float64 `json:"wpa" bson:"wpa"`            // Win Probability Added
	SuccessPlay   bool    `json:"success_play" bson:"success_play"`
	AirYards      int     `json:"air_yards" bson:"air_yards"`
	YardsAfterCatch int   `json:"yards_after_catch" bson:"yards_after_catch"`
	
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
