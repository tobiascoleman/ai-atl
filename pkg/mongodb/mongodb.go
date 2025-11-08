package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect establishes a connection to MongoDB
func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(50).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

// CreateIndexes creates necessary MongoDB indexes for performance
func CreateIndexes(ctx context.Context, db *mongo.Database) error {
	// Players collection indexes
	playerIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"nfl_id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"team": 1, "position": 1},
		},
	}
	_, err := db.Collection("players").Indexes().CreateMany(ctx, playerIndexes)
	if err != nil {
		return err
	}

	// Games collection indexes
	gameIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"game_id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"season": 1, "week": 1},
		},
	}
	_, err = db.Collection("games").Indexes().CreateMany(ctx, gameIndexes)
	if err != nil {
		return err
	}

	// Plays collection indexes
	playIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"game_id": 1, "play_id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"season": 1, "week": 1},
		},
	}
	_, err = db.Collection("plays").Indexes().CreateMany(ctx, playIndexes)
	if err != nil {
		return err
	}

	// Users collection indexes
	userIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"email": 1},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err = db.Collection("users").Indexes().CreateMany(ctx, userIndexes)
	if err != nil {
		return err
	}

	// Lineups collection indexes
	lineupIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"user_id": 1, "week": 1},
		},
	}
	_, err = db.Collection("lineups").Indexes().CreateMany(ctx, lineupIndexes)
	if err != nil {
		return err
	}

	// Votes collection indexes
	voteIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"player_id": 1, "week": 1},
		},
	}
	_, err = db.Collection("votes").Indexes().CreateMany(ctx, voteIndexes)

	return err
}
