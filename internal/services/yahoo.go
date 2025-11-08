package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ai-atl/nfl-platform/internal/config"
	"github.com/ai-atl/nfl-platform/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/oauth2"
)

type YahooTeam struct {
	TeamKey    string `json:"team_key"`
	Name       string `json:"name"`
	LeagueName string `json:"league_name"`
	LogoURL    string `json:"logo_url,omitempty"`
}

type YahooService struct {
	db          *mongo.Database
	oauthConfig *oauth2.Config
	httpClient  *http.Client
}

func NewYahooService(db *mongo.Database, cfg *config.Config) *YahooService {
	var oauthCfg *oauth2.Config
	if cfg.YahooClientID != "" && cfg.YahooClientSecret != "" && cfg.YahooRedirectURL != "" {
		oauthCfg = &oauth2.Config{
			ClientID:     cfg.YahooClientID,
			ClientSecret: cfg.YahooClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://api.login.yahoo.com/oauth2/request_auth",
				TokenURL: "https://api.login.yahoo.com/oauth2/get_token",
			},
			RedirectURL: cfg.YahooRedirectURL,
			Scopes:      []string{"fspt-r"},
		}
	}

	return &YahooService{
		db:          db,
		oauthConfig: oauthCfg,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *YahooService) Enabled() bool {
	return s.oauthConfig != nil
}

func (s *YahooService) AuthCodeURL(state string) (string, error) {
	if s.oauthConfig == nil {
		return "", errors.New("yahoo oauth not configured")
	}

	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

func (s *YahooService) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	if s.oauthConfig == nil {
		return nil, errors.New("yahoo oauth not configured")
	}

	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return token, nil
}

func (s *YahooService) TokenFromUser(user *models.User) (*oauth2.Token, error) {
	if user.YahooAccessToken == "" || user.YahooRefreshToken == "" {
		return nil, errors.New("user not linked to yahoo")
	}

	return &oauth2.Token{
		AccessToken:  user.YahooAccessToken,
		RefreshToken: user.YahooRefreshToken,
		Expiry:       user.YahooTokenExpiry,
		TokenType:    "bearer",
	}, nil
}

func (s *YahooService) RefreshIfNeeded(ctx context.Context, user *models.User, token *oauth2.Token) (*oauth2.Token, error) {
	if s.oauthConfig == nil {
		return nil, errors.New("yahoo oauth not configured")
	}

	source := s.oauthConfig.TokenSource(ctx, token)
	refreshedToken, err := source.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Only update if token actually changed/extended.
	if refreshedToken.AccessToken != user.YahooAccessToken || refreshedToken.RefreshToken != "" && refreshedToken.RefreshToken != user.YahooRefreshToken {
		update := bson.M{
			"$set": bson.M{
				"yahoo_access_token": refreshedToken.AccessToken,
				"yahoo_token_expiry": refreshedToken.Expiry,
				"updated_at":         time.Now(),
			},
		}

		if refreshedToken.RefreshToken != "" {
			update["$set"].(bson.M)["yahoo_refresh_token"] = refreshedToken.RefreshToken
		}

		if _, err := s.db.Collection("users").UpdateByID(ctx, user.ID, update); err != nil {
			return nil, fmt.Errorf("failed to persist refreshed token: %w", err)
		}

		user.YahooAccessToken = refreshedToken.AccessToken
		if refreshedToken.RefreshToken != "" {
			user.YahooRefreshToken = refreshedToken.RefreshToken
		}
		user.YahooTokenExpiry = refreshedToken.Expiry
	}

	return refreshedToken, nil
}

func (s *YahooService) SaveToken(ctx context.Context, userID bson.ObjectID, token *oauth2.Token, guid string) error {
	if s.oauthConfig == nil {
		return errors.New("yahoo oauth not configured")
	}

	update := bson.M{
		"$set": bson.M{
			"yahoo_access_token":  token.AccessToken,
			"yahoo_refresh_token": token.RefreshToken,
			"yahoo_token_expiry":  token.Expiry,
			"yahoo_guid":          guid,
			"updated_at":          time.Now(),
		},
	}

	_, err := s.db.Collection("users").UpdateByID(ctx, userID, update)
	if err != nil {
		return fmt.Errorf("failed to store yahoo tokens: %w", err)
	}

	return nil
}

func (s *YahooService) LoadUser(ctx context.Context, userID string) (*models.User, error) {
	objID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	var user models.User
	if err := s.db.Collection("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *YahooService) FetchTeams(ctx context.Context, token *oauth2.Token) ([]YahooTeam, error) {
	if s.oauthConfig == nil {
		return nil, errors.New("yahoo oauth not configured")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://fantasysports.yahooapis.com/fantasy/v2/users;use_login=1/games;game_keys=nfl/teams?format=json",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query yahoo api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo api returned status %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode yahoo response: %w", err)
	}

	teams := extractTeams(payload)

	return teams, nil
}

func extractTeams(payload map[string]any) []YahooTeam {
	fantasyContent := toMap(payload["fantasy_content"])
	users := toMap(fantasyContent["users"])
	if users == nil {
		return nil
	}

	var teams []YahooTeam

	for key, userEntry := range users {
		if key == "count" {
			continue
		}

		entryMap := toMap(userEntry)
		if entryMap == nil {
			continue
		}

		userSlice := toSlice(entryMap["user"])
		for _, userItem := range userSlice {
			itemMap := toMap(userItem)
			if itemMap == nil {
				continue
			}

			gamesMap := toMap(itemMap["games"])
			if gamesMap == nil {
				continue
			}

			gameCount := toInt(gamesMap["count"])
			for i := 0; i < gameCount; i++ {
				gameWrapper := toMap(gamesMap[strconv.Itoa(i)])
				if gameWrapper == nil {
					continue
				}

				gameSlice := toSlice(gameWrapper["game"])
				var leagueName string

				for _, gameItem := range gameSlice {
					gm := toMap(gameItem)
					if gm == nil {
						continue
					}

					if name, ok := gm["name"].(string); ok && leagueName == "" {
						leagueName = name
					}

					teamsMap := toMap(gm["teams"])
					if teamsMap == nil {
						continue
					}

					teamCount := toInt(teamsMap["count"])
					for ti := 0; ti < teamCount; ti++ {
						teamWrapper := toMap(teamsMap[strconv.Itoa(ti)])
						if teamWrapper == nil {
							continue
						}

						teamSlice := toSlice(teamWrapper["team"])
						team := YahooTeam{LeagueName: leagueName}

						for _, teamPart := range teamSlice {
							tm := toMap(teamPart)
							if tm == nil {
								continue
							}

							if key, ok := tm["team_key"].(string); ok {
								team.TeamKey = key
							}

							if name, ok := tm["name"].(string); ok {
								team.Name = name
							}

							if logoMap := toMap(tm["team_logos"]); logoMap != nil {
								logoCount := toInt(logoMap["count"])
								if logoCount > 0 {
									firstLogo := toMap(logoMap["0"])
									if firstLogo != nil {
										logoSlice := toSlice(firstLogo["team_logo"])
										for _, logo := range logoSlice {
											if logoMap := toMap(logo); logoMap != nil {
												if url, ok := logoMap["url"].(string); ok {
													team.LogoURL = url
												}
											}
										}
									}
								}
							}
						}

						if team.TeamKey != "" && team.Name != "" {
							teams = append(teams, team)
						}
					}
				}
			}
		}
	}

	return dedupeTeams(teams)
}

func toMap(v any) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func toSlice(v any) []any {
	if v == nil {
		return nil
	}
	if s, ok := v.([]any); ok {
		return s
	}
	return nil
}

func toInt(v any) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case string:
		n, _ := strconv.Atoi(val)
		return n
	default:
		return 0
	}
}

func dedupeTeams(teams []YahooTeam) []YahooTeam {
	if len(teams) == 0 {
		return teams
	}

	seen := make(map[string]int)
	var result []YahooTeam

	for _, team := range teams {
		key := strings.ToLower(team.TeamKey)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = 1
		result = append(result, team)
	}

	return result
}
