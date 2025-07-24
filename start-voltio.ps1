# Voltio Services Quick Start Script for Local Development
# Run this script to quickly start all services locally

param(
    [switch]$Build,
    [switch]$Stop,
    [switch]$Logs,
    [switch]$Status
)

$ErrorActionPreference = "Stop"

Write-Host "🚀 Voltio Services Manager" -ForegroundColor Green

# Check if Docker is running
try {
    docker version | Out-Null
} catch {
    Write-Host "❌ Docker is not running. Please start Docker Desktop first." -ForegroundColor Red
    exit 1
}

# Change to script directory
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Definition
Set-Location $scriptPath

if ($Stop) {
    Write-Host "🛑 Stopping Voltio services..." -ForegroundColor Yellow
    docker-compose down
    exit 0
}

if ($Logs) {
    Write-Host "📋 Showing logs for Voltio services..." -ForegroundColor Cyan
    docker-compose logs -f
    exit 0
}

if ($Status) {
    Write-Host "📊 Checking status of Voltio services..." -ForegroundColor Cyan
    docker-compose ps
    exit 0
}

# Create .env file if it doesn't exist
if (!(Test-Path ".env")) {
    Write-Host "⚙️ Creating .env file from template..." -ForegroundColor Yellow
    Copy-Item ".env.example" ".env"
    Write-Host "✏️ Please edit .env file with your configuration before running services" -ForegroundColor Yellow
}

# Create logs directory
if (!(Test-Path "logs")) {
    New-Item -ItemType Directory -Path "logs" | Out-Null
    Write-Host "📁 Created logs directory" -ForegroundColor Green
}

if ($Build) {
    Write-Host "🔨 Building Voltio services..." -ForegroundColor Blue
    docker-compose build --no-cache
}

Write-Host "🚀 Starting Voltio services..." -ForegroundColor Green
docker-compose up -d

Write-Host ""
Write-Host "✅ Voltio services started successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "📋 Useful commands:" -ForegroundColor Cyan
Write-Host "  View logs:    .\start-voltio.ps1 -Logs" -ForegroundColor White
Write-Host "  Check status: .\start-voltio.ps1 -Status" -ForegroundColor White
Write-Host "  Stop:         .\start-voltio.ps1 -Stop" -ForegroundColor White
Write-Host "  Rebuild:      .\start-voltio.ps1 -Build" -ForegroundColor White
Write-Host ""
Write-Host "🌐 Services are available at:" -ForegroundColor Cyan
Write-Host "  WebSocket Server: http://localhost:8081" -ForegroundColor White
Write-Host ""
Write-Host "📊 Check status with: docker-compose ps" -ForegroundColor Cyan
