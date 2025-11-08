package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Connect establishes a connection to MongoDB
func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	// Use ServerAPI for MongoDB Atlas compatibility
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	clientOptions := options.Client().
		ApplyURI(uri).
		SetServerAPIOptions(serverAPI).
		SetMaxPoolSize(50).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Second).
		SetConnectTimeout(30 * time.Second).        // Longer timeout for initial connection
		SetServerSelectionTimeout(30 * time.Second) // Longer timeout for Atlas

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}

// CreateIndexes creates necessary MongoDB indexes for performance
func CreateIndexes(ctx context.Context, db *mongo.Database) error {
	// Players collection indexes
	playerIndexes := []mongo.IndexModel{
		{
			// Compound unique index: player can have multiple entries (one per season)
			Keys:    bson.D{{"nfl_id", 1}, {"season", 1}}, // Use bson.D for ordered keys
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"team", 1}, {"position", 1}},
		},
		{
			Keys: bson.D{{"season", 1}},
		},
	}
	_, err := db.Collection("players").Indexes().CreateMany(ctx, playerIndexes)
	if err != nil {
		return err
	}

	// Player stats collection indexes
	playerStatsIndexes := []mongo.IndexModel{
		{
			// Compound unique index: one stats entry per player per season per season_type
			Keys:    bson.D{{"nfl_id", 1}, {"season", 1}, {"season_type", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"season", 1}},
		},
	}
	_, err = db.Collection("player_stats").Indexes().CreateMany(ctx, playerStatsIndexes)
	if err != nil {
		return err
	}

	// Games collection indexes
	gameIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"game_id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"season", 1}, {"week", 1}},
		},
	}
	_, err = db.Collection("games").Indexes().CreateMany(ctx, gameIndexes)
	if err != nil {
		return err
	}

	// Plays collection indexes
	playIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"game_id", 1}, {"play_id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"season", 1}, {"week", 1}},
		},
	}
	_, err = db.Collection("plays").Indexes().CreateMany(ctx, playIndexes)
	if err != nil {
		return err
	}

	// Users collection indexes
	userIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"email", 1}},
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
			Keys: bson.D{{"user_id", 1}, {"week", 1}},
		},
	}
	_, err = db.Collection("lineups").Indexes().CreateMany(ctx, lineupIndexes)
	if err != nil {
		return err
	}

	// Votes collection indexes
	voteIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"player_id", 1}, {"week", 1}},
		},
	}
	_, err = db.Collection("votes").Indexes().CreateMany(ctx, voteIndexes)

	return err
}
