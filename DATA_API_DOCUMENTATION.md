# 📊 Data API Documentation - For Dev 2

## ✅ What's Ready

All NFL data is loaded and queryable through REST API endpoints:
- ✅ **~18,000 player rosters** (with injury status)
- ✅ **~15,000 player stats** (2017-2025)
- ✅ **~950,000 plays** with EPA/WPA
- ✅ **~20,000 NGS stats** (advanced metrics)
- ✅ **Games & schedules**

---

## 🔐 Authentication

All endpoints require JWT token:
```bash
Authorization: Bearer <your_jwt_token>
```

Get token from `/api/v1/auth/login`

---

## 📡 Base URL

```
http://localhost:8080/api/v1/data
```

---

## 🎯 Quick Examples

### Get Player EPA
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/data/players/00-0036945/epa?season=2024"
```

**Response:**
```json
{
  "nfl_id": "00-0036945",
  "season": 2024,
  "epa": 0.23,
  "play_count": 856
}
```

### Get Injured Players
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/data/injuries?season=2025"
```

**Response:**
```json
{
  "season": 2025,
  "count": 47,
  "players": [
    {
      "nfl_id": "00-0035228",
      "name": "Aaron Rodgers",
      "team": "NYJ",
      "position": "QB",
      "status": "INA",
      "status_description_abbr": "R01",
      "week": 4
    }
  ]
}
```

### Get NGS Advanced Stats
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/data/players/00-0036945/ngs?stat_type=receiving&season=2024"
```

**Response:**
```json
{
  "nfl_id": "00-0036945",
  "stat_type": "receiving",
  "season": 2024,
  "count": 2,
  "stats": [
    {
      "avg_separation": 3.2,
      "avg_yac_above_expectation": 1.5,
      "catch_percentage": 68.3,
      "targets": 142
    }
  ]
}
```

---

## 📚 Complete API Reference

### **PLAYER ENDPOINTS**

#### Get Player
```
GET /data/players/:nfl_id?season=2024
```
Returns player roster info for a season.

#### Get Player Stats
```
GET /data/players/:nfl_id/stats?season=2024
```
Returns seasonal statistics (passing/rushing/receiving yards, TDs, etc.)

#### Get Player EPA
```
GET /data/players/:nfl_id/epa?season=2024
```
Calculates average EPA from all plays involving the player.

**Use this for**: Trade analyzer, betting analysis, player rankings

#### Get Player Plays
```
GET /data/players/:nfl_id/plays?season=2024&limit=100
```
Returns individual plays the player was involved in.

**Use this for**: Play-by-play analysis, situational usage

#### Get Player NGS
```
GET /data/players/:nfl_id/ngs?stat_type=receiving&season=2024
```
Returns Next Gen Stats (advanced metrics).

**stat_type options**: `passing`, `rushing`, `receiving`

**Use this for**: 
- QBs: completion % above expectation, time to throw
- RBs: yards over expected, efficiency
- WRs: separation, cushion, YAC above expected

#### Get Player Summary
```
GET /data/players/:nfl_id/summary?season=2024
```
Returns everything: player info, stats, EPA, NGS in one call.

**Use this for**: Player profile pages, comprehensive analysis

---

### **TEAM ENDPOINTS**

#### Get Team Players
```
GET /data/teams/:team/players?season=2025
```
Returns all players on a team's roster.

**Example**: `/data/teams/DAL/players?season=2025`

#### Get Team EPA
```
GET /data/teams/:team/epa?season=2024
```
Calculates team's offensive EPA.

**Use this for**: Betting analysis, matchup evaluation

#### Get Team Plays
```
GET /data/teams/:team/plays?season=2024&limit=100
```
Returns plays for/against a team.

#### Get Team Depth Chart
```
GET /data/teams/:team/depth-chart?season=2025
```
Returns roster organized by position.

**Use this for**: Injury impact analysis, finding backups

**Response:**
```json
{
  "team": "DAL",
  "season": 2025,
  "depth_chart": {
    "QB": [ ... ],
    "RB": [ ... ],
    "WR": [ ... ]
  }
}
```

#### Get Upcoming Games
```
GET /data/teams/:team/upcoming
```
Returns next 5 games for a team.

---

### **POSITION ENDPOINTS**

#### Get Players by Position
```
GET /data/positions/:position?season=2025
```
Returns all players at a position (limit 100).

**Examples**: 
- `/data/positions/QB?season=2025`
- `/data/positions/WR?season=2025`

**Use this for**: Position rankings, waiver wire analysis

---

### **INJURY ENDPOINTS**

#### Get Injured Players
```
GET /data/injuries?season=2025
```
Returns all players with injury status (INA or injury designation).

**Response includes**:
- `status`: "ACT" or "INA"
- `status_description_abbr`: "R01" (IR), "R04" (PUP), etc.
- `week`: Latest week updated

**Use this for**: Injury impact predictions, waiver recommendations

---

### **GAME ENDPOINTS**

#### Get Games by Season
```
GET /data/games?season=2024&week=1
```
Returns games for a season/week.

#### Get Game
```
GET /data/games/:game_id
```
Returns specific game info (scores, Vegas lines, etc.)

#### Get Game Plays
```
GET /data/games/:game_id/plays
```
Returns all plays from a game.

**Use this for**: Game script analysis, situational breakdowns

---

### **NGS LEADER ENDPOINTS**

#### Get NGS Leaders
```
GET /data/ngs/leaders?stat_type=passing&season=2024&metric=completion_percentage_above_expectation&limit=10
```

**Available Metrics**:

**Passing**:
- `completion_percentage_above_expectation`
- `avg_time_to_throw`
- `avg_completed_air_yards`
- `avg_intended_air_yards`

**Rushing**:
- `rush_yards_over_expected`
- `efficiency`
- `avg_time_to_los`

**Receiving**:
- `avg_separation`
- `avg_yac_above_expectation`
- `catch_percentage`

**Use this for**: Rankings, player comparisons, waiver analysis

---

## 🤖 Using in AI Services

### Example: Betting Analyzer

```go
// In internal/services/betting_analyzer.go

func (s *BettingAnalyzer) AnalyzeLine(ctx context.Context, gameID string) (*BettingAnalysis, error) {
    // Get game
    game, _ := s.dataService.GetGame(ctx, gameID)
    
    // Get team EPAs
    homeEPA, _, _ := s.dataService.CalculateTeamEPA(ctx, game.HomeTeam, game.Season)
    awayEPA, _, _ := s.dataService.CalculateTeamEPA(ctx, game.AwayTeam, game.Season)
    
    // Get game plays for trends
    plays, _ := s.dataService.GetGamePlays(ctx, gameID)
    
    // Generate AI analysis with REAL data
    prompt := fmt.Sprintf(`
        Analyze this betting line:
        
        %s (EPA: %.2f) vs %s (EPA: %.2f)
        Vegas Line: %s %+.1f
        Over/Under: %.1f
        
        Based on EPA trends and play data, provide:
        1. Best bet recommendation
        2. Confidence level
        3. Key factors
    `, game.HomeTeam, homeEPA, game.AwayTeam, awayEPA, 
       game.HomeTeam, game.VegasLine, game.OverUnder)
    
    return s.gemini.GenerateWithRetry(ctx, prompt, 3)
}
```

### Example: Trade Analyzer

```go
// In internal/services/trade_analyzer.go

func (s *TradeAnalyzer) CompareP layers(ctx context.Context, player1ID, player2ID string) (*TradeComparison, error) {
    // Get comprehensive data
    p1Summary, _ := s.dataService.GetPlayerSummary(ctx, player1ID, 2024)
    p2Summary, _ := s.dataService.GetPlayerSummary(ctx, player2ID, 2024)
    
    // Extract key metrics
    p1EPA := p1Summary["epa"].(float64)
    p2EPA := p2Summary["epa"].(float64)
    
    p1Stats := p1Summary["stats"].([]models.PlayerStats)
    p2Stats := p2Summary["stats"].([]models.PlayerStats)
    
    p1NGS := p1Summary["ngs"].([]models.NextGenStat)
    p2NGS := p2Summary["ngs"].([]models.NextGenStat)
    
    // Generate AI comparison with REAL metrics
    prompt := fmt.Sprintf(`
        Compare these two players for a trade:
        
        Player 1:
        - EPA: %.2f
        - Stats: %d yards, %d TDs
        - NGS: %.1f separation, %.1f YAC above expected
        
        Player 2:
        - EPA: %.2f
        - Stats: %d yards, %d TDs
        - NGS: %.1f separation, %.1f YAC above expected
        
        Which player has more value? Why?
    `, p1EPA, p1Stats[0].ReceivingYards, p1Stats[0].ReceivingTDs,
       p1NGS[0].AvgSeparation, p1NGS[0].AvgYACAboveExpectation,
       p2EPA, p2Stats[0].ReceivingYards, p2Stats[0].ReceivingTDs,
       p2NGS[0].AvgSeparation, p2NGS[0].AvgYACAboveExpectation)
    
    return s.gemini.GenerateWithRetry(ctx, prompt, 3)
}
```

### Example: Injury Impact

```go
// In internal/services/injury_analyzer.go

func (s *InjuryAnalyzer) AnalyzeInjury(ctx context.Context, playerID string) (*InjuryImpact, error) {
    // Get injured player
    player, _ := s.dataService.GetPlayer(ctx, playerID, 2025)
    
    // Get team depth chart
    depthChart, _ := s.dataService.GetTeamDepthChart(ctx, player.Team, 2025)
    backups := depthChart[player.Position]
    
    // Get EPA for backups
    backupData := []string{}
    for _, backup := range backups {
        if backup.NFLID != playerID {
            epa, playCount, _ := s.dataService.CalculatePlayerEPA(ctx, backup.NFLID, 2024)
            backupData = append(backupData, fmt.Sprintf("%s: %.2f EPA (%d plays)", 
                backup.Name, epa, playCount))
        }
    }
    
    // Generate AI analysis with REAL depth chart
    prompt := fmt.Sprintf(`
        %s (%s) is injured (Status: %s).
        
        Team Depth Chart:
        %s
        
        Predict:
        1. Who benefits most?
        2. Expected snap/target increases
        3. Fantasy recommendations
    `, player.Name, player.Team, player.Status, strings.Join(backupData, "\n"))
    
    return s.gemini.GenerateWithRetry(ctx, prompt, 3)
}
```

---

## 🎯 Key Data Points for AI

### For Betting Analysis:
- ✅ Team EPA (`/data/teams/:team/epa`)
- ✅ Recent plays (`/data/teams/:team/plays`)
- ✅ Game info with Vegas lines (`/data/games/:game_id`)

### For Trade Analysis:
- ✅ Player EPA (`/data/players/:nfl_id/epa`)
- ✅ Player stats (`/data/players/:nfl_id/stats`)
- ✅ NGS metrics (`/data/players/:nfl_id/ngs`)
- ✅ Full summary (`/data/players/:nfl_id/summary`)

### For Injury Analysis:
- ✅ Injured players list (`/data/injuries`)
- ✅ Team depth chart (`/data/teams/:team/depth-chart`)
- ✅ Backup player EPA
- ✅ Real injury status (INA/IR)

### For Waiver Wire:
- ✅ Players by position (`/data/positions/WR`)
- ✅ EPA for each player
- ✅ NGS metrics (separation, efficiency)
- ✅ Recent plays/usage

---

## 📊 Data Service (For Direct Use)

If you need more control, import the service directly:

```go
import "github.com/ai-atl/nfl-platform/internal/services"

dataService := services.NewDataService(db)

// Then call methods directly:
epa, playCount, err := dataService.CalculatePlayerEPA(ctx, playerID, season)
players, err := dataService.GetInjuredPlayers(ctx, season)
ngsStats, err := dataService.GetPlayerNGS(ctx, playerID, "receiving", season)
```

**Available Methods**:
- `GetPlayer()`, `GetPlayersByTeam()`, `GetPlayersByPosition()`
- `GetInjuredPlayers()`
- `GetPlayerStats()`
- `GetPlayerPlays()`, `GetTeamPlays()`, `GetGamePlays()`
- `CalculatePlayerEPA()`, `CalculateTeamEPA()`
- `GetPlayerNGS()`, `GetNGSLeaders()`
- `GetGame()`, `GetGamesBySeason()`, `GetUpcomingGames()`
- `GetPlayerSummary()`, `GetTeamDepthChart()`

---

## 🚀 Testing Endpoints

### Start the API
```bash
cd /Users/tobycoleman/aiatl/ideas/ai-atl
go run cmd/api/main.go
```

### Test an Endpoint
```bash
# Get a JWT token first
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}' \
  | jq -r '.token')

# Test player EPA
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/data/players/00-0036945/epa?season=2024"

# Test injuries
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/data/injuries?season=2025"

# Test NGS leaders
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/data/ngs/leaders?stat_type=receiving&season=2024&metric=avg_separation&limit=10"
```

---

## ✅ What You Can Build

With these endpoints, you can now:

### Betting Analyzer
- Query team EPA trends
- Get play-by-play for game script analysis
- Compare against Vegas lines

### Trade Analyzer
- Get player stats + EPA + NGS in one call
- Calculate true player value
- AI-powered trade recommendations

### Injury Impact
- Query real injured players (INA status)
- Get actual depth chart
- Predict beneficiaries with real data

### Waiver Wire
- Find players by position/EPA
- Get NGS efficiency metrics
- Identify undervalued players

---

## 🎯 Your Tasks (Dev 2)

1. **Update AI Services** (2 hours)
   - Betting analyzer: Use `/data/teams/:team/epa` and `/data/teams/:team/plays`
   - Trade analyzer: Use `/data/players/:nfl_id/summary`
   - Injury analyzer: Use `/data/injuries` and `/data/teams/:team/depth-chart`

2. **Frontend Integration** (1-2 hours)
   - Display injury badges from `/data/injuries`
   - Show NGS metrics on player cards
   - EPA trends in charts

3. **Demo Prep** (30 min)
   - Test AI features with real data
   - Prepare 2-3 wow moments

---

## 💬 Questions?

All endpoints are live and ready to use. The data is real, comprehensive, and ready for your AI features!

**Data loaded**: ✅  
**Endpoints working**: ✅  
**Documentation complete**: ✅  
**Ready for AI integration**: ✅  

🚀 **Go build something awesome!**

