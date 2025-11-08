## Chatbot Data Integration Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                        User Asks Question                            │
│              "How is Patrick Mahomes performing?"                    │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Chatbot Service (Ask Method)                      │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ▼
         ┌───────────────────────────────────────┐
         │   STEP 1: Extract Query Intent        │
         │   - Uses AI to parse question         │
         │   - Extracts: player names, teams,    │
         │     positions, stat types, season     │
         └───────────────┬───────────────────────┘
                         │
                         ▼
              ┌─────────────────────┐
              │  QueryIntent Object │
              │  {                  │
              │    PlayerNames: ["Patrick Mahomes"]
              │    Teams: ["KC"]    │
              │    StatTypes: ["passing", "epa"]
              │    Season: 2024     │
              │    NeedsData: true  │
              │  }                  │
              └─────────┬───────────┘
                        │
                        ▼
         ┌───────────────────────────────────────┐
         │   STEP 2: Retrieve Database Stats     │
         │                                        │
         │   For each player:                    │
         │   ├─ Find player by name              │
         │   ├─ Get injury status                │
         │   ├─ Get season stats                 │
         │   └─ Calculate EPA                    │
         │                                        │
         │   For each team:                      │
         │   ├─ Calculate team EPA               │
         │   └─ Get injured players              │
         │                                        │
         │   For each position:                  │
         │   └─ List top players                 │
         └───────────────┬───────────────────────┘
                         │
                         ▼
         ┌───────────────────────────────────────┐
         │   MongoDB Collections Queried:        │
         │   ├─ players (roster + injuries)      │
         │   ├─ player_stats (season stats)      │
         │   └─ plays (for EPA calculations)     │
         └───────────────┬───────────────────────┘
                         │
                         ▼
              ┌─────────────────────┐
              │  Stats Context      │
              │  "=== DATABASE ===" │
              │  Patrick Mahomes:   │
              │  - 3928 pass yds    │
              │  - 26 TDs, 11 INTs  │
              │  - EPA: 0.234       │
              └─────────┬───────────┘
                        │
                        ▼
         ┌───────────────────────────────────────┐
         │   STEP 3: Build Enhanced Prompt       │
         │                                        │
         │   Combines:                           │
         │   ├─ Original question                │
         │   ├─ User lineup context              │
         │   └─ Database stats                   │
         │                                        │
         │   Instructs AI to reference actual    │
         │   stats from database                 │
         └───────────────┬───────────────────────┘
                         │
                         ▼
         ┌───────────────────────────────────────┐
         │   STEP 4: Send to Gemini AI           │
         │   - Enhanced prompt with real data    │
         │   - AI generates data-driven response │
         │   - References specific stats         │
         └───────────────┬───────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        AI Response                                   │
│                                                                      │
│   "Patrick Mahomes is having a solid 2024 season. According to      │
│   our database, he's thrown for 3,928 yards with 26 touchdowns      │
│   and 11 interceptions. His EPA of 0.234 over 856 plays shows       │
│   he's been efficient, averaging positive expected points on        │
│   his plays. He remains a top-tier QB1 for fantasy..."             │
└─────────────────────────────────────────────────────────────────────┘

```

## Key Improvements

### Before

- AI relied on training data (potentially outdated)
- No access to actual current stats
- Could hallucinate statistics
- Generic responses without specific numbers

### After

- AI has real-time database access
- References actual current season stats
- Accurate numbers from MongoDB
- Data-driven, specific responses
- Includes EPA and advanced metrics
- Aware of injury status

## Example Queries Handled

1. **Player Performance**

   - "How is Patrick Mahomes doing?" → Gets stats + EPA

2. **Player Comparison**

   - "Travis Kelce vs George Kittle?" → Gets both players' stats

3. **Injury Check**

   - "Is CMC injured?" → Checks injury status in DB

4. **Team Analysis**

   - "How's the Chiefs offense?" → Gets team EPA + players

5. **Position Rankings**

   - "Top RBs this season?" → Lists players by position

6. **Start/Sit Decisions**
   - "Should I start Josh Allen?" → Gets stats + lineup context
