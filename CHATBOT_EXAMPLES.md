# Example Chatbot Interactions with Database Integration

## Example 1: Player Performance Query

### User Input

```
"How is Patrick Mahomes performing this season?"
```

### Behind the Scenes

**Step 1: Query Intent Extraction**

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

**Step 2: Database Queries**

- Find player by name: "Patrick Mahomes" in 2024 season
- Get player stats from `player_stats` collection
- Calculate EPA from `plays` collection

**Step 3: Retrieved Data**

```
=== RELEVANT DATABASE STATS ===

## Patrick Mahomes (QB - KC)
- **2024 REG Stats**:
  - Passing: 3928 yards, 26 TDs, 11 INTs
- **EPA**: 0.234 (over 856 plays)
```

**Step 4: AI Response**

```
Patrick Mahomes is having a solid 2024 season. According to our database,
he's thrown for 3,928 yards with 26 touchdowns and 11 interceptions through
the current week. His EPA of 0.234 over 856 plays shows he's been quite
efficient, averaging positive expected points on his plays.

For fantasy purposes, Mahomes remains a strong QB1 option. His touchdown
rate is excellent, and the positive EPA indicates he's making good decisions
with the football. The 11 interceptions are something to monitor, but they
haven't significantly impacted his fantasy value given his high yardage and
touchdown production.
```

---

## Example 2: Player Comparison

### User Input

```
"Should I start Travis Kelce or George Kittle?"
```

### Behind the Scenes

**Query Intent:**

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

**Retrieved Data:**

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

**AI Response:**

```
Based on the stats from our database, this is an interesting matchup between
two elite tight ends with different profiles.

Travis Kelce (KC):
- 97 receptions on 121 targets (80.2% catch rate)
- 823 yards, 3 TDs
- EPA: 0.156

George Kittle (SF):
- 65 receptions on 87 targets (74.7% catch rate)
- 1,021 yards, 6 TDs
- EPA: 0.189

Kittle has the edge in big-play ability and touchdowns, averaging 15.7 yards
per reception compared to Kelce's 8.5. His higher EPA (0.189 vs 0.156) and
touchdown production make him the slightly better play this week.

However, Kelce's high target volume (121 targets) provides a safer floor in
PPR formats. If you need consistency, go with Kelce. If you need ceiling
and explosive upside, Kittle is your guy.

Recommendation: Kittle for upside, Kelce for safety.
```

---

## Example 3: Injury Check

### User Input

```
"Is Christian McCaffrey injured?"
```

**Query Intent:**

```json
{
  "player_names": ["Christian McCaffrey"],
  "teams": ["SF"],
  "positions": ["RB"],
  "stat_types": ["injuries"],
  "season": 2024,
  "needs_data": true
}
```

**Retrieved Data:**

```
=== RELEVANT DATABASE STATS ===

## Christian McCaffrey (RB - SF)
- **Injury Status**: R01 (Week 3)
- **2024 REG Stats**:
  - Rushing: 202 yards, 1 TD
  - Receiving: 14 rec, 146 yards, 0 TDs, 19 targets
```

**AI Response:**

```
Yes, according to our database, Christian McCaffrey is currently on the
Reserve/Injured list (status: R01) as of Week 3. This means he's out
indefinitely.

His 2024 stats show limited action before the injury:
- Rushing: 202 yards, 1 TD
- Receiving: 14 receptions for 146 yards on 19 targets

If you have CMC on your roster, you'll need to find a replacement. Jordan
Mason has been handling the majority of touches in San Francisco's backfield
during McCaffrey's absence and would be a strong waiver pickup if available.
```

---

## Example 4: Team Analysis

### User Input

```
"How is the Kansas City Chiefs offense doing?"
```

**Query Intent:**

```json
{
  "player_names": [],
  "teams": ["KC"],
  "positions": [],
  "stat_types": ["epa", "injuries"],
  "season": 2024,
  "needs_data": true
}
```

**Retrieved Data:**

```
=== RELEVANT DATABASE STATS ===

## Team: KC
- **Team EPA**: 0.186 (over 1247 plays)
- **Injured Players**: Rashee Rice (WR), Hollywood Brown (WR)
```

**AI Response:**

```
The Kansas City Chiefs offense is performing well in 2024, with a team EPA
of 0.186 over 1,247 plays. This positive EPA indicates they're consistently
gaining more expected points than average, which is excellent.

However, there are some injury concerns to monitor:
- Rashee Rice (WR) - Currently injured
- Hollywood Brown (WR) - Currently injured

With two key receivers on the injury report, Patrick Mahomes has had to rely
more heavily on Travis Kelce and the remaining healthy receivers. Despite
these injuries, the Chiefs offense is still producing at a high level, which
speaks to their depth and coaching.

For fantasy purposes, this elevates Kelce's value even more, as he's one of
the few reliable healthy weapons. Any backup receivers getting snaps could
be worth monitoring on waivers.
```

---

## Example 5: General Question (No Database Needed)

### User Input

```
"What should I look for in a waiver wire pickup?"
```

**Query Intent:**

```json
{
  "player_names": [],
  "teams": [],
  "positions": [],
  "stat_types": [],
  "season": 2024,
  "needs_data": false
}
```

**Retrieved Data:**

```
(No database query performed - general question)
```

**AI Response:**

```
When evaluating waiver wire pickups, here are the key factors to consider:

1. **Opportunity**: Look for players who will get significant playing time
   due to injuries, trades, or coaching decisions. Volume is king in fantasy.

2. **Role in Offense**: Target players with defined roles - RB1, WR2,
   starting TE - rather than committee backs or WR3s.

3. **Upcoming Schedule**: Check the next 3-4 weeks of matchups. Favorable
   matchups can make a mediocre player fantasy-relevant.

4. **Target Share/Touch Count**: For RBs, look for 15+ touches. For WRs,
   target share above 20% is ideal.

5. **Red Zone Usage**: Players who get goal-line carries or red zone targets
   have higher touchdown upside.

6. **Recent Trends**: Look at the last 2-3 games, not just one breakout week.

7. **Your Team Needs**: Don't just grab the "best" player available - target
   positions where you need depth or are dealing with byes/injuries.

Remember: It's better to be a week early on a waiver claim than a week late!
```

---

## Key Differences from Before

### Before Database Integration

- Responses were based purely on AI training data
- Could hallucinate statistics or provide outdated info
- Generic advice without specific current numbers
- No access to injury status or real-time EPA

### After Database Integration

- Responses include actual database statistics
- References specific, current numbers
- Can check real-time injury status
- Includes advanced metrics like EPA
- Data-driven recommendations with concrete evidence
