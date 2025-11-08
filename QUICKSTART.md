# 🚀 Quick Start Guide - AI-ATL NFL Platform

Get your full-stack AI-powered NFL platform running in **2 minutes**.

---

## Prerequisites

- Go 1.23+ installed
- Node.js 18+ installed
- MongoDB running (local or Atlas)
- Google Gemini API key ([Get one free](https://ai.google.dev/))

---

## Step 1: Clone & Setup (30 seconds)

```bash
# Navigate to project
cd /Users/tobycoleman/aiatl/ideas/ai-atl

# Install backend dependencies
go mod download

# Install frontend dependencies
cd frontend
npm install
cd ..
```

---

## Step 2: Configure Environment (30 seconds)

### Backend `.env`

Create `/Users/tobycoleman/aiatl/ideas/ai-atl/.env`:

```bash
# MongoDB
MONGO_URI=mongodb://localhost:27017/nfl_platform
# OR use MongoDB Atlas:
# MONGO_URI=mongodb+srv://user:pass@cluster.mongodb.net/nfl_platform

# JWT Secret (any random string)
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# Gemini AI
GEMINI_API_KEY=your-gemini-api-key-here

# Server
PORT=8080

# Redis (optional, for caching)
REDIS_URL=redis://localhost:6379
```

### Frontend `.env.local`

Create `/Users/tobycoleman/aiatl/ideas/ai-atl/frontend/.env.local`:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## Step 3: Start Backend (15 seconds)

```bash
# Terminal 1
cd /Users/tobycoleman/aiatl/ideas/ai-atl
go run cmd/api/main.go
```

You should see:
```
✓ Connected to MongoDB
✓ Created indexes
✓ Server starting on :8080
```

**Backend is live at:** `http://localhost:8080`

---

## Step 4: Start Frontend (15 seconds)

```bash
# Terminal 2
cd /Users/tobycoleman/aiatl/ideas/ai-atl/frontend
npm run dev
```

You should see:
```
✓ Ready in 2s
✓ Local: http://localhost:3000
```

**Frontend is live at:** `http://localhost:3000`

---

## Step 5: Load Sample Data (Optional, 30 seconds)

```bash
# Terminal 3
cd /Users/tobycoleman/aiatl/ideas/ai-atl
go run scripts/load_sample_data.go
```

This creates:
- Sample players (Patrick Mahomes, Travis Kelce, etc.)
- Sample games
- Sample plays
- All with EPA metrics

---

## 🎉 You're Ready!

### Quick Test

1. **Open browser:** `http://localhost:3000`
2. **Click "Sign Up"**
3. **Create account:**
   - Email: `demo@test.com`
   - Username: `demo`
   - Password: `demo123`
4. **Auto-redirected to dashboard**

### Try These Features

**🤖 AI Chat**
1. Click "AI Chat" in sidebar
2. Ask: "Who should I start at WR this week?"
3. See AI response

**🧠 Game Script Predictor** ⭐
1. Click "AI Insights" in sidebar
2. Enter game ID: `2024_09_KC_BUF`
3. Click "Predict"
4. See AI game flow prediction

**👥 Players**
1. Click "Players" in sidebar
2. Filter by position: "WR"
3. See EPA metrics

**📈 Trade Analyzer**
1. Click "Trade Analyzer" in sidebar
2. Enter player IDs (comma-separated)
3. Click "Analyze Trade"
4. See AI grades

---

## 🔍 Verify Everything Works

### Backend Health Check
```bash
curl http://localhost:8080/health
# Should return: {"status":"ok"}
```

### Test API Endpoints

**Register User:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@test.com",
    "username": "testuser",
    "password": "test123"
  }'
```

**Get Players:**
```bash
# Use token from register response
curl http://localhost:8080/api/v1/players \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## 🐛 Troubleshooting

### Backend Won't Start

**Error:** `failed to connect to MongoDB`
- **Fix:** Check MongoDB is running
  ```bash
  # If using local MongoDB:
  brew services start mongodb-community
  
  # Or use MongoDB Atlas cloud (recommended)
  ```

**Error:** `Gemini API key not set`
- **Fix:** Add `GEMINI_API_KEY` to `.env`

### Frontend Won't Start

**Error:** `Cannot GET /api/v1/...`
- **Fix:** Backend not running. Start backend first.

**Error:** `EADDRINUSE: address already in use :::3000`
- **Fix:** Port 3000 in use. Kill the process:
  ```bash
  lsof -ti:3000 | xargs kill
  ```

### Can't Login/Register

**Error:** 401 Unauthorized
- **Fix:** Check JWT_SECRET is set in `.env`

**Error:** Network Error
- **Fix:** Check `NEXT_PUBLIC_API_URL` in frontend `.env.local`

### AI Features Not Working

**Error:** "Failed to fetch prediction"
- **Fix 1:** Check Gemini API key is valid
- **Fix 2:** Check internet connection (needs to call Gemini API)
- **Fix 3:** Check Gemini API rate limits (60/min free tier)

---

## 📱 Accessing From Other Devices

### On Same Network

**Find your local IP:**
```bash
# Mac/Linux
ipconfig getifaddr en0

# Example: 192.168.1.10
```

**Update frontend `.env.local`:**
```bash
NEXT_PUBLIC_API_URL=http://192.168.1.10:8080
```

**Restart frontend**, then access from phone/tablet:
```
http://192.168.1.10:3000
```

---

## 🏗️ Alternative: Using Make

If you have `make` installed:

```bash
# Setup everything
make setup

# Run backend
make run

# Load sample data
make load-sample-data
```

---

## 🚀 Deploy for Hackathon Demo

### Quick Deploy (5 minutes)

**Backend - Render:**
1. Push code to GitHub
2. Go to [render.com](https://render.com)
3. Click "New Web Service"
4. Connect GitHub repo
5. Set build command: `go build -o bin/api cmd/api/main.go`
6. Set start command: `./bin/api`
7. Add environment variables (same as `.env`)
8. Deploy!

**Frontend - Vercel:**
1. Go to [vercel.com](https://vercel.com)
2. Click "Import Project"
3. Select `frontend` directory
4. Add env var: `NEXT_PUBLIC_API_URL=https://your-render-url.com`
5. Deploy!

**Database - MongoDB Atlas:**
1. Go to [mongodb.com/cloud/atlas](https://www.mongodb.com/cloud/atlas)
2. Create free cluster
3. Get connection string
4. Update `MONGO_URI` in Render

---

## 📊 Demo Script

### For Judges (3 minutes)

**Minute 1: Introduction**
- "We built an AI-powered NFL fantasy platform"
- "Our unique feature: Game Script Prediction"
- "Uses Gemini AI + NFLverse data + EPA metrics"

**Minute 2: Demo**
- Show Game Script Predictor
- Show AI Chatbot
- Quick tour of other features

**Minute 3: Tech Stack**
- Go backend with Gin
- Next.js 14 frontend
- MongoDB database
- Google Gemini AI
- NFLverse data integration

---

## 🎓 Learning Resources

### Customize the Project

**Add a new API endpoint:**
1. Create handler in `internal/handlers/`
2. Add route in `cmd/api/main.go`
3. Test with curl

**Add a new frontend page:**
1. Create file in `app/dashboard/`
2. Add to navigation in `app/dashboard/layout.tsx`
3. Test in browser

**Add a new AI feature:**
1. Create service in `internal/services/`
2. Add handler in `internal/handlers/`
3. Add frontend page
4. Test end-to-end

### Read More
- `README.md` - Backend documentation
- `frontend/README.md` - Frontend documentation
- `claude.md` - Developer guidelines
- `PROJECT_OVERVIEW.md` - Full project overview
- `DEPLOYMENT.md` - Deployment guide

---

## ✅ Final Checklist

Before demo:
- [ ] Backend running ✅
- [ ] Frontend running ✅
- [ ] MongoDB connected ✅
- [ ] Sample data loaded ✅
- [ ] Gemini API working ✅
- [ ] Can register/login ✅
- [ ] Can chat with AI ✅
- [ ] Can predict game script ✅
- [ ] Can analyze trades ✅
- [ ] Can browse players ✅

---

## 🎉 You're All Set!

**Your AI-powered NFL platform is live and ready to demo!**

### Quick Links
- Frontend: http://localhost:3000
- Backend: http://localhost:8080
- API Docs: http://localhost:8080/api/v1/health

### Need Help?
- Check `PROJECT_OVERVIEW.md` for full details
- Review `claude.md` for coding patterns
- Check `DEPLOYMENT.md` for production setup

---

**Built for ATL Hackathon 2025** 🏈  
**Go crush it!** 🚀

