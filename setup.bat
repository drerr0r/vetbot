@echo off
echo 🐾 Setting up VetBot on Windows...

:: Check if Go is installed
go version >nul 2>&1
if errorlevel 1 (
    echo ❌ Go is not installed. Please install Go 1.21+
    pause
    exit /b 1
)

echo ✅ Go is installed

:: Check if Docker is running
docker version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker is not running. Please start Docker Desktop
    pause
    exit /b 1
)

echo ✅ Docker is running

:: Create .env file if it doesn't exist
if not exist ".env" (
    echo 📝 Creating .env file from .env.example
    copy .env.example .env
    echo ⚠️ Please edit .env file with your actual values
) else (
    echo ✅ .env file already exists
)

echo 📦 Downloading Go dependencies...
go mod download

echo 🐘 Starting PostgreSQL database...
docker-compose -f docker-compose.windows.yml up -d postgres

echo ⏳ Waiting for database to be ready...
timeout /t 15 /nobreak

echo 🗃️ Running database migrations...
for %%f in (migrations\*.sql) do (
    echo Applying %%~nxf...
    type "%%f" | docker exec -i vetbot_db psql -U vetbot_user -d vetbot
)

echo 🔨 Building application...
go build -o vetbot.exe ./cmd/vetbot

if exist "vetbot.exe" (
    echo ✅ Application built successfully!
) else (
    echo ❌ Build failed
)

echo 🎉 Setup completed successfully!
echo.
echo Next steps:
echo 1. Edit .env file with your Telegram Bot Token
echo 2. Run: vetbot.exe
echo 3. Or run with Docker: docker-compose up
pause