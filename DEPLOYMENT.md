# Deployment Guide - AI-ATL NFL Platform

## Quick Start (Local Development)

### 1. Prerequisites
```bash
# Install Go 1.21+
go version

# Install MongoDB (or use Docker)
docker run -d -p 27017:27017 --name nfl-mongodb mongo:latest

# Get Gemini API Key
# Visit: https://makersuite.google.com/app/apikey
```

### 2. Setup
```bash
# Clone and navigate
git clone https://github.com/ai-atl/nfl-platform
cd nfl-platform

# Install dependencies
go mod download

# Create .env file
cp .env.example .env
# Edit .env and add your GEMINI_API_KEY
```

### 3. Initialize Database
```bash
# Create indexes
go run scripts/create_indexes.go

# Load sample data (optional)
go run scripts/load_sample_data.go
```

### 4. Run
```bash
# Start the API
go run cmd/api/main.go

# Or use Make
make run

# Server runs on http://localhost:8080
```

## Testing the API

### Health Check
```bash
curl http://localhost:8080/health
```

### Register a User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "password123"
  }'
```

### Login and Get Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### Test AI Features (use token from login)
```bash
# Game Script Prediction
curl -X GET "http://localhost:8080/api/v1/insights/game_script?game_id=2024_09_KC_BUF" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# AI Chatbot
curl -X POST http://localhost:8080/api/v1/chatbot/ask \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{"question": "Who should I start at RB this week?"}'
```

## Production Deployment

### Option 1: Railway

1. Create account at [railway.app](https://railway.app)
2. Install Railway CLI:
   ```bash
   npm install -g @railway/cli
   railway login
   ```
3. Deploy:
   ```bash
   railway init
   railway up
   ```
4. Add environment variables in Railway dashboard
5. Add MongoDB plugin in Railway

### Option 2: Render

1. Create account at [render.com](https://render.com)
2. Connect GitHub repo
3. Create new Web Service
4. Build command: `go build -o nfl-api cmd/api/main.go`
5. Start command: `./nfl-api`
6. Add environment variables
7. Create MongoDB Atlas database

### Option 3: Docker

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o nfl-api cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/nfl-api .
EXPOSE 8080
CMD ["./nfl-api"]
```

```bash
# Build and run
docker build -t nfl-platform .
docker run -p 8080:8080 --env-file .env nfl-platform
```

### Option 4: Google Cloud Run

```bash
# Install gcloud CLI
gcloud auth login

# Build and deploy
gcloud run deploy nfl-platform \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

## Environment Variables

Required for production:
```
MONGODB_URI=mongodb+srv://user:pass@cluster.mongodb.net/nfl_platform
DB_NAME=nfl_platform
JWT_SECRET=your-super-secret-jwt-key-change-this
GEMINI_API_KEY=your_gemini_api_key
REDIS_URL=redis://user:pass@redis-host:6379
ENVIRONMENT=production
PORT=8080
```

## MongoDB Setup (Production)

### MongoDB Atlas (Recommended)
1. Create account at [mongodb.com/cloud/atlas](https://www.mongodb.com/cloud/atlas)
2. Create free cluster (M0)
3. Add IP whitelist (0.0.0.0/0 for development)
4. Create database user
5. Get connection string
6. Add to MONGODB_URI environment variable

### Indexes
Indexes are automatically created on first run via `mongodb.CreateIndexes()`

Key indexes:
- players: `nfl_id` (unique), `team + position` (compound)
- games: `game_id` (unique), `season + week` (compound)
- users: `email` (unique)
- lineups: `user_id + week` (compound)

## Monitoring

### Health Checks
```bash
# Basic health
curl https://your-domain.com/health

# MongoDB connection
curl https://your-domain.com/api/v1/players?limit=1
```

### Logs
```bash
# Railway
railway logs

# Render
# View in dashboard

# Docker
docker logs container-id

# Google Cloud Run
gcloud run services logs read nfl-platform
```

## Performance Optimization

### 1. Enable Caching
Add Redis for caching Gemini responses:
```go
import "github.com/go-redis/redis/v8"

rdb := redis.NewClient(&redis.Options{
    Addr: os.Getenv("REDIS_URL"),
})
```

### 2. Connection Pooling
Already configured in `pkg/mongodb/mongodb.go`:
- MaxPoolSize: 50
- MinPoolSize: 10

### 3. Rate Limiting
Add middleware for API rate limiting:
```go
import "github.com/gin-gonic/gin"
import "golang.org/x/time/rate"

func RateLimitMiddleware() gin.HandlerFunc {
    limiter := rate.NewLimiter(10, 20) // 10 requests/sec, burst of 20
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, gin.H{"error": "Rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## Security Checklist

- [ ] Change JWT_SECRET to a strong random value
- [ ] Use HTTPS in production
- [ ] Enable MongoDB authentication
- [ ] Whitelist MongoDB IPs
- [ ] Set CORS origins (update middleware/cors.go)
- [ ] Add rate limiting
- [ ] Enable request logging
- [ ] Set up error monitoring (Sentry)
- [ ] Regular dependency updates

## Scaling

### Horizontal Scaling
The API is stateless and can be scaled horizontally:
- Add load balancer
- Deploy multiple instances
- Use shared MongoDB and Redis

### Database Scaling
- Use MongoDB Atlas auto-scaling
- Add read replicas for read-heavy workloads
- Consider sharding for large datasets

### Cost Optimization
- MongoDB Atlas M0: Free
- Gemini API: ~$0.001 per request
- Railway/Render: $5-20/month
- **Total**: < $25/month for moderate usage

## Troubleshooting

### API not starting
```bash
# Check logs
go run cmd/api/main.go

# Common issues:
# - MongoDB not running
# - Port 8080 already in use
# - Missing environment variables
```

### MongoDB connection fails
```bash
# Test connection
mongosh "mongodb://localhost:27017/nfl_platform"

# Check URI format
# mongodb://user:pass@host:port/database
```

### Gemini API errors
```bash
# Verify API key
curl https://generativelanguage.googleapis.com/v1beta/models?key=YOUR_KEY

# Check quota
# Visit: https://makersuite.google.com/app/apikey
```

## Support

For issues or questions:
- GitHub Issues: [github.com/ai-atl/nfl-platform/issues](https://github.com/ai-atl/nfl-platform/issues)
- Documentation: See README.md and claude.md
- NFLverse Data: [github.com/nflverse](https://github.com/nflverse)

---

**Built for ATL Hackathon 2025** 🚀

