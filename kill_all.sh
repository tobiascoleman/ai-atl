#!/bin/bash

# Kill all AI-ATL services
echo "🛑 Stopping all AI-ATL services..."

# Kill Flask service
pkill -9 -f "python3.*app.py"
echo "✓ Flask service stopped"

# Kill Go API service
pkill -9 -f "go run"
echo "✓ Go API service stopped"

# Kill Next.js frontend
pkill -9 -f "next dev"
echo "✓ Next.js frontend stopped"

# Sleep briefly to allow processes to terminate
sleep 2

# Force kill any remaining processes on the ports
lsof -ti:5002,8080,3000 | xargs kill -9 2>/dev/null

echo "✅ All services killed"
