#!/bin/bash
set -e

echo "📦 Installing frontend deps..."
cd /workspace/web && npm install

echo "🚀 Starting Vite (background)..."
cd /workspace/web && npx vite --host 0.0.0.0 --port 5173 2>&1 | tee /tmp/vite.log &

echo "🔥 Starting Air (Go live reload)..."
cd /workspace && air 2>&1 | tee /tmp/air.log &

echo "✅ All services started! Logs are now streamed to stdout/stderr and /tmp/*.log"
echo "📍 App: http://localhost"
echo "Go backend: http://localhost:8080"
echo "Vite frontend: http://localhost:5173"

echo "📌 Use 'docker logs -f <container>' to suivre les logs, ou 'tail -f /tmp/vite.log /tmp/air.log' dans le conteneur"

# Keep the container alive by waiting for background jobs
wait