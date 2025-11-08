package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID                bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Email             string        `json:"email" bson:"email"`
	Username          string        `json:"username" bson:"username"`
	Password          string        `json:"-" bson:"password"` // Password hash, never send in JSON
	CreatedAt         time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at" bson:"updated_at"`
	YahooAccessToken  string        `json:"-" bson:"yahoo_access_token,omitempty"`
	YahooRefreshToken string        `json:"-" bson:"yahoo_refresh_token,omitempty"`
	YahooTokenExpiry  time.Time     `json:"-" bson:"yahoo_token_expiry,omitempty"`
	YahooGuid         string        `json:"-" bson:"yahoo_guid,omitempty"`
}

// UserResponse is used for API responses (excludes password)
type UserResponse struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	CreatedAt      time.Time `json:"created_at"`
	YahooConnected bool      `json:"yahoo_connected"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:             u.ID.Hex(),
		Email:          u.Email,
		Username:       u.Username,
		CreatedAt:      u.CreatedAt,
		YahooConnected: u.YahooAccessToken != "",
	}
}
