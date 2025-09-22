# VetBot Windows Setup Script
Write-Host "🐾 Setting up VetBot on Windows..." -ForegroundColor Green

# Check if Go is installed
$goVersion = go version 2>$null
if (-not $goVersion) {
    Write-Host "❌ Go is not installed. Please install Go 1.21+" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Go is installed: $goVersion" -ForegroundColor Green

# Check if Docker is running
try {
    $dockerVersion = docker version --format '{{.Server.Version}}' 2>$null
    if (-not $dockerVersion) {
        Write-Host "❌ Docker is not running. Please start Docker Desktop" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Docker is running: $dockerVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker is not available. Please install and start Docker Desktop" -ForegroundColor Red
    exit 1
}

# Create .env file if it doesn't exist
if (-not (Test-Path .env)) {
    Write-Host "📝 Creating .env file from .env.example" -ForegroundColor Yellow
    Copy-Item .env.example .env
    Write-Host "⚠️ Please edit .env file with your actual values" -ForegroundColor Yellow
} else {
    Write-Host "✅ .env file already exists" -ForegroundColor Green
}

# Download dependencies
Write-Host "📦 Downloading Go dependencies..." -ForegroundColor Yellow
go mod download

# Start database using Docker Compose
Write-Host "🐘 Starting PostgreSQL database..." -ForegroundColor Yellow
docker-compose up -d postgres

# Wait for database to be ready
Write-Host "⏳ Waiting for database to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 15

# Check if database is running
$dbHealthy = docker ps --filter "name=vetbot_db" --filter "health=healthy" --format "{{.Names}}"
if (-not $dbHealthy) {
    Write-Host "⚠️ Database is starting, waiting a bit more..." -ForegroundColor Yellow
    Start-Sleep -Seconds 10
}

# Run migrations manually (since make might not work well on Windows)
Write-Host "🗃️ Running database migrations..." -ForegroundColor Yellow
$migrationFiles = Get-ChildItem "migrations" -Filter "*.sql" | Sort-Object Name
foreach ($file in $migrationFiles) {
    Write-Host "Applying $($file.Name)..." -ForegroundColor Cyan
    docker exec -i vetbot_db psql -U vetbot_user -d vetbot -c "SELECT 1;" 2>$null
    if ($LASTEXITCODE -eq 0) {
        Get-Content $file.FullName | docker exec -i vetbot_db psql -U vetbot_user -d vetbot
    } else {
        Write-Host "❌ Database is not ready yet. Please run migrations manually later." -ForegroundColor Red
        break
    }
}

# Build application
Write-Host "🔨 Building application..." -ForegroundColor Yellow
go build -o vetbot.exe ./cmd/vetbot

if (Test-Path "vetbot.exe") {
    Write-Host "✅ Application built successfully!" -ForegroundColor Green
} else {
    Write-Host "❌ Build failed" -ForegroundColor Red
}

Write-Host "🎉 Setup completed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Edit .env file with your Telegram Bot Token" -ForegroundColor White
Write-Host "2. Run: .\vetbot.exe" -ForegroundColor White
Write-Host "3. Or run with Docker: docker-compose up" -ForegroundColor White