# Chatbot Database Integration - Summary of Changes

## Overview

The chatbot has been enhanced to retrieve and utilize real data from the MongoDB database when answering user questions. This makes responses more accurate, data-driven, and based on actual current statistics rather than AI training data.

## Changes Made

### 1. Updated `internal/services/chatbot.go`

#### New Structures

- **`QueryIntent` struct**: Represents extracted query parameters from user questions
  - Player names
  - Team abbreviations
  - Positions
  - Stat types (passing, rushing, receiving, EPA, injuries)
  - Season year
  - Whether data lookup is needed

#### New Service Components

- **`dataService`**: Instance of DataService for querying MongoDB
  - Added to ChatbotService struct
  - Initialized in NewChatbotService constructor

#### New Methods

**`extractQueryIntent(ctx, question)`**

- Uses Gemini AI to analyze user questions
- Extracts relevant query parameters
- Returns QueryIntent struct with parsed information
- Example: "How is Patrick Mahomes doing?" → extracts player name, stat types, season

**`retrieveRelevantStats(ctx, intent)`**

- Fetches data from MongoDB based on QueryIntent
- Queries multiple collections:
  - `players` - for roster info and injury status
  - `player_stats` - for season statistics
  - `plays` - for EPA calculations
- Formats stats into readable context string
- Handles player-specific, team-specific, and position-based queries

**`findPlayersByName(ctx, name, season)`**

- Searches for players using regex matching
- Case-insensitive search
- Returns up to 5 matching players
- Handles partial name matches

**`containsStatType(statTypes, target)`**

- Helper to check if specific stat type is requested
- Used to conditionally fetch EPA, injuries, etc.

#### Updated Methods

**`Ask(ctx, userID, question)` - Main Flow Enhanced**

```go
1. Get user lineup context (existing)
2. Extract query intent from question (NEW)
3. Retrieve relevant stats from database (NEW)
4. Build enhanced prompt with stats (UPDATED)
5. Generate AI response (existing)
```

**`buildChatbotPrompt(question, lineups, statsContext)` - Now Accepts Stats**

- Added `statsContext` parameter
- Includes database stats in prompt
- Instructs AI to reference actual numbers
- Maintains existing lineup context

### 2. New Imports Added

```go
"encoding/json"     // For parsing JSON from AI responses
"strings"           // For string manipulation
"time"              // For date handling
"go.mongodb.org/mongo-driver/v2/mongo/options"  // For query options
```

## Data Flow

```
User Question
    ↓
Extract Intent (AI parses question)
    ↓
Query MongoDB (fetch relevant stats)
    ↓
Build Enhanced Prompt (add stats to context)
    ↓
Generate Response (AI uses real data)
    ↓
Return to User
```

## Example Transformations

### Before (No Database Integration)

**User:** "How is Patrick Mahomes doing this season?"

**AI Prompt:**

```
You are an expert NFL fantasy football advisor...
User Question: How is Patrick Mahomes doing this season?
```

**Response:** Generic answer based on training data (possibly outdated)

### After (With Database Integration)

**User:** "How is Patrick Mahomes doing this season?"

**AI Prompt:**

```
You are an expert NFL fantasy football advisor...

Database Stats:
=== RELEVANT DATABASE STATS ===

## Patrick Mahomes (QB - KC)
- **2024 REG Stats**:
  - Passing: 3928 yards, 26 TDs, 11 INTs
- **EPA**: 0.234 (over 856 plays)

User Question: How is Patrick Mahomes doing this season?

IMPORTANT: Reference the specific stats provided from our database in your answer.
```

**Response:** Data-driven answer with specific numbers from the database

## Database Collections Used

1. **players**: Roster data, teams, positions, injury status
2. **player_stats**: Season statistics (passing, rushing, receiving)
3. **plays**: Play-by-play data for EPA calculations

## Types of Queries Supported

### Player Queries

- Individual player stats
- Player comparisons
- Injury status checks
- EPA analysis

### Team Queries

- Team offense/defense stats
- Team EPA calculations
- Team injury reports

### Position Queries

- Top players by position
- Position-based rankings

## Error Handling

- **Graceful degradation**: If data retrieval fails, chatbot still responds
- **Fallback behavior**: Continues without database context if needed
- **Flexible matching**: Handles partial player name matches
- **Safe parsing**: JSON parsing errors don't crash the service

## Testing

### Files Created

1. **`test_chatbot_with_data.md`**: Comprehensive testing guide
2. **`chatbot_flow_diagram.md`**: Visual flow diagram
3. **`test_chatbot.sh`**: Automated test script

### Test the Changes

```bash
# 1. Start the backend
make run

# 2. Run the test script
./test_chatbot.sh

# Or test manually via API
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.token')

curl -X POST http://localhost:8080/api/v1/chatbot/ask \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"question": "How is Patrick Mahomes performing this season?"}'
```

## Benefits

1. **Accuracy**: Uses real database stats, not hallucinated data
2. **Timeliness**: Always uses current data from MongoDB
3. **Specificity**: References actual numbers in responses
4. **Comprehensive**: Combines multiple data sources (stats, EPA, injuries)
5. **Flexible**: Handles various question types automatically
6. **Scalable**: As more data is added, chatbot automatically has access

## Future Enhancements

Potential improvements to consider:

- Add caching for frequently requested players
- Include Next Gen Stats (NGS) when querying
- Add weekly trend analysis
- Include matchup history
- Add weather data integration
- Support scoring system customization
- Add confidence scores for recommendations

## Configuration

No configuration changes needed. The system:

- Uses existing MongoDB connection
- Uses existing Gemini API client
- Uses existing DataService methods
- Works with current authentication system

## Backward Compatibility

- Existing chatbot API endpoints unchanged
- No breaking changes to request/response format
- Gracefully handles missing data
- Falls back to general responses if needed

## Code Quality

- ✅ Compiles successfully
- ✅ No lint errors
- ✅ Follows existing code patterns
- ✅ Includes error handling
- ✅ Well-documented with comments
- ✅ Type-safe with proper structs
