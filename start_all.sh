#!/bin/bash

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "🚀 Starting AI-ATL Application..."
echo ""

# Kill any existing processes
pkill -f "python3.*app.py"
pkill -f "go run.*main.go"
pkill -f "next dev"
sleep 2

# Start Flask service
echo "▶️  Starting Flask ESPN service (port 5002)..."
python3 app.py > flask.log 2>&1 &
FLASK_PID=$!
sleep 3

if ps -p $FLASK_PID > /dev/null; then
   echo "✅ Flask service started (PID: $FLASK_PID)"
else
   echo "❌ Flask failed to start"
   tail -10 flask.log
fi

# Start Go API
echo "▶️  Starting Go API backend (port 8080)..."
go run ./cmd/api/main.go > api.log 2>&1 &
GO_PID=$!
sleep 4

if ps -p $GO_PID > /dev/null; then
   echo "✅ Go API started (PID: $GO_PID)"
else
   echo "❌ Go API failed to start"
   tail -10 api.log
fi

# Start Next.js
echo "▶️  Starting Next.js frontend (port 3000)..."
cd frontend
npm run dev > ../frontend.log 2>&1 &
NEXT_PID=$!
cd ..
sleep 5

if ps -p $NEXT_PID > /dev/null; then
   echo "✅ Next.js started (PID: $NEXT_PID)"
else
   echo "❌ Next.js failed to start"
   tail -10 frontend.log
fi

echo ""
echo "🎉 Application started!"
echo ""
echo "📍 Access points:"
echo "   Frontend:  http://localhost:3000"
echo "   Go API:    http://localhost:8080"
echo "   Flask API: http://localhost:5002"
echo ""
echo "💡 To stop all services, run: pkill -f 'python3.*app.py|go run|next dev'"
