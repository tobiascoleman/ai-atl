# Testing Chatbot with Database Integration

## Overview

The chatbot has been updated to retrieve relevant stats from MongoDB before generating responses. This provides the AI model with actual database context to give more accurate, data-driven answers.

## How It Works

### 1. Query Intent Extraction

When a user asks a question, the system first uses AI to analyze the question and extract:

- **Player names** mentioned (e.g., "Patrick Mahomes", "Travis Kelce")
- **Team abbreviations** (e.g., "KC", "BUF")
- **Positions** (QB, RB, WR, TE, etc.)
- **Stat types** requested (passing, rushing, receiving, EPA, injuries)
- **Season** (current year if not specified)
- **Whether database lookup is needed**

### 2. Database Retrieval

Based on the extracted intent, the system queries MongoDB for:

- Player roster information and injury status
- Season statistics (passing yards, rushing yards, receiving stats, etc.)
- EPA (Expected Points Added) calculations from play-by-play data
- Team-level statistics
- Position-specific player lists

### 3. Context Enhancement

The retrieved stats are formatted and added to the AI prompt, providing concrete data like:

```
=== RELEVANT DATABASE STATS ===

## Patrick Mahomes (QB - KC)
- **2024 REG Stats**:
  - Passing: 3928 yards, 26 TDs, 11 INTs
- **EPA**: 0.234 (over 856 plays)

## Travis Kelce (TE - KC)
- **2024 REG Stats**:
  - Receiving: 97 rec, 823 yards, 3 TDs, 121 targets
- **EPA**: 0.156 (over 342 plays)
```

### 4. AI Response Generation

The AI model then generates a response using both:

- The original user question
- The user's lineup context (if available)
- **The actual stats from the database**

## Example Questions That Benefit

### Player-Specific Questions

- "How is Patrick Mahomes performing this season?"
- "Should I start Travis Kelce or George Kittle?"
- "What are Lamar Jackson's stats?"

### Team Questions

- "How is the Chiefs offense doing?"
- "Are there any injured players on the Bills?"
- "What's the team EPA for Buffalo?"

### Position-Based Questions

- "Who are the top running backs this season?"
- "Show me QB stats"
- "Which receivers have the best EPA?"

### Injury Questions

- "Is Christian McCaffrey injured?"
- "Who's on the injury report for Kansas City?"

## Testing the Updated Chatbot

### Prerequisites

1. MongoDB connection is established
2. NFL data is loaded in the database
3. Backend server is running: `make run`
4. Frontend is running: `cd frontend && npm run dev`

### Test Steps

1. **Login to the application**

   - Navigate to http://localhost:3000
   - Create an account or login

2. **Access the chatbot**

   - Go to the chatbot/assistant page

3. **Ask specific questions about players**

   ```
   "How many yards did Patrick Mahomes throw this season?"
   "Compare Travis Kelce and George Kittle's receiving stats"
   "Is Christian McCaffrey injured?"
   ```

4. **Verify the response includes actual database stats**
   - The AI should reference specific numbers from the database
   - Stats should match what's in MongoDB
   - Responses should be more accurate and data-driven

### API Testing with cURL

You can also test directly via the API:

```bash
# First, get a JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.token')

# Ask the chatbot a question
curl -X POST http://localhost:8080/api/v1/chatbot/ask \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "How is Patrick Mahomes performing this season?"
  }'
```

## Implementation Details

### Files Modified

1. **`internal/services/chatbot.go`**
   - Added `QueryIntent` struct to represent extracted query parameters
   - Added `extractQueryIntent()` to analyze questions using AI
   - Added `retrieveRelevantStats()` to fetch data from MongoDB
   - Added `findPlayersByName()` for player name matching
   - Updated `buildChatbotPrompt()` to include database stats
   - Modified `Ask()` to orchestrate the entire flow

### Key Features

1. **Smart Query Extraction**: Uses AI to understand what data is being requested
2. **Flexible Name Matching**: Case-insensitive regex matching for player names
3. **Multi-Source Data**: Combines player stats, EPA, injuries, and team data
4. **Graceful Degradation**: If data retrieval fails, still provides a response
5. **Context-Aware**: Includes user's lineup information when available

### Data Sources Used

- **players collection**: Roster data, injury status, team, position
- **player_stats collection**: Season statistics (passing, rushing, receiving)
- **plays collection**: Play-by-play data for EPA calculations
- **Aggregation queries**: Team-level EPA and stats

## Benefits

1. **Accuracy**: Responses based on actual database stats, not hallucinated data
2. **Current Data**: Always uses the most recent data loaded in MongoDB
3. **Comprehensive**: Can pull from multiple collections and aggregate data
4. **Flexible**: Handles various question types and combinations
5. **Scalable**: As more data is added to MongoDB, chatbot automatically has access

## Future Enhancements

Potential improvements to consider:

- Add Next Gen Stats (NGS) data when available
- Include matchup history and opponent analysis
- Add weekly trends and projections
- Cache frequently requested player data
- Add confidence scores for recommendations
- Include weather data for outdoor games
- Add scoring system configuration for fantasy points
