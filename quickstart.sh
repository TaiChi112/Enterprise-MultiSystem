#!/usr/bin/env bash
# ============================================================================
# POS & WMS MVP - QUICK START GUIDE
# ============================================================================

echo "🚀 POS & WMS MVP - Quick Start"
echo ""

# Check prerequisites
echo "📋 Checking prerequisites..."

if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose"
    exit 1
fi

echo "✓ Go $(go version | awk '{print $3}')"
echo "✓ Docker $(docker --version | awk '{print $3}')"
echo "✓ Docker Compose $(docker-compose --version | awk '{print $3}')"
echo ""

# Step 1: Start PostgreSQL
echo "📦 Step 1: Starting PostgreSQL..."
docker-compose up -d
echo ""

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if docker-compose exec -T postgres pg_isready -U postgres &> /dev/null; then
        echo "✓ PostgreSQL is ready"
        break
    fi
    sleep 1
done
echo ""

# Step 2: Build the application
echo "🔨 Step 2: Building application..."
go build -o bin/api ./services/pos-api/cmd/api/main.go
if [ $? -eq 0 ]; then
    echo "✓ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi
echo ""

# Step 3: Run the application
echo "🚀 Step 3: Starting server..."
echo ""
echo "📌 Server will run on http://localhost:3000"
echo ""
echo "📚 API Documentation: http://localhost:3000/api/health"
echo ""
echo "Try these endpoints:"
echo "  GET    http://localhost:3000/api/health"
echo "  POST   http://localhost:3000/api/branches      (create branch)"
echo "  POST   http://localhost:3000/api/products      (create product)"
echo "  POST   http://localhost:3000/api/sales         (process sale)"
echo ""
echo "Run test-api.sh for automated testing"
echo ""
echo "To stop: Press Ctrl+C"
echo ""

./bin/api
