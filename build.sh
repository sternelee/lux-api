#!/bin/bash

# Build script for lux-api Lambda function

set -e

# Set Go environment variables for Lambda
export GOOS=linux
export GOARCH=arm64
export CGO_ENABLED=0

# Build the binary
echo "Building lux-api for Lambda..."
go build -o bootstrap main.go

# Make the binary executable
chmod +x bootstrap

echo "Build completed successfully!"
echo "The bootstrap binary is ready for deployment to AWS Lambda."
