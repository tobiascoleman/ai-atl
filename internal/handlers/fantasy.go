package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ai-atl/nfl-platform/internal/config"
	"github.com/ai-atl/nfl-platform/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type yahooStateClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type FantasyHandler struct {
	yahoo *services.YahooService
	cfg   *config.Config
}

func NewFantasyHandler(cfg *config.Config, yahooService *services.YahooService) *FantasyHandler {
	return &FantasyHandler{
		yahoo: yahooService,
		cfg:   cfg,
	}
}

func (h *FantasyHandler) GetAuthURL(c *gin.Context) {
	if !h.yahoo.Enabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "fantasy integration is not configured"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	state, err := h.buildState(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate oauth state"})
		return
	}

	url, err := h.yahoo.AuthCodeURL(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *FantasyHandler) Callback(c *gin.Context) {
	if !h.yahoo.Enabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "fantasy integration is not configured"})
		return
	}

	state := c.Query("state")
	code := c.Query("code")

	if state == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing state or code"})
		return
	}

	claims, err := h.parseState(state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	ctx := c.Request.Context()

	token, err := h.yahoo.Exchange(ctx, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("oauth exchange failed: %v", err)})
		return
	}

	userObjID, err := bson.ObjectIDFromHex(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user reference"})
		return
	}

	guid := ""
	if guidVal := token.Extra("xoauth_yahoo_guid"); guidVal != nil {
		guid = fmt.Sprintf("%v", guidVal)
	}

	if err := h.yahoo.SaveToken(ctx, userObjID, token, guid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	redirect := strings.TrimRight(h.cfg.ClientAppURL, "/") + "/dashboard/fantasy?connected=1"
	c.Redirect(http.StatusTemporaryRedirect, redirect)
}

func (h *FantasyHandler) Status(c *gin.Context) {
	if !h.yahoo.Enabled() {
		c.JSON(http.StatusOK, gin.H{
			"enabled":   false,
			"connected": false,
		})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	user, err := h.yahoo.LoadUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
		return
	}

	connected := user.YahooAccessToken != "" && user.YahooRefreshToken != ""

	c.JSON(http.StatusOK, gin.H{
		"enabled":   true,
		"connected": connected,
	})
}

func (h *FantasyHandler) Teams(c *gin.Context) {
	if !h.yahoo.Enabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "fantasy integration is not configured"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	user, err := h.yahoo.LoadUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
		return
	}

	if user.YahooAccessToken == "" || user.YahooRefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "yahoo account not connected"})
		return
	}

	token, err := h.yahoo.TokenFromUser(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refreshedToken, err := h.yahoo.RefreshIfNeeded(c.Request.Context(), user, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	teams, err := h.yahoo.FetchTeams(c.Request.Context(), refreshedToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connected": true,
		"teams":     teams,
	})
}

func (h *FantasyHandler) buildState(userID string) (string, error) {
	nonce, err := randomNonce(16)
	if err != nil {
		return "", err
	}

	claims := yahooStateClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "yahoo_oauth_state",
			ID:        nonce,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	state, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return state, nil
}

func (h *FantasyHandler) parseState(state string) (*yahooStateClaims, error) {
	claims := &yahooStateClaims{}
	token, err := jwt.ParseWithClaims(state, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(h.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if token == nil || !token.Valid {
		return nil, errors.New("invalid state token")
	}

	if claims.UserID == "" {
		return nil, errors.New("state missing user id")
	}

	return claims, nil
}

func randomNonce(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
