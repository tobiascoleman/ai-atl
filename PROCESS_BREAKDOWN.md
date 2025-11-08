# Detailed Breakdown: How the Database-Integrated Chatbot Works

## Table of Contents

1. [High-Level Overview](#high-level-overview)
2. [Step-by-Step Process](#step-by-step-process)
3. [Component Breakdown](#component-breakdown)
4. [Data Structures](#data-structures)
5. [Example Walkthrough](#example-walkthrough)
6. [Technical Details](#technical-details)

---

## High-Level Overview

The chatbot now operates in a **two-stage AI process**:

```
Stage 1: Understanding (AI analyzes the question)
    ↓
Stage 2: Retrieval (Query MongoDB for relevant data)
    ↓
Stage 3: Response (AI generates answer using real data)
```

**Before:** AI answered based solely on training data (potentially outdated/inaccurate)
**After:** AI first determines what data is needed, retrieves it from MongoDB, then answers with actual stats

---

## Step-by-Step Process

### Step 1: User Asks a Question

```go
// User sends a question through the API
question := "How is Patrick Mahomes performing this season?"
```

**What happens:**

- Frontend sends POST request to `/api/v1/chatbot/ask`
- Handler validates JWT token and extracts user ID
- Calls `ChatbotService.Ask(ctx, userID, question)`

---

### Step 2: Get User Context (Existing Feature)

```go
// Get user's lineup context
objID, _ := bson.ObjectIDFromHex(userID)

var lineups []models.FantasyLineup
cursor, err := s.db.Collection("lineups").Find(ctx, bson.M{"user_id": objID})
if err == nil {
    cursor.All(ctx, &lineups)
}
```

**What happens:**

- Queries MongoDB for user's fantasy lineups
- This provides context about which players the user has
- Used to personalize responses

**Example Output:**

```
User has lineup with: QB: Josh Allen, RB1: Derrick Henry, RB2: Saquon Barkley, ...
```

---

### Step 3: Extract Query Intent (NEW!)

```go
// Extract query intent from the question
intent, err := s.extractQueryIntent(ctx, question)
if err != nil {
    // If extraction fails, continue without data context
    intent = &QueryIntent{NeedsData: false}
}
```

**What happens:**
This is where **AI analyzes the question** to understand what database queries are needed.

#### How `extractQueryIntent()` Works:

1. **Creates a specialized prompt for AI:**

```go
extractionPrompt := fmt.Sprintf(`Analyze this fantasy football question and extract the data requirements.
Return ONLY a valid JSON object with this structure:
{
  "player_names": ["name1", "name2"],
  "teams": ["team1"],
  "positions": ["QB", "RB"],
  "stat_types": ["passing", "rushing", "receiving", "epa"],
  "season": 2024,
  "needs_data": true
}

Rules:
- player_names: Full player names mentioned
- teams: Team abbreviations (e.g., ["KC", "BUF"])
- positions: Positions mentioned (QB, RB, WR, TE, K, DEF)
- stat_types: Types of stats (passing, rushing, receiving, epa, injuries)
- season: Year mentioned or 2024 if current season
- needs_data: true if specific players/teams/stats mentioned

Question: %s

Return only the JSON object, no explanation.`, question)
```

2. **Sends to Gemini AI:**

```go
response, err := s.gemini.Generate(ctx, extractionPrompt)
```

3. **AI analyzes and returns JSON:**
   For "How is Patrick Mahomes performing this season?", AI returns:

```json
{
  "player_names": ["Patrick Mahomes"],
  "teams": ["KC"],
  "positions": ["QB"],
  "stat_types": ["passing", "epa"],
  "season": 2024,
  "needs_data": true
}
```

4. **Parses the JSON response:**

```go
// Clean up response (remove markdown code blocks, etc.)
response = strings.TrimSpace(response)
response = strings.Trim(response, "`")
if strings.HasPrefix(response, "json") {
    response = strings.TrimPrefix(response, "json")
    response = strings.TrimSpace(response)
}

// Parse into QueryIntent struct
var intent QueryIntent
if err := json.Unmarshal([]byte(response), &intent); err != nil {
    return nil, fmt.Errorf("failed to parse intent JSON: %w", err)
}
```

**Why this is clever:**

- Uses AI's natural language understanding to extract structured data
- No need to write complex regex or NLP parsing logic
- Handles variations in how users ask questions
- Robust to typos and different phrasings

---

### Step 4: Retrieve Relevant Stats (NEW!)

```go
// Retrieve relevant stats from database if needed
var statsContext string
if intent.NeedsData {
    statsContext, err = s.retrieveRelevantStats(ctx, intent)
    if err != nil {
        statsContext = "Unable to retrieve some requested stats from database."
    }
}
```

**What happens:**
If `NeedsData` is true, we query MongoDB for the specific information identified in the intent.

#### How `retrieveRelevantStats()` Works:

This is a **comprehensive data aggregator** that builds a formatted stats report.

**1. Initialize the stats builder:**

```go
var statsBuilder strings.Builder
statsBuilder.WriteString("\n=== RELEVANT DATABASE STATS ===\n\n")

currentSeason := time.Now().Year()
if intent.Season == 0 {
    intent.Season = currentSeason  // Default to current year
}
```

**2. Process each player mentioned:**

```go
for _, playerName := range intent.PlayerNames {
    // Find player by name (case-insensitive regex search)
    players, err := s.findPlayersByName(ctx, playerName, intent.Season)
    if err != nil || len(players) == 0 {
        continue  // Skip if not found
    }

    player := players[0]  // Use first match
    statsBuilder.WriteString(fmt.Sprintf("## %s (%s - %s)\n",
        player.Name, player.Position, player.Team))
```

**3. Get injury status:**

```go
// Check if player is injured
if player.Status == "INA" || player.StatusDescriptionAbbr != "" {
    statsBuilder.WriteString(fmt.Sprintf("- **Injury Status**: %s (Week %d)\n",
        player.StatusDescriptionAbbr, player.Week))
}
```

**4. Get season statistics:**

```go
// Query player_stats collection
stats, err := s.dataService.GetPlayerStats(ctx, player.NFLID, intent.Season)
if err == nil && len(stats) > 0 {
    for _, stat := range stats {
        statsBuilder.WriteString(fmt.Sprintf("- **%d %s Stats**:\n",
            stat.Season, stat.SeasonType))

        // Add passing stats if applicable
        if stat.PassingYards > 0 {
            statsBuilder.WriteString(fmt.Sprintf("  - Passing: %d yards, %d TDs, %d INTs\n",
                stat.PassingYards, stat.PassingTDs, stat.Interceptions))
        }

        // Add rushing stats if applicable
        if stat.RushingYards > 0 {
            statsBuilder.WriteString(fmt.Sprintf("  - Rushing: %d yards, %d TDs\n",
                stat.RushingYards, stat.RushingTDs))
        }

        // Add receiving stats if applicable
        if stat.Receptions > 0 {
            statsBuilder.WriteString(fmt.Sprintf("  - Receiving: %d rec, %d yards, %d TDs, %d targets\n",
                stat.Receptions, stat.ReceivingYards, stat.ReceivingTDs, stat.Targets))
        }
    }
}
```

**5. Calculate EPA if requested:**

```go
// EPA is calculated from play-by-play data
if s.containsStatType(intent.StatTypes, "epa") {
    epa, playCount, err := s.dataService.CalculatePlayerEPA(ctx, player.NFLID, intent.Season)
    if err == nil && playCount > 0 {
        statsBuilder.WriteString(fmt.Sprintf("- **EPA**: %.3f (over %d plays)\n",
            epa, playCount))
    }
}
```

**6. Process team data if mentioned:**

```go
for _, team := range intent.Teams {
    statsBuilder.WriteString(fmt.Sprintf("## Team: %s\n", team))

    // Get team EPA
    epa, playCount, err := s.dataService.CalculateTeamEPA(ctx, team, intent.Season)
    if err == nil && playCount > 0 {
        statsBuilder.WriteString(fmt.Sprintf("- **Team EPA**: %.3f (over %d plays)\n",
            epa, playCount))
    }

    // Get injured players on team
    if s.containsStatType(intent.StatTypes, "injuries") {
        players, err := s.dataService.GetPlayersByTeam(ctx, team, intent.Season)
        if err == nil {
            var injured []string
            for _, p := range players {
                if p.Status == "INA" || p.StatusDescriptionAbbr != "" {
                    injured = append(injured, fmt.Sprintf("%s (%s)", p.Name, p.Position))
                }
            }
            if len(injured) > 0 {
                statsBuilder.WriteString(fmt.Sprintf("- **Injured Players**: %s\n",
                    strings.Join(injured, ", ")))
            }
        }
    }
}
```

**7. Process position data if mentioned:**

```go
for _, position := range intent.Positions {
    players, err := s.dataService.GetPlayersByPosition(ctx, position, intent.Season)
    if err == nil && len(players) > 0 {
        statsBuilder.WriteString(fmt.Sprintf("## Top %s Players (limited to first 10)\n", position))
        count := 0
        for _, p := range players {
            if count >= 10 {
                break
            }
            statsBuilder.WriteString(fmt.Sprintf("- %s (%s)\n", p.Name, p.Team))
            count++
        }
    }
}
```

**Example Output:**

```
=== RELEVANT DATABASE STATS ===

## Patrick Mahomes (QB - KC)
- **2024 REG Stats**:
  - Passing: 3928 yards, 26 TDs, 11 INTs
- **EPA**: 0.234 (over 856 plays)
```

---

### Step 5: Build Enhanced Prompt

```go
// Build context-aware prompt with database stats
prompt := s.buildChatbotPrompt(question, lineups, statsContext)
```

**What happens:**
Combines everything into a comprehensive prompt for the AI:

```go
func (s *ChatbotService) buildChatbotPrompt(question string, lineups []models.FantasyLineup, statsContext string) string {
    contextInfo := "No lineup information available."
    if len(lineups) > 0 {
        lineup := lineups[len(lineups)-1]
        contextInfo = fmt.Sprintf("User's current lineup: %v", lineup.Positions)
    }

    // Add database stats context if available
    dataContext := ""
    if statsContext != "" {
        dataContext = fmt.Sprintf("\n\nDatabase Stats:\n%s", statsContext)
    }

    return fmt.Sprintf(`You are an expert NFL fantasy football advisor with access to advanced EPA metrics and player data.

User Context:
%s%s

User Question: %s

Provide specific, actionable fantasy football advice based on:
1. The actual stats from our database shown above
2. Recent player performance and trends
3. Matchup analysis
4. Injury reports
5. Advanced metrics (EPA, target share, snap counts)

IMPORTANT: Reference the specific stats provided from our database in your answer. Use the actual numbers shown.

Be conversational but data-driven. Explain your reasoning.`,
        contextInfo,
        dataContext,
        question,
    )
}
```

**Final Prompt Sent to AI:**

```
You are an expert NFL fantasy football advisor with access to advanced EPA metrics and player data.

User Context:
User's current lineup: map[QB:Josh Allen RB1:Derrick Henry RB2:Saquon Barkley ...]

Database Stats:
=== RELEVANT DATABASE STATS ===

## Patrick Mahomes (QB - KC)
- **2024 REG Stats**:
  - Passing: 3928 yards, 26 TDs, 11 INTs
- **EPA**: 0.234 (over 856 plays)

User Question: How is Patrick Mahomes performing this season?

Provide specific, actionable fantasy football advice based on:
1. The actual stats from our database shown above
2. Recent player performance and trends
3. Matchup analysis
4. Injury reports
5. Advanced metrics (EPA, target share, snap counts)

IMPORTANT: Reference the specific stats provided from our database in your answer. Use the actual numbers shown.

Be conversational but data-driven. Explain your reasoning.
```

---

### Step 6: Generate AI Response

```go
// Get AI response
response, err := s.gemini.GenerateWithRetry(ctx, prompt, 3)
if err != nil {
    return "", fmt.Errorf("failed to generate response: %w", err)
}

return response, nil
```

**What happens:**

- Sends the enhanced prompt to Gemini AI
- AI now has access to:
  - Original question
  - User's lineup
  - **Real database stats**
  - Instructions to use those stats
- Retries up to 3 times if there's an error
- Returns the final response

**Example Response:**

```
Patrick Mahomes is having a solid 2024 season. According to our database,
he's thrown for 3,928 yards with 26 touchdowns and 11 interceptions. His EPA
of 0.234 over 856 plays shows he's been quite efficient, averaging positive
expected points on his plays.

For fantasy purposes, Mahomes remains a strong QB1 option. His touchdown rate
is excellent, and the positive EPA indicates he's making good decisions with
the football...
```

---

## Component Breakdown

### Key Components

#### 1. ChatbotService Struct

```go
type ChatbotService struct {
    db          *mongo.Database  // Direct database access
    gemini      *gemini.Client   // AI client for query parsing & responses
    dataService *DataService     // Service for complex database queries
}
```

#### 2. QueryIntent Struct

```go
type QueryIntent struct {
    PlayerNames []string // ["Patrick Mahomes", "Travis Kelce"]
    Teams       []string // ["KC", "BUF"]
    Positions   []string // ["QB", "RB"]
    StatTypes   []string // ["passing", "rushing", "epa"]
    Season      int      // 2024
    NeedsData   bool     // true if database lookup required
}
```

#### 3. Helper Methods

**`findPlayersByName(ctx, name, season)`**

- Searches players collection with case-insensitive regex
- Returns up to 5 matching players
- Handles partial name matches

**`containsStatType(statTypes, target)`**

- Checks if a specific stat type was requested
- Case-insensitive comparison
- Defaults to including all stats

---

## Data Structures

### MongoDB Collections Used

1. **`players`** collection:

```go
type Player struct {
    NFLID    string  // Unique player ID
    Season   int     // Year
    Name     string  // Player name
    Team     string  // Team abbreviation
    Position string  // QB, RB, WR, etc.
    Status   string  // INA (injured) or ACT (active)
    StatusDescriptionAbbr string  // R01, P02, etc.
    Week     int     // Latest week for status
}
```

2. **`player_stats`** collection:

```go
type PlayerStats struct {
    NFLID          string
    Season         int
    SeasonType     string  // REG, POST
    PassingYards   int
    PassingTDs     int
    Interceptions  int
    RushingYards   int
    RushingTDs     int
    Receptions     int
    ReceivingYards int
    ReceivingTDs   int
    Targets        int
}
```

3. **`plays`** collection:

- Used for EPA calculations
- Contains play-by-play data with EPA values
- Aggregated to calculate averages

---

## Example Walkthrough

### Complete Example: "Should I start Travis Kelce or George Kittle?"

**Step 1: User Question**

```
"Should I start Travis Kelce or George Kittle?"
```

**Step 2: Extract Intent**
AI analyzes and returns:

```json
{
  "player_names": ["Travis Kelce", "George Kittle"],
  "teams": ["KC", "SF"],
  "positions": ["TE"],
  "stat_types": ["receiving", "epa"],
  "season": 2024,
  "needs_data": true
}
```

**Step 3: Retrieve Stats**

Query 1: Find "Travis Kelce" in players (2024)

```go
s.findPlayersByName(ctx, "Travis Kelce", 2024)
// Returns: Player{NFLID: "00-0032423", Name: "Travis Kelce", Team: "KC", Position: "TE"}
```

Query 2: Get Travis Kelce stats

```go
s.dataService.GetPlayerStats(ctx, "00-0032423", 2024)
// Returns: PlayerStats{Receptions: 97, ReceivingYards: 823, ReceivingTDs: 3, Targets: 121}
```

Query 3: Calculate Travis Kelce EPA

```go
s.dataService.CalculatePlayerEPA(ctx, "00-0032423", 2024)
// Returns: EPA: 0.156, PlayCount: 342
```

Query 4-6: Repeat for George Kittle

**Built Stats Context:**

```
=== RELEVANT DATABASE STATS ===

## Travis Kelce (TE - KC)
- **2024 REG Stats**:
  - Receiving: 97 rec, 823 yards, 3 TDs, 121 targets
- **EPA**: 0.156 (over 342 plays)

## George Kittle (TE - SF)
- **2024 REG Stats**:
  - Receiving: 65 rec, 1021 yards, 6 TDs, 87 targets
- **EPA**: 0.189 (over 298 plays)
```

**Step 4: Build Prompt**
Combines user context + stats + question

**Step 5: AI Response**

```
Based on the stats from our database, this is an interesting matchup:

Travis Kelce: 97 receptions on 121 targets (80.2% catch rate), 823 yards, 3 TDs, EPA: 0.156
George Kittle: 65 receptions on 87 targets (74.7% catch rate), 1,021 yards, 6 TDs, EPA: 0.189

Kittle has the edge in big-play ability and touchdowns, averaging 15.7 yards per reception
compared to Kelce's 8.5. His higher EPA (0.189 vs 0.156) and touchdown production make him
the slightly better play this week.

However, Kelce's high target volume (121 targets) provides a safer floor in PPR formats.

Recommendation: Kittle for upside, Kelce for safety.
```

---

## Technical Details

### Error Handling

- Graceful degradation if intent extraction fails
- Continues without data context if retrieval fails
- Retries AI generation up to 3 times
- Handles missing players/stats gracefully

### Performance Considerations

- Uses MongoDB indexes for fast queries
- Limits position queries to 10 results
- Uses regex for flexible name matching
- Context timeouts prevent hanging requests

### Scalability

- Can query multiple players/teams in parallel
- Efficient aggregation for EPA calculations
- Reuses DataService for optimized queries
- Stateless service (easily horizontally scalable)

---

## Summary

The chatbot now uses a **two-stage AI approach**:

1. **Stage 1 AI (Understanding)**: Analyzes the question to extract structured query parameters
2. **Database Queries**: Retrieves relevant real-time data from MongoDB
3. **Stage 2 AI (Response)**: Generates answer using actual database stats

This makes responses:

- ✅ **Accurate** - Based on real data, not hallucinations
- ✅ **Current** - Uses up-to-date database information
- ✅ **Specific** - References actual numbers
- ✅ **Comprehensive** - Combines stats, EPA, injuries
- ✅ **Flexible** - Handles various question types automatically
