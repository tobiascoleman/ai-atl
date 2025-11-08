# Visual Flow Diagram: Database-Integrated Chatbot

## The Complete Flow

```
┌─────────────────────────────────────────────────────────────┐
│  USER                                                        │
│  "How is Patrick Mahomes performing this season?"           │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ HTTP POST /api/v1/chatbot/ask
                        ▼
┌─────────────────────────────────────────────────────────────┐
│  HANDLER (internal/handlers/chatbot.go)                     │
│  • Validates JWT token                                      │
│  • Extracts user ID                                         │
│  • Calls ChatbotService.Ask()                               │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│  CHATBOT SERVICE - Ask() Method                             │
│  (internal/services/chatbot.go)                             │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
        ┌───────────────────────────────────────┐
        │   STEP 1: Get User Context            │
        │   Query MongoDB: "lineups" collection │
        │   • Find user's fantasy lineups       │
        │   • Get player roster                 │
        └───────────────┬───────────────────────┘
                        │
                        ▼
        ┌───────────────────────────────────────────────────────┐
        │   STEP 2: Extract Query Intent                        │
        │   Method: extractQueryIntent()                        │
        │                                                        │
        │   ┌─────────────────────────────────────────────┐    │
        │   │  Build Extraction Prompt:                   │    │
        │   │  "Analyze this question and extract:        │    │
        │   │   - player_names                            │    │
        │   │   - teams                                   │    │
        │   │   - positions                               │    │
        │   │   - stat_types                              │    │
        │   │   - season                                  │    │
        │   │   - needs_data                              │    │
        │   │  Return as JSON"                            │    │
        │   └────────────────┬────────────────────────────┘    │
        │                    │                                  │
        │                    ▼                                  │
        │   ┌─────────────────────────────────────────────┐    │
        │   │  Send to Gemini AI                          │    │
        │   │  (pkg/gemini/client.go)                     │    │
        │   └────────────────┬────────────────────────────┘    │
        │                    │                                  │
        │                    ▼                                  │
        │   ┌─────────────────────────────────────────────┐    │
        │   │  AI Returns JSON:                           │    │
        │   │  {                                          │    │
        │   │    "player_names": ["Patrick Mahomes"],    │    │
        │   │    "teams": ["KC"],                        │    │
        │   │    "positions": ["QB"],                    │    │
        │   │    "stat_types": ["passing", "epa"],      │    │
        │   │    "season": 2024,                         │    │
        │   │    "needs_data": true                      │    │
        │   │  }                                          │    │
        │   └────────────────┬────────────────────────────┘    │
        │                    │                                  │
        │                    ▼                                  │
        │   ┌─────────────────────────────────────────────┐    │
        │   │  Parse JSON → QueryIntent struct            │    │
        │   └─────────────────────────────────────────────┘    │
        └────────────────────┬──────────────────────────────────┘
                             │
                             ▼
        ┌────────────────────────────────────────────────────────┐
        │   STEP 3: Retrieve Relevant Stats                      │
        │   Method: retrieveRelevantStats()                      │
        │   (Only if intent.NeedsData == true)                   │
        │                                                         │
        │   For each player in intent.PlayerNames:               │
        │   ┌─────────────────────────────────────────────────┐ │
        │   │ 3A. Find Player by Name                         │ │
        │   │ Method: findPlayersByName()                     │ │
        │   │ • Query: players collection                     │ │
        │   │ • Regex search (case-insensitive)               │ │
        │   │ • Returns: Player{NFLID, Name, Team, Position} │ │
        │   └───────────────────┬─────────────────────────────┘ │
        │                       │                                │
        │                       ▼                                │
        │   ┌─────────────────────────────────────────────────┐ │
        │   │ 3B. Get Season Stats                            │ │
        │   │ DataService.GetPlayerStats()                    │ │
        │   │ • Query: player_stats collection                │ │
        │   │ • Filter: {nfl_id: "00-0036945", season: 2024} │ │
        │   │ • Returns: PassingYards, TDs, INTs, etc.       │ │
        │   └───────────────────┬─────────────────────────────┘ │
        │                       │                                │
        │                       ▼                                │
        │   ┌─────────────────────────────────────────────────┐ │
        │   │ 3C. Calculate EPA                               │ │
        │   │ DataService.CalculatePlayerEPA()                │ │
        │   │ • Query: plays collection                       │ │
        │   │ • Aggregate: All plays involving player         │ │
        │   │ • Calculate: Average EPA across all plays       │ │
        │   │ • Returns: EPA value + play count               │ │
        │   └───────────────────┬─────────────────────────────┘ │
        │                       │                                │
        │   For each team in intent.Teams:                       │
        │   ┌─────────────────────────────────────────────────┐ │
        │   │ 3D. Get Team EPA                                │ │
        │   │ DataService.CalculateTeamEPA()                  │ │
        │   │ • Query: plays collection                       │ │
        │   │ • Filter: {possession_team: "KC"}               │ │
        │   └───────────────────┬─────────────────────────────┘ │
        │                       │                                │
        │                       ▼                                │
        │   ┌─────────────────────────────────────────────────┐ │
        │   │ 3E. Get Injured Players                         │ │
        │   │ DataService.GetPlayersByTeam()                  │ │
        │   │ • Query: players collection                     │ │
        │   │ • Filter: Players with injury status            │ │
        │   └─────────────────────────────────────────────────┘ │
        │                                                         │
        │   Build formatted stats string:                        │
        │   ┌─────────────────────────────────────────────────┐ │
        │   │ === RELEVANT DATABASE STATS ===                 │ │
        │   │                                                  │ │
        │   │ ## Patrick Mahomes (QB - KC)                    │ │
        │   │ - **2024 REG Stats**:                           │ │
        │   │   - Passing: 3928 yards, 26 TDs, 11 INTs       │ │
        │   │ - **EPA**: 0.234 (over 856 plays)               │ │
        │   └─────────────────────────────────────────────────┘ │
        └────────────────────┬────────────────────────────────────┘
                             │
                             ▼
        ┌─────────────────────────────────────────────────────┐
        │   STEP 4: Build Enhanced Prompt                     │
        │   Method: buildChatbotPrompt()                      │
        │                                                      │
        │   Combines:                                         │
        │   ├─ System instructions                            │
        │   ├─ User lineup context                            │
        │   ├─ Database stats (from Step 3)                   │
        │   └─ Original question                              │
        │                                                      │
        │   Result: Comprehensive prompt for AI               │
        └────────────────────┬────────────────────────────────┘
                             │
                             ▼
        ┌─────────────────────────────────────────────────────┐
        │   STEP 5: Generate AI Response                      │
        │   Method: gemini.GenerateWithRetry()                │
        │                                                      │
        │   ┌──────────────────────────────────────────────┐  │
        │   │ Send complete prompt to Gemini AI            │  │
        │   │ • Include: Question + Context + Stats        │  │
        │   │ • Retry up to 3 times if error               │  │
        │   └───────────────────┬──────────────────────────┘  │
        │                       │                              │
        │                       ▼                              │
        │   ┌──────────────────────────────────────────────┐  │
        │   │ AI generates response using:                 │  │
        │   │ • Actual database stats                      │  │
        │   │ • User lineup context                        │  │
        │   │ • Fantasy football expertise                 │  │
        │   │                                              │  │
        │   │ Returns: Data-driven answer with            │  │
        │   │          specific numbers                    │  │
        │   └──────────────────────────────────────────────┘  │
        └────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│  RESPONSE SENT TO USER                                       │
│                                                              │
│  "Patrick Mahomes is having a solid 2024 season.            │
│  According to our database, he's thrown for 3,928 yards     │
│  with 26 touchdowns and 11 interceptions. His EPA of 0.234  │
│  over 856 plays shows he's been quite efficient..."         │
└─────────────────────────────────────────────────────────────┘
```

---

## MongoDB Collections and Their Role

```
┌──────────────────────────────────────────────────────────────┐
│  MongoDB Collections Used                                     │
└──────────────────────────────────────────────────────────────┘

┌─────────────────────┐
│  lineups            │  → User's fantasy lineups
│  • user_id          │     (for personalization)
│  • positions        │
│  • players          │
└─────────────────────┘

┌─────────────────────┐
│  players            │  → Roster data
│  • nfl_id           │     • Player info
│  • name             │     • Team
│  • team             │     • Position
│  • position         │     • Injury status
│  • status           │
│  • season           │
└─────────────────────┘

┌─────────────────────┐
│  player_stats       │  → Season statistics
│  • nfl_id           │     • Passing yards/TDs
│  • season           │     • Rushing yards/TDs
│  • season_type      │     • Receiving stats
│  • passing_yards    │     • Targets
│  • passing_tds      │
│  • rushing_yards    │
│  • receptions       │
└─────────────────────┘

┌─────────────────────┐
│  plays              │  → Play-by-play data
│  • game_id          │     • For EPA calculation
│  • play_id          │     • Individual play EPA
│  • passer_player_id │     • Aggregated for averages
│  • rusher_player_id │
│  • receiver_player_id│
│  • epa              │
│  • season           │
└─────────────────────┘
```

---

## Data Flow Diagram

```
┌────────┐
│  USER  │
└───┬────┘
    │
    │ Question: "How is Patrick Mahomes doing?"
    ▼
┌────────────────────────────────────────────────────────┐
│                     AI STAGE 1                          │
│                 (Understanding)                         │
│                                                         │
│  Input: User question                                  │
│  Output: Structured query parameters                   │
│                                                         │
│  "Patrick Mahomes" → player_names: ["Patrick Mahomes"] │
│  "performing"      → stat_types: ["passing", "epa"]    │
│  "this season"     → season: 2024                      │
└───────────────────────────┬────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────┐
│                     DATABASE QUERIES                    │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐  │
│  │   players    │  │ player_stats │  │    plays    │  │
│  │              │  │              │  │             │  │
│  │ Find player  │→ │ Get stats    │→ │ Calc EPA    │  │
│  │ by name      │  │ for season   │  │ from plays  │  │
│  └──────────────┘  └──────────────┘  └─────────────┘  │
│                                                         │
│  Returns: Formatted stats with real numbers            │
└───────────────────────────┬────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────┐
│                     AI STAGE 2                          │
│                   (Response Generation)                 │
│                                                         │
│  Inputs:                                               │
│  • Original question                                   │
│  • User lineup context                                 │
│  • Database stats (actual numbers!)                    │
│                                                         │
│  Output: Data-driven response with specific stats      │
└───────────────────────────┬────────────────────────────┘
                            │
                            ▼
                       ┌────────┐
                       │  USER  │
                       └────────┘
                  Receives accurate answer
```

---

## Comparison: Before vs After

### BEFORE (No Database Integration)

```
User Question
     ↓
AI Response
(Based on training data only)
```

**Problems:**

- ❌ Potentially outdated stats
- ❌ Could hallucinate numbers
- ❌ No real-time injury status
- ❌ Generic responses

---

### AFTER (With Database Integration)

```
User Question
     ↓
AI Analyzes Question
     ↓
Query MongoDB for Relevant Data
     ↓
AI Generates Response with Real Stats
```

**Benefits:**

- ✅ Current, accurate stats
- ✅ Real numbers from database
- ✅ Real-time injury status
- ✅ Data-driven responses
- ✅ References specific metrics

---

## Key Innovation: Two-Stage AI

### Stage 1: Understanding (Query Extraction)

```
AI acts as a "Query Parser"
Input: Natural language question
Output: Structured database query parameters
```

### Stage 2: Response (Answer Generation)

```
AI acts as a "Fantasy Football Advisor"
Input: Question + Real database stats
Output: Data-driven analysis and recommendations
```

This separation allows the system to be:

- **Flexible**: Handles various question types
- **Accurate**: Uses real data, not assumptions
- **Scalable**: Can add new data sources easily
- **Robust**: Graceful degradation if queries fail
