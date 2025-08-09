#!/bin/bash

# Build script for compiling lux to WASM for Cloudflare Workers

echo "Building lux WASM module for Cloudflare Workers..."

# Set environment variables for WASM build
export GOOS=js
export GOARCH=wasm

# Build the WASM module
echo "Compiling Go code to WASM..."
go build -o worker.wasm -ldflags="-s -w" -tags wasm main.wasm.go

if [ $? -eq 0 ]; then
    echo "✅ WASM module built successfully: worker.wasm"
    echo "Size: $(ls -lh worker.wasm | awk '{print $5}')"
else
    echo "❌ Failed to build WASM module"
    exit 1
fi

# Optional: Optimize the WASM module using wasm-opt if available
if command -v wasm-opt &> /dev/null; then
    echo "Attempting to optimize WASM module with wasm-opt..."
    wasm-opt -O3 -o worker-optimized.wasm worker.wasm 2>/dev/null
    if [ $? -eq 0 ]; then
        mv worker-optimized.wasm worker.wasm
        echo "✅ WASM module optimized"
        echo "Optimized size: $(ls -lh worker.wasm | awk '{print $5}')"
    else
        echo "⚠️  wasm-opt optimization failed, using unoptimized version"
    fi
else
    echo "ℹ️  wasm-opt not found. Skipping optimization."
    echo "   Install with: npm install -g wasm-opt"
fi

echo ""
echo "Build complete! Next steps:"
echo "1. Deploy to Cloudflare Workers using: wrangler publish"
echo "2. Or test locally using: wrangler dev"