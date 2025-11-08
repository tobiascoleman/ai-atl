# AI-ATL NFL Platform

> AI-Powered NFL Fantasy & Analytics Platform - Built for ATL Hackathon 2025

## 🎯 What We Built

An NFL fantasy platform that predicts **HOW games will unfold** using AI, not just final scores. We combine NFLverse's advanced EPA metrics with Google Gemini AI to give fantasy players quantified, data-driven insights.

### Novel Features

1. **AI Game Script Predictor** - Predicts game flow quarter-by-quarter and player usage changes
2. **EPA-Based Analysis** - Uses Expected Points Added metrics to find undervalued players
3. **Context-Aware Chatbot** - Fantasy advice based on your actual lineup and league settings
4. **Injury Impact Predictor** - Quantifies how injuries affect backup player opportunities
5. **Yahoo Fantasy Sync (PoC)** - Securely link your Yahoo account to preview active NFL fantasy teams

## 🏗️ Tech Stack

- **Backend**: Go 1.21+ with Gin framework
- **Database**: MongoDB
- **AI**: Google Gemini API
- **Data Source**: NFLverse (play-by-play, EPA, Next Gen Stats)
- **Auth**: JWT

## 🚀 Quick Start

### Prerequisites

- Go 1.23+
- MongoDB (or use Docker)
- Gemini API key

### Installation

```bash
# Clone the repo
git clone https://github.com/ai-atl/nfl-platform
cd nfl-platform

# Install dependencies
go mod download

# Set up environment variables
cp .env.example .env
# Edit .env with your Gemini API key

# Start MongoDB (if using Docker)
docker run -d -p 27017:27017 --name mongodb mongo

# Run the API
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

### 📊 Load Data

**Maximum data (26 seasons, 1M+ plays, 30-60 min)**:
```bash
make load-maximum-data
```

> 🔐 **Yahoo Fantasy OAuth**: add `YAHOO_CLIENT_ID`, `YAHOO_CLIENT_SECRET`, `YAHOO_REDIRECT_URL`, and `CLIENT_APP_URL` to your `.env` to enable the new fantasy integration. See `ENV_SETUP.md` for full instructions.

### API Endpoints

#### Authentication

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","username":"user","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

#### AI Features (require auth token)

```bash
# Game Script Prediction
curl -X GET "http://localhost:8080/api/v1/insights/game_script?game_id=123" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Ask the AI Chatbot
curl -X POST http://localhost:8080/api/v1/chatbot/ask \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"question":"Who should I start at RB this week?"}'

# Trade Analyzer
curl -X POST http://localhost:8080/api/v1/trades/analyze \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"team_a_gives":["player1"],"team_a_gets":["player2"],"team_b_gives":["player2"],"team_b_gets":["player1"]}'

# Fantasy Teams (Yahoo OAuth required)
curl -X GET http://localhost:8080/api/v1/fantasy/teams \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 📁 Project Structure

```
cmd/api/                  # Application entry point
internal/
├── models/              # Data models (User, Player, Game, etc.)
├── handlers/            # HTTP request handlers
├── services/            # Business logic (AI services, analytics)
├── middleware/          # Auth, CORS, logging
└── config/              # Configuration management
pkg/
├── gemini/             # Gemini API client
├── nflverse/           # NFLverse data fetching
└── mongodb/            # Database utilities
```

## 🎨 Key Features Explained

### 1. AI Game Script Predictor

Traditional fantasy tools only show projections. We predict **how** the game will unfold:

```json
{
  "game_id": "KC_vs_BUF",
  "predicted_flow": "Chiefs likely to build early lead, leading to more passing from Buffalo in 2nd half",
  "player_impacts": [
    {
      "player_name": "Josh Allen",
      "impact": "+30% passing attempts in 2nd half",
      "reasoning": "Playing from behind, game script favors passing"
    }
  ],
  "confidence_score": 0.85
}
```

### 2. EPA-Based Player Analysis

We use Expected Points Added (EPA) from NFLverse to find efficient players before the market catches on:

- **EPA per play**: Measures true player efficiency
- **Success rate**: Consistency metric
- **Target/snap share**: Opportunity metrics

### 3. Context-Aware AI Chatbot

Unlike generic chatbots, ours knows your roster and gives personalized advice:

```
You: "Should I start Player X or Player Y at RB?"
Bot: "Based on your roster and this week's matchups, I recommend Player X.
      He faces a defense ranked 28th against the run, and the game script
      favors a heavy rushing attack (team is -7 favorites). Player Y faces
      a tougher matchup and his team is likely to be trailing."
```

## 🧠 The AI Advantage

### What Makes This Different?

1. **Historical Pattern Matching**: We analyze 3+ years of play-by-play data to find similar game scripts
2. **Quantified Predictions**: "30% more touches" not "should see increased opportunity"
3. **Multi-Factor Analysis**: Combines Vegas lines, weather, injuries, team tendencies
4. **Contextual Reasoning**: AI explains the "why" behind every prediction

## 🏗️ Architecture Decisions

### Why Go?

- Fast compilation for rapid iteration
- Excellent concurrency for handling multiple AI requests
- Simple deployment (single binary)
- Strong MongoDB driver support

### Why MongoDB?

- Flexible schema for NFLverse data
- Embedded documents for player stats
- Fast queries with proper indexing
- Easy to scale

### Why Gemini?

- Cost-effective (~$0.001 per request)
- Good at structured reasoning
- Fast response times
- JSON output support

## 📊 Data Sources

All data comes from [NFLverse](https://github.com/nflverse), which provides:

- **Play-by-play data** with EPA/WPA calculations
- **Player stats** (weekly and seasonal)
- **Next Gen Stats** (player tracking data)
- **Injury reports**
- **Betting lines**

## 🔐 Security

- JWT-based authentication
- Password hashing with bcrypt
- MongoDB connection pooling
- Rate limiting (TODO: implement)

## 🚧 Future Enhancements

- [ ] Real-time score updates via WebSocket
- [ ] Social features (community voting, leaderboards)
- [ ] Mobile app (React Native)
- [ ] Historical accuracy tracking
- [ ] Fine-tuned ML models for predictions
- [ ] Integration with actual fantasy platforms (ESPN, Yahoo)

## 👥 Team

Built by Team AI-ATL for the ATL Hackathon 2025

## 📄 License

MIT License - see LICENSE file for details

## 🙏 Acknowledgments

- [NFLverse](https://github.com/nflverse) for providing amazing NFL data
- [Google Gemini](https://ai.google.dev/) for the AI API
- The Go and MongoDB communities

---

**Built with ❤️ in 36 hours**
