#!/bin/bash

# VetBot Deployment Script
set -e

ENV=${1:-production}
DOCKER_IMAGE="vetbot/app:latest"

echo "🚀 Deploying VetBot ($ENV environment)..."

# Validate environment
if [[ ! "$ENV" =~ ^(development|staging|production)$ ]]; then
    echo "❌ Invalid environment: $ENV"
    echo "Usage: $0 [development|staging|production]"
    exit 1
fi

echo "📦 Building Docker image..."
docker build -t $DOCKER_IMAGE .

# Stop existing container
echo "🛑 Stopping existing container..."
docker-compose down || true

# Update images
echo "🔄 Pulling latest images..."
docker-compose pull

# Start services
echo "🚀 Starting services..."
docker-compose up -d

# Health check
echo "🏥 Performing health check..."
sleep 10

# Check if application is running
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Deployment successful!"
else
    echo "❌ Deployment failed - application not responding"
    docker-compose logs vetbot
    exit 1
fi

# Clean up old images
echo "🧹 Cleaning up old images..."
docker image prune -f

echo "🎉 VetBot deployed successfully to $ENV environment!"