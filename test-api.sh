#!/bin/bash

# Test script for lux-api

echo "Testing Lux API..."
echo ""

# Start wrangler dev in background
echo "Starting Cloudflare Workers development server..."
npx wrangler dev --port 8787 &
WRANGLER_PID=$!

# Wait for server to start
echo "Waiting for server to start..."
sleep 5

# Test health endpoint
echo "Testing health endpoint..."
curl -s http://localhost:8787/api/health | jq '.' || echo "Health check failed"

echo ""
echo "Testing video extraction (this may take a moment)..."
# Test with a simple URL (using a test URL)
curl -s -X POST http://localhost:8787/api/extract \
  -H "Content-Type: application/json" \
  -d '{"url":"https://www.youtube.com/watch?v=dQw4w9WgXcQ"}' | jq '.' || echo "Extraction test failed"

# Kill the wrangler process
echo ""
echo "Stopping development server..."
kill $WRANGLER_PID 2>/dev/null

echo "Test complete!"