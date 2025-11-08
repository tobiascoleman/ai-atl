# Environment Variables Setup Guide

## ✅ Required Environment Files

You need to create **2 environment files** manually:

1. **Backend:** `/Users/tobycoleman/aiatl/ideas/ai-atl/.env`
2. **Frontend:** `/Users/tobycoleman/aiatl/ideas/ai-atl/frontend/.env.local`

---

## 📝 Backend `.env` File

Create this file: `/Users/tobycoleman/aiatl/ideas/ai-atl/.env`

```bash
# MongoDB Configuration
# Option 1: Local MongoDB
MONGO_URI=mongodb://localhost:27017/nfl_platform

# Option 2: MongoDB Atlas (cloud - recommended)
# MONGO_URI=mongodb+srv://username:password@cluster.mongodb.net/nfl_platform?retryWrites=true&w=majority

# JWT Secret (REQUIRED - change this!)
JWT_SECRET=my-super-secret-jwt-key-for-hackathon-2025

# Google Gemini AI API Key (REQUIRED)
# Get one free at: https://ai.google.dev/
GEMINI_API_KEY=YOUR_ACTUAL_GEMINI_API_KEY_HERE

# Server Configuration
PORT=8080

# Redis (optional - comment out if not using)
# REDIS_URL=redis://localhost:6379

# Environment
ENV=development

# Yahoo Fantasy Sports (optional, enables account linking)
# Create credentials at https://developer.yahoo.com/fantasysports/guide/#register
YAHOO_CLIENT_ID=your-yahoo-client-id
YAHOO_CLIENT_SECRET=your-yahoo-client-secret
# Callback must match your Yahoo app configuration
YAHOO_REDIRECT_URL=http://localhost:8080/api/v1/fantasy/oauth/callback
# Frontend URL used after successful linking
CLIENT_APP_URL=http://localhost:3000
```

### 🔑 Required Values

1. **MONGO_URI**
   - Local: `mongodb://localhost:27017/nfl_platform`
   - Atlas: Get from MongoDB Atlas dashboard
2. **JWT_SECRET**
   - Any random string (for hackathon, anything works)
   - Example: `hackathon-2025-secret-key`
3. **GEMINI_API_KEY** ⭐ **MOST IMPORTANT**

   - Get free at: https://ai.google.dev/
   - Click "Get API Key" → Create in new project
   - Copy the key

4. **Yahoo Fantasy Credentials (optional, required for new fantasy page)**
   - Register an app in the [Yahoo Developer Portal](https://developer.yahoo.com/apps/)
   - Enable Fantasy Sports API access and note the client ID/secret
   - Set callback URL to `http://localhost:8080/api/v1/fantasy/oauth/callback`
   - Update `CLIENT_APP_URL` if your frontend runs on a different host

---

## 📝 Frontend `.env.local` File

Create this file: `/Users/tobycoleman/aiatl/ideas/ai-atl/frontend/.env.local`

```bash
# Backend API URL
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### 🔑 Required Values

1. **NEXT_PUBLIC_API_URL**
   - Local development: `http://localhost:8080`
   - Production: Your deployed backend URL

---

## 🚀 Quick Setup Commands

### Option 1: Manual Creation

```bash
# Backend .env
cd /Users/tobycoleman/aiatl/ideas/ai-atl
cat > .env << 'EOF'
MONGO_URI=mongodb://localhost:27017/nfl_platform
JWT_SECRET=hackathon-2025-secret-key
GEMINI_API_KEY=YOUR_KEY_HERE
PORT=8080
ENV=development
EOF

# Frontend .env.local
cd /Users/tobycoleman/aiatl/ideas/ai-atl/frontend
cat > .env.local << 'EOF'
NEXT_PUBLIC_API_URL=http://localhost:8080
EOF
```

### Option 2: Using Text Editor

```bash
# Backend
nano /Users/tobycoleman/aiatl/ideas/ai-atl/.env
# Or use VSCode, TextEdit, etc.

# Frontend
nano /Users/tobycoleman/aiatl/ideas/ai-atl/frontend/.env.local
```

---

## ✅ Verify Your Setup

### 1. Check Files Exist

```bash
# Backend
ls -la /Users/tobycoleman/aiatl/ideas/ai-atl/.env

# Frontend
ls -la /Users/tobycoleman/aiatl/ideas/ai-atl/frontend/.env.local
```

### 2. Check Backend Loads Variables

```bash
cd /Users/tobycoleman/aiatl/ideas/ai-atl
go run cmd/api/main.go
```

**Should see:**

```
✓ Connected to MongoDB
✓ Created indexes
✓ Server starting on :8080
```

**If you see errors:**

- `MongoDB connection failed` → Check MONGO_URI
- `Gemini API key not set` → Check GEMINI_API_KEY
- `JWT secret not set` → Check JWT_SECRET

### 3. Check Frontend Loads Variables

```bash
cd /Users/tobycoleman/aiatl/ideas/ai-atl/frontend
npm run dev
```

**Should see:**

```
✓ Ready in 2s
✓ Local: http://localhost:3000
```

**Test API connection:**

- Open http://localhost:3000
- Try to register a user
- Should work without CORS errors

---

## 🔧 Common Issues

### Issue: "MongoDB connection failed"

**Solutions:**

1. **Using Local MongoDB?**

   ```bash
   # Install (Mac)
   brew install mongodb-community

   # Start MongoDB
   brew services start mongodb-community

   # Verify it's running
   mongosh
   ```

2. **Using MongoDB Atlas?** (Recommended)
   - Go to https://www.mongodb.com/cloud/atlas
   - Create free cluster
   - Click "Connect" → "Connect your application"
   - Copy connection string
   - Replace `<password>` with your password
   - Update MONGO_URI in `.env`

### Issue: "Gemini API key not set" or "401 Unauthorized" from Gemini

**Solution:**

1. Go to https://ai.google.dev/
2. Click "Get API Key in Google AI Studio"
3. Click "Create API Key"
4. Copy the key (starts with `AIza...`)
5. Update GEMINI_API_KEY in `.env`

### Issue: Frontend can't connect to backend

**Check:**

1. Backend is running on port 8080
2. Frontend `.env.local` has `NEXT_PUBLIC_API_URL=http://localhost:8080`
3. No typos in the URL
4. Restart frontend after changing `.env.local`

### Issue: "CORS error" in browser console

**Solution:**

- Backend has CORS middleware (already included)
- Make sure frontend URL matches in backend CORS config
- Restart both servers

---

## 🎯 Production Setup

When deploying:

### Backend (e.g., Render, Railway)

Add these environment variables in the dashboard:

```
MONGO_URI=mongodb+srv://user:pass@cluster.mongodb.net/nfl_platform
JWT_SECRET=super-secure-random-string-for-production
GEMINI_API_KEY=your-actual-key
PORT=8080
ENV=production
```

### Frontend (e.g., Vercel, Netlify)

Add this environment variable:

```
NEXT_PUBLIC_API_URL=https://your-backend-url.onrender.com
```

---

## 📋 Checklist

Before running the app:

- [ ] Backend `.env` file created
- [ ] Frontend `.env.local` file created
- [ ] MONGO_URI is set (local or Atlas)
- [ ] JWT_SECRET is set (any string)
- [ ] GEMINI_API_KEY is set (from Google AI)
- [ ] PORT is set to 8080
- [ ] NEXT_PUBLIC_API_URL points to backend
- [ ] MongoDB is running (if using local)
- [ ] Files are NOT committed to git (they're in .gitignore)

---

## 🔒 Security Notes

**NEVER commit these files to git!**

They're already in `.gitignore`, but double-check:

```bash
# Should return nothing
git status | grep ".env"
```

If you see `.env` files in git status:

```bash
git rm --cached .env
git rm --cached frontend/.env.local
```

---

## 🆘 Quick Test

After setup, test everything:

```bash
# Terminal 1: Start backend
cd /Users/tobycoleman/aiatl/ideas/ai-atl
go run cmd/api/main.go

# Terminal 2: Start frontend
cd /Users/tobycoleman/aiatl/ideas/ai-atl/frontend
npm run dev

# Terminal 3: Test API
curl http://localhost:8080/health
# Should return: {"status":"ok"}

# Browser: Test frontend
# Open http://localhost:3000
# Try to register/login
```

---

## 📞 Still Having Issues?

1. Check `QUICKSTART.md` for step-by-step setup
2. Review backend logs for specific error messages
3. Check browser console for frontend errors
4. Verify all services are running:
   - MongoDB (if local)
   - Backend on :8080
   - Frontend on :3000

---

**Once these files are created correctly, everything should work!** 🚀
