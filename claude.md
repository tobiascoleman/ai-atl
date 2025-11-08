# AI-ATL NFL Platform - Development Guide

**Project**: NFL Fantasy & Analytics Platform with AI-Powered Insights  
**Timeline**: 36-hour hackathon  
**Stack**: Go API + MongoDB + Gemini AI + React/Next.js  
**Data Source**: NFLverse (https://github.com/nflverse)

---

## 🎯 Project Overview

Building an NFL fantasy platform with **AI Game Script Prediction** as our primary differentiator. Uses NFLverse's rich historical data (play-by-play, EPA metrics) combined with Gemini AI to predict how games will unfold and impact fantasy players.

### Key Features (Priority Order)
1. ✅ **AI Game Script Predictor** (NOVEL - our main selling point)
2. ✅ **AI Chatbot** (fantasy advice with context)
3. ✅ Fantasy lineup manager with optimization
4. ✅ Trade analyzer with EPA metrics
5. ✅ Waiver wire recommendations
6. Social voting (if time permits)

### Competitive Advantages
- NFLverse EPA/WPA data (efficiency metrics others don't use)
- AI-powered game flow prediction (not just scores)
- Historical pattern matching across 3+ years of play data
- Quantified injury impact predictions

---

## 🏗️ Architecture

### Tech Stack
```
Backend:  Go 1.21+ with Gin framework
Database: MongoDB via official Go driver
AI:       Google Gemini API
Cache:    Redis
Jobs:     Go routines / cron jobs
Frontend: React/Next.js (separate repo)
Auth:     JWT with golang-jwt
Data:     NFLverse (Parquet files via Apache Arrow Go)
```

### Project Structure
```
cmd/
└── api/
    └── main.go          # Application entry point
internal/
├── models/              # Data models
│   ├── user.go
│   ├── player.go
│   ├── game.go
│   ├── play.go         # Historical play data
│   ├── lineup.go
│   └── vote.go
├── handlers/           # HTTP handlers (controllers)
│   ├── auth.go
│   ├── players.go
│   ├── lineups.go
│   ├── insights.go
│   ├── trades.go
│   └── chatbot.go
├── services/           # Business logic
│   ├── game_script.go
│   ├── epa_analyzer.go
│   ├── injury_analyzer.go
│   └── chatbot.go
├── middleware/
│   ├── auth.go
│   ├── cors.go
│   └── logger.go
└── config/
    └── config.go
pkg/
├── nflverse/           # NFLverse data fetching
│   └── client.go
├── gemini/             # Gemini API client
│   └── client.go
└── mongodb/
    └── mongodb.go
```

---

## 📋 Coding Standards (HACKATHON MODE)

### General Principles
- **Speed > Perfection**: Working features beat perfect code
- **MVP mindset**: Get core features working, polish later
- **Fail fast**: Test critical paths immediately
- **Document as you go**: Add comments for complex logic
- **Git often**: Commit working code frequently

### Go Conventions

#### Service Pattern
```go
// Good: Services handle complex business logic
type GameScriptPredictor struct {
	db     *mongo.Database
	gemini *gemini.Client
}

func NewGameScriptPredictor(db *mongo.Database) *GameScriptPredictor {
	return &GameScriptPredictor{
		db:     db,
		gemini: gemini.NewClient(),
	}
}

func (gsp *GameScriptPredictor) Predict(ctx context.Context, gameID string) (*GameScript, error) {
	game, err := gsp.getGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	
	context := gsp.buildContext(game)
	prompt := gsp.buildPrompt(context)
	return gsp.gemini.Generate(ctx, prompt)
}

// Usage in handler
func (h *InsightHandler) GameScript(c *gin.Context) {
	predictor := services.NewGameScriptPredictor(h.db)
	result, err := predictor.Predict(c.Request.Context(), c.Query("game_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
```

#### MongoDB Patterns
```go
// Good: Use embedded documents for nested data
type Player struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	WeeklyStats []WeeklyStat       `bson:"weekly_stats" json:"weekly_stats"`
	EPAPerPlay  float64            `bson:"epa_per_play" json:"epa_per_play"`
}

// Queries with MongoDB Go driver
filter := bson.M{
	"team":     "KC",
	"position": "WR",
}
opts := options.Find().SetSort(bson.D{{"epa_per_play", -1}})
cursor, err := collection.Find(ctx, filter, opts)
```

#### Error Handling (Keep It Simple)
```go
// Good: Return errors, let caller handle
func (s *PlayerService) GetPlayer(ctx context.Context, id string) (*Player, error) {
	var player Player
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid player ID: %w", err)
	}
	
	err = s.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&player)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("player not found")
		}
		return nil, err
	}
	return &player, nil
}

// For API handlers
func (h *PlayerHandler) Get(c *gin.Context) {
	player, err := h.service.GetPlayer(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, player)
}
```

### Gemini API Patterns

#### Prompt Engineering
```go
// Good: Structured prompts with clear context
func (s *GameScriptService) buildPrompt(game models.Game) string {
	return fmt.Sprintf(`You are an NFL fantasy expert analyzing game scripts.

Game Details:
- Teams: %s vs %s
- Vegas Line: %.1f
- Over/Under: %.1f

Task: Predict how this game will flow quarter by quarter.
Focus on: Pass/run ratios, player usage, scoring pace.

Respond in JSON format:
{
  "game_flow": "description",
  "player_impacts": [
    {"player": "Name", "impact": "+30%% touches", "reasoning": "why"}
  ]
}`, game.HomeTeam, game.AwayTeam, game.VegasLine, game.OverUnder)
}

// Parse response
func parseResponse(response string) (*GameScript, error) {
	var result GameScript
	err := json.Unmarshal([]byte(response), &result)
	if err != nil {
		// Fallback if AI doesn't return valid JSON
		return &GameScript{GameFlow: response}, nil
	}
	return &result, nil
}
```

#### Rate Limiting & Caching
```go
// Cache Gemini responses (identical queries)
func (c *Client) GenerateWithCache(ctx context.Context, prompt string) (string, error) {
	hash := md5.Sum([]byte(prompt))
	cacheKey := fmt.Sprintf("gemini:%x", hash)
	
	// Check cache first
	if cached, found := cache.Get(cacheKey); found {
		return cached.(string), nil
	}
	
	// Generate and cache
	response, err := c.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}
	
	cache.Set(cacheKey, response, 1*time.Hour)
	return response, nil
}
```

### NFLverse Data Patterns

#### Loading Data
```go
// Good: Download and cache parquet files
type NflverseService struct {
	httpClient *http.Client
	cacheDir   string
}

func (n *NflverseService) FetchPlayerStats(ctx context.Context, season int) ([]PlayerStat, error) {
	cacheKey := fmt.Sprintf("nflverse:player_stats:%d", season)
	
	// Check cache first
	if cached, found := cache.Get(cacheKey); found {
		return cached.([]PlayerStat), nil
	}
	
	url := fmt.Sprintf("https://github.com/nflverse/nflverse-data/releases/download/player_stats/player_stats_%d.parquet", season)
	stats, err := n.downloadAndParse(ctx, url)
	if err != nil {
		return nil, err
	}
	
	// Cache for 24 hours
	cache.Set(cacheKey, stats, 24*time.Hour)
	return stats, nil
}

func (n *NflverseService) downloadAndParse(ctx context.Context, url string) ([]PlayerStat, error) {
	// Download parquet file
	resp, err := n.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Parse with Arrow (implementation details)
	// Return parsed data as slice
	return parseParquetData(resp.Body)
}
```

#### Processing Large Datasets
```go
// Good: Batch insert for performance
func (s *SyncService) SyncPlayByPlay(ctx context.Context, plays []models.Play) error {
	collection := s.db.Collection("plays")
	batchSize := 1000
	
	for i := 0; i < len(plays); i += batchSize {
		end := i + batchSize
		if end > len(plays) {
			end = len(plays)
		}
		
		batch := plays[i:end]
		docs := make([]interface{}, len(batch))
		for j, play := range batch {
			docs[j] = play
		}
		
		_, err := collection.InsertMany(ctx, docs)
		if err != nil {
			return fmt.Errorf("batch insert failed: %w", err)
		}
	}
	return nil
}
```

---

## 🧪 Testing Strategy (HACKATHON EDITION)

### What to Test
**Priority 1 (MUST TEST)**:
- [ ] User authentication (register/login)
- [ ] Gemini API integration (can generate responses)
- [ ] NFLverse data fetching (can download & parse)
- [ ] Game Script Predictor (core feature)
- [ ] Chatbot (core feature)

**Priority 2 (If Time)**:
- [ ] Lineup creation/retrieval
- [ ] Trade analyzer calculations
- [ ] Background job execution

**Priority 3 (Nice to Have)**:
- [ ] Edge cases
- [ ] Full test coverage

### Quick Testing Patterns

#### Go Testing
```go
// Test NFLverse integration
func TestNFLverseIntegration(t *testing.T) {
	service := NewNflverseService()
	stats, err := service.FetchPlayerStats(context.Background(), 2024)
	if err != nil {
		t.Fatalf("Failed to fetch stats: %v", err)
	}
	t.Logf("Loaded %d player records", len(stats))
}

// Test Gemini
func TestGemini(t *testing.T) {
	client := gemini.NewClient()
	response, err := client.Generate(context.Background(), "Who won Super Bowl LVIII?")
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}
	t.Logf("Response: %s", response)
}

// Test Game Script Predictor
func TestGameScriptPredictor(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	service := services.NewGameScriptService(db)
	
	result, err := service.PredictGameScript(context.Background(), "some_game")
	if err != nil {
		t.Fatalf("Failed to predict: %v", err)
	}
	t.Logf("Prediction: %+v", result)
}
```

#### Manual API Testing (cURL)
```bash
# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","username":"testuser","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"password123"}'

# Test chatbot (with token)
curl -X POST http://localhost:8080/api/v1/chatbot/ask \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"question":"Who should I start at RB this week?"}'
```

#### Frontend Integration Testing
```javascript
// Test API calls from frontend
const testAPI = async () => {
  // Login
  const loginRes = await fetch('/api/v1/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: 'test@test.com', password: 'password123' })
  });
  const { token } = await loginRes.json();
  
  // Test game script predictor
  const scriptRes = await fetch('/api/v1/insights/game_script?game_id=123', {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  const prediction = await scriptRes.json();
  console.log(prediction);
};
```

---

## 🚀 Development Workflow

### Setup (First Time)
```bash
# Backend
git clone <repo>
cd nfl-platform
go mod download

# Start MongoDB and Redis
# MongoDB: docker run -d -p 27017:27017 mongo
# Redis: docker run -d -p 6379:6379 redis

# Run the API
go run cmd/api/main.go

# Or build and run
go build -o nfl-api cmd/api/main.go
./nfl-api

# Frontend (separate terminal)
cd frontend
npm install
npm run dev
```

### Environment Variables
Create `.env` file:
```
MONGODB_URI=mongodb://localhost:27017/nfl_platform_dev
REDIS_URL=redis://localhost:6379/0
GEMINI_API_KEY=your_key_here
JWT_SECRET=your_secret_here
NFLVERSE_CACHE_DIR=/tmp/nflverse
```

### Git Workflow (Fast Iterations)
```bash
# Create feature branch
git checkout -b feature/game-script-predictor

# Commit working code frequently
git add .
git commit -m "Add basic game script predictor service"
git push origin feature/game-script-predictor

# Merge to main when feature works
git checkout main
git merge feature/game-script-predictor
git push origin main
```

### Daily Standup (Async)
Post in team chat:
- ✅ What I completed
- 🚧 What I'm working on now
- 🚨 Blockers (if any)

---

## 📊 Data Models

### User
```go
type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email     string             `json:"email" bson:"email"`
	Username  string             `json:"username" bson:"username"`
	Password  string             `json:"-" bson:"password"` // Password hash, never send in JSON
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// Indexes created in mongodb package:
// - email: unique index
```

### Player
```go
type Player struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	NFLID    string             `json:"nfl_id" bson:"nfl_id"`
	Name     string             `json:"name" bson:"name"`
	Team     string             `json:"team" bson:"team"`
	Position string             `json:"position" bson:"position"`

	// Weekly stats from NFLverse
	WeeklyStats []WeeklyStat `json:"weekly_stats" bson:"weekly_stats"`

	// Advanced metrics from NFLverse
	EPAPerPlay   float64 `json:"epa_per_play" bson:"epa_per_play"`
	SuccessRate  float64 `json:"success_rate" bson:"success_rate"`
	SnapShare    float64 `json:"snap_share" bson:"snap_share"`
	TargetShare  float64 `json:"target_share" bson:"target_share"`

	// Injury data
	InjuryStatus  string                 `json:"injury_status" bson:"injury_status"`
	InjuryDetails map[string]interface{} `json:"injury_details" bson:"injury_details"`

	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type WeeklyStat struct {
	Week             int     `json:"week" bson:"week"`
	Season           int     `json:"season" bson:"season"`
	Yards            int     `json:"yards" bson:"yards"`
	Touchdowns       int     `json:"touchdowns" bson:"touchdowns"`
	EPA              float64 `json:"epa" bson:"epa"`
	ProjectedPoints  float64 `json:"projected_points" bson:"projected_points"`
	ActualPoints     float64 `json:"actual_points" bson:"actual_points"`
}

// Indexes: nfl_id (unique), team+position (compound)
```

### Game
```go
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

// Indexes: game_id (unique), season+week (compound)
```

### Play (Historical Data)
```go
type Play struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
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
}

// Indexes: game_id+play_id (unique compound), season+week (compound)
```

### FantasyLineup
```go
type FantasyLineup struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`

	Week   int `json:"week" bson:"week"`
	Season int `json:"season" bson:"season"`

	// Positions map: QB, RB1, RB2, WR1, WR2, WR3, TE, FLEX, K, DEF
	Positions map[string]string `json:"positions" bson:"positions"` // position -> player_id

	ProjectedPoints float64 `json:"projected_points" bson:"projected_points"`
	ActualPoints    float64 `json:"actual_points" bson:"actual_points"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// Indexes: user_id+week (compound)
```

---

## 🎨 API Design

### Authentication
```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
DELETE /api/v1/auth/logout
```

### Players
```
GET    /api/v1/players
GET    /api/v1/players/:id
GET    /api/v1/players/:id/stats?season=2024&week=9
```

### Fantasy Lineups
```
GET    /api/v1/lineups
POST   /api/v1/lineups
GET    /api/v1/lineups/:id
PUT    /api/v1/lineups/:id
DELETE /api/v1/lineups/:id
POST   /api/v1/lineups/optimize  # AI optimization
```

### Insights (Core Features)
```
GET    /api/v1/insights/game_script?game_id=123
POST   /api/v1/insights/injury_impact
       Body: { player_id: "123" }
GET    /api/v1/insights/streaks?player_id=123
GET    /api/v1/insights/top_performers?week=9&type=over
GET    /api/v1/insights/waiver_gems
```

### Trade Analyzer
```
POST   /api/v1/trades/analyze
       Body: {
         team_a_gives: ["player1", "player2"],
         team_a_gets: ["player3"],
         team_b_gives: ["player3"],
         team_b_gets: ["player1", "player2"]
       }
```

### Chatbot
```
POST   /api/v1/chatbot/ask
       Body: { question: "Who should I start at RB?" }
GET    /api/v1/chatbot/history
```

### Social/Voting
```
POST   /api/v1/votes
       Body: { player_id: "123", prediction: "over", stat_line: 85.5 }
GET    /api/v1/votes/consensus?player_id=123&week=9
```

---

## 🔧 Common Tasks

### Adding a New Service
```go
// 1. Create service file
// internal/services/new_feature.go
package services

type NewFeatureService struct {
	db *mongo.Database
}

func NewNewFeatureService(db *mongo.Database) *NewFeatureService {
	return &NewFeatureService{db: db}
}

func (s *NewFeatureService) Perform(ctx context.Context, param string) (*Result, error) {
	// Service logic here
	return &Result{}, nil
}

// 2. Create handler
// internal/handlers/new_feature.go
package handlers

type NewFeatureHandler struct {
	service *services.NewFeatureService
}

func NewNewFeatureHandler(db *mongo.Database) *NewFeatureHandler {
	return &NewFeatureHandler{
		service: services.NewNewFeatureService(db),
	}
}

func (h *NewFeatureHandler) Create(c *gin.Context) {
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	result, err := h.service.Perform(c.Request.Context(), req.Param)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// 3. Add route in cmd/api/main.go
newFeature := protected.Group("/new_feature")
{
	handler := handlers.NewNewFeatureHandler(mongoClient.Database(cfg.DBName))
	newFeature.POST("", handler.Create)
}
```

### Adding a Background Job
```go
// 1. Create job function
// internal/jobs/sync_job.go
package jobs

func SyncData(ctx context.Context, db *mongo.Database, param string) error {
	// Job logic here
	return nil
}

// 2. Trigger job with goroutine
go func() {
	if err := jobs.SyncData(context.Background(), db, param); err != nil {
		log.Printf("Job failed: %v", err)
	}
}()

// 3. Or use a cron scheduler
import "github.com/robfig/cron/v3"

c := cron.New()
c.AddFunc("0 * * * *", func() { // Every hour
	jobs.SyncData(context.Background(), db, "param")
})
c.Start()
```

### Querying NFLverse Data
```go
// Find players by EPA
ctx := context.Background()
filter := bson.M{"position": "WR"}
opts := options.Find().
	SetSort(bson.D{{"epa_per_play", -1}}).
	SetLimit(10)

cursor, err := db.Collection("players").Find(ctx, filter, opts)
var topReceivers []models.Player
cursor.All(ctx, &topReceivers)

// Find plays in specific game script
filter = bson.M{
	"season": 2024,
	"score_differential": bson.M{"$gte": -7, "$lte": 7},
	"quarter": bson.M{"$gte": 3},
}
cursor, err = db.Collection("plays").Find(ctx, filter)

// Aggregate stats
pipeline := mongo.Pipeline{
	{{Key: "$match", Value: bson.M{"team": "KC", "season": 2024}}},
	{{Key: "$group", Value: bson.M{
		"_id": nil,
		"avg_epa": bson.M{"$avg": "$epa"},
	}}},
}
cursor, err = db.Collection("plays").Aggregate(ctx, pipeline)
```

---

## 🐛 Debugging Guide

### Application Logs
```bash
# Run with verbose logging
go run cmd/api/main.go 2>&1 | tee app.log

# Watch logs in real-time
tail -f app.log

# Filter for errors
tail -f app.log | grep "error"
```

### Go Debugging
```go
// Add debug logging
import "log"

log.Printf("Debug: variable value = %+v", variable)
log.Printf("Error occurred: %v", err)

// Use delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug cmd/api/main.go
```

### Testing Individual Functions
```bash
# Run specific test
go test ./internal/services -run TestGameScriptPredictor -v

# Test with coverage
go test ./... -cover

# Test single package
go test -v ./pkg/gemini
```

### MongoDB Queries
```bash
# Connect to MongoDB
mongosh nfl_platform

# View collections
show collections

# Query data
db.players.find({ team: "KC" })
db.plays.countDocuments({ season: 2024 })

# Check indexes
db.players.getIndexes()
```

### Common Issues

**Issue**: Gemini API returns errors
```go
// Solution: Check rate limits and add retry logic
func (c *Client) GenerateWithRetry(ctx context.Context, prompt string, retries int) (string, error) {
	var lastErr error
	for i := 0; i < retries; i++ {
		result, err := c.Generate(ctx, prompt)
		if err == nil {
			return result, nil
		}
		lastErr = err
		log.Printf("Gemini error (attempt %d/%d): %v", i+1, retries, err)
		time.Sleep(time.Second * time.Duration(i+1))
	}
	return "", fmt.Errorf("failed after %d retries: %w", retries, lastErr)
}
```

**Issue**: NFLverse data download fails
```go
// Solution: Add timeout and fallback
func fetchWithTimeout(ctx context.Context, url string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("NFLverse download timeout: %s", url)
		return nil, err
	}
	defer resp.Body.Close()
	
	return io.ReadAll(resp.Body)
}
```

**Issue**: MongoDB connection lost
```go
// Solution: Reconnect automatically
func ensureConnection(client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx, nil); err != nil {
		log.Println("MongoDB connection lost, reconnecting...")
		return client.Connect(ctx)
	}
	return nil
}
```

---

## 📈 Performance Tips

### Caching Strategy
```go
// Use Redis or in-memory cache
import "github.com/patrickmn/go-cache"

// Initialize cache
c := cache.New(5*time.Minute, 10*time.Minute)

// Cache expensive computations
func getPlayerStatsWithCache(playerID string) (*PlayerStats, error) {
	cacheKey := fmt.Sprintf("player_stats:%s", playerID)
	
	if cached, found := c.Get(cacheKey); found {
		return cached.(*PlayerStats), nil
	}
	
	stats := expensiveCalculation(playerID)
	c.Set(cacheKey, stats, 1*time.Hour)
	return stats, nil
}

// Cache Gemini responses
func generateWithCache(prompt string) (string, error) {
	hash := md5.Sum([]byte(prompt))
	cacheKey := fmt.Sprintf("gemini:%x", hash)
	
	if cached, found := c.Get(cacheKey); found {
		return cached.(string), nil
	}
	
	response, err := geminiClient.Generate(context.Background(), prompt)
	if err != nil {
		return "", err
	}
	
	c.Set(cacheKey, response, 24*time.Hour)
	return response, nil
}
```

### Database Optimization
```go
// Good: Use projection to limit fields
opts := options.Find().SetProjection(bson.M{
	"name":          1,
	"team":          1,
	"epa_per_play":  1,
})
cursor, err := collection.Find(ctx, filter, opts)

// Good: Use indexes (created in mongodb package)
// Ensure indexes exist for frequently queried fields

// Good: Batch operations
filter := bson.M{"team": "KC"}
update := bson.M{"$set": bson.M{"division": "AFC West"}}
result, err := collection.UpdateMany(ctx, filter, update)

// Use connection pooling (already configured in mongodb.Connect)
```

### Concurrent Processing
```go
// Process data concurrently with goroutines
func processPlayersConc urently(players []Player) []Result {
	results := make([]Result, len(players))
	var wg sync.WaitGroup
	
	for i, player := range players {
		wg.Add(1)
		go func(idx int, p Player) {
			defer wg.Done()
			results[idx] = processPlayer(p)
		}(i, player)
	}
	
	wg.Wait()
	return results
}

// Limit concurrency with worker pool
func processWithWorkerPool(tasks []Task, numWorkers int) {
	tasksChan := make(chan Task, len(tasks))
	var wg sync.WaitGroup
	
	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasksChan {
				processTask(task)
			}
		}()
	}
	
	// Send tasks
	for _, task := range tasks {
		tasksChan <- task
	}
	close(tasksChan)
	
	wg.Wait()
}
```

---

## 🎯 Hackathon Checklist

### Before Demo (Priority)
- [ ] Core features work end-to-end
- [ ] Game Script Predictor produces results
- [ ] Chatbot responds to questions
- [ ] Frontend looks presentable
- [ ] Demo data loaded
- [ ] No critical bugs in demo flow

### Demo Flow
1. Show homepage with live scores
2. Ask chatbot a question → get AI response
3. Select a game → show Game Script Prediction
4. Show player with hot streak + AI explanation
5. Show trade analyzer in action

### Pitch Points
- "First platform to predict HOW games unfold, not just scores"
- "Uses NFLverse's advanced EPA metrics"
- "AI trained on 3+ years of play-by-play data"
- "Quantified predictions: '30% more touches', not vague advice"

---

## 🤝 Team Coordination

### Role Assignments
- **Backend Lead**: Go API + NFLverse integration
- **AI Lead**: Gemini integration + prompt engineering
- **Frontend Lead**: React/Next.js UI
- **DevOps**: Deployment + environment setup

### Communication
- Use Slack/Discord for real-time coordination
- Prefix messages: `[BACKEND]`, `[FRONTEND]`, `[BLOCKER]`
- Share API endpoints as soon as they're ready
- Test integrations early and often

### Code Reviews (Minimal)
- Quick PR reviews (< 5 min)
- Focus on: Does it work? Is it breaking anything?
- Defer style/optimization discussions to post-hackathon

---

## 📚 Resources

### NFLverse Documentation
- Main repo: https://github.com/nflverse
- Data dictionary: https://www.nflverse.com/articles/dictionary.html
- Play-by-play guide: https://www.nflverse.com/articles/nflfastR.html

### Gemini API
- Docs: https://ai.google.dev/docs
- Pricing: https://ai.google.dev/pricing
- Go client examples: https://github.com/google/generative-ai-go

### Go + MongoDB
- MongoDB Go driver: https://www.mongodb.com/docs/drivers/go/current/
- Gin framework: https://gin-gonic.com/docs/
- Go best practices: https://golang.org/doc/effective_go

---

## 🏁 Launch Checklist

### Pre-Launch
- [ ] Environment variables set in production
- [ ] MongoDB indexes created
- [ ] Redis configured
- [ ] Background jobs running
- [ ] CORS configured for frontend domain
- [ ] Rate limiting on expensive endpoints

### Deployment
```bash
# Backend (Railway/Render)
git push origin main  # Auto-deploy

# Frontend (Vercel)
npm run build
vercel --prod

# Database
# Use MongoDB Atlas (free tier)
```

### Post-Launch
- [ ] Monitor error logs
- [ ] Check Gemini API usage
- [ ] Verify background jobs running
- [ ] Test critical user flows

---

**Last Updated**: November 8, 2025  
**Version**: 1.0 - Hackathon Edition

Good luck team! 🚀🏈


