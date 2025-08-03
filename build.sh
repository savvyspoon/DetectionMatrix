#!/bin/bash
# Build script for RiskMatrix

# Ensure CGO is enabled for go-sqlite3
export CGO_ENABLED=1

# Build the application
echo "Building RiskMatrix with CGO enabled..."
go build -o server ./cmd/server

echo "Build complete. Run with: ./server"