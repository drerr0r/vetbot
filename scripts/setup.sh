#!/bin/bash

# VetBot Setup Script
set -e

echo "🐾 Setting up VetBot..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+"
    exit 1
fi

echo "✅ Go is installed: $(go version)"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker"
    exit 1
fi

echo "✅ Docker is installed: $(docker --version)"

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose"
    exit 1
fi

echo "✅ Docker Compose is installed: $(docker-compose --version)"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from .env.example"
    cp .env.example .env
    echo "⚠️ Please edit .env file with your actual values"
else
    echo "✅ .env file already exists"
fi

# Download dependencies
echo "📦 Downloading Go dependencies..."
go mod download

# Start database
echo "🐘 Starting PostgreSQL database..."
docker-compose up -d postgres

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
sleep 10

# Run migrations
echo "🗃️ Running database migrations..."
make migrate-up

# Build application
echo "🔨 Building application..."
make build

echo "🎉 Setup completed successfully!"
echo ""
echo "Next steps:"
echo "1. Edit .env file with your Telegram Bot Token"
echo "2. Run: make run"
echo "3. Or run with Docker: make compose-up"