# ==============================================================================
# Script PowerShell para Iniciar Microservicios Go de Voltio Backend
# ==============================================================================

Write-Host "🚀 Iniciando Voltio Backend Microservicios Go..." -ForegroundColor Cyan
Write-Host ""

# Verificar que Docker esté corriendo
Write-Host "🔍 Verificando Docker..." -ForegroundColor Yellow
$dockerRunning = docker info 2>$null
if (-not $dockerRunning) {
    Write-Host "❌ Docker no está corriendo. Por favor inicia Docker Desktop." -ForegroundColor Red
    exit 1
}
Write-Host "✅ Docker está activo" -ForegroundColor Green
Write-Host ""

# Verificar servicios existentes
Write-Host "🔍 Verificando servicios existentes..." -ForegroundColor Yellow
$containers = docker ps --format "{{.Names}}" 2>$null

$postgresRunning = $containers -match "postgres-local"
$influxdbRunning = $containers -match "voltio-influxdb"
$rabbitmqRunning = $containers -match "voltio-rabbitmq"

if (-not $postgresRunning) {
    Write-Host "⚠️  WARNING: PostgreSQL (postgres-local) no está corriendo" -ForegroundColor Yellow
    Write-Host "   Los microservicios necesitan PostgreSQL para funcionar." -ForegroundColor Yellow
}

if (-not $influxdbRunning) {
    Write-Host "⚠️  WARNING: InfluxDB (voltio-influxdb) no está corriendo" -ForegroundColor Yellow
    Write-Host "   Los consumers necesitan InfluxDB para almacenar métricas." -ForegroundColor Yellow
}

if (-not $rabbitmqRunning) {
    Write-Host "⚠️  WARNING: RabbitMQ (voltio-rabbitmq) no está corriendo" -ForegroundColor Yellow
    Write-Host "   Los microservicios necesitan RabbitMQ para mensajería." -ForegroundColor Yellow
}

if ($postgresRunning -and $influxdbRunning -and $rabbitmqRunning) {
    Write-Host "✅ Todos los servicios externos están corriendo" -ForegroundColor Green
} else {
    Write-Host ""
    $continue = Read-Host "¿Deseas continuar de todas formas? (s/n) [n]"
    if ($continue -ne "s" -and $continue -ne "S") {
        Write-Host "❌ Operación cancelada. Por favor inicia los servicios necesarios primero." -ForegroundColor Red
        exit 1
    }
}

Write-Host ""

# Verificar si existe .env, si no, copiar de .env.example
if (-not (Test-Path ".env")) {
    Write-Host "📝 Creando archivo .env desde .env.example..." -ForegroundColor Yellow
    Copy-Item ".env.example" ".env"
    Write-Host "✅ Archivo .env creado" -ForegroundColor Green
    Write-Host ""
}

# Preguntar si desea reconstruir las imágenes
$rebuild = Read-Host "¿Deseas reconstruir las imágenes? (s/n) [s]"
$buildFlag = ""
if ($rebuild -ne "n" -and $rebuild -ne "N") {
    $buildFlag = "--build"
    Write-Host "🔨 Se reconstruirán las imágenes..." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "🐳 Levantando microservicios Go con Docker Compose..." -ForegroundColor Cyan
docker-compose -f docker-compose.local.yml up -d $buildFlag

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ ¡Microservicios iniciados correctamente!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📊 Microservicios Corriendo:" -ForegroundColor Cyan
    Write-Host "  ✅ WebSocket Server:       http://localhost:8081" -ForegroundColor White
    Write-Host "  ✅ Automation Engine:       (sin puerto expuesto)" -ForegroundColor White
    Write-Host "  ✅ PIR Consumer:            (sin puerto expuesto)" -ForegroundColor White
    Write-Host "  ✅ DHT22 Consumer:          (sin puerto expuesto)" -ForegroundColor White
    Write-Host "  ✅ Light Consumer:          (sin puerto expuesto)" -ForegroundColor White
    Write-Host "  ✅ PZEM Consumer:           (sin puerto expuesto)" -ForegroundColor White
    Write-Host "  ✅ Notification Consumer:   (sin puerto expuesto)" -ForegroundColor White
    Write-Host ""
    Write-Host "🔗 Servicios Externos (ya corriendo):" -ForegroundColor Cyan
    Write-Host "  - PostgreSQL:   localhost:5432 (mike/trike)" -ForegroundColor White
    Write-Host "  - InfluxDB UI:  http://localhost:8086" -ForegroundColor White
    Write-Host "  - RabbitMQ UI:  http://localhost:15672 (admin/trike)" -ForegroundColor White
    Write-Host "  - API Voltio:   http://localhost:8000" -ForegroundColor White
    Write-Host ""
    Write-Host "📝 Comandos útiles:" -ForegroundColor Cyan
    Write-Host "  Ver logs:       docker-compose -f docker-compose.local.yml logs -f" -ForegroundColor White
    Write-Host "  Ver estado:     docker-compose -f docker-compose.local.yml ps" -ForegroundColor White
    Write-Host "  Detener:        .\stop-local.ps1" -ForegroundColor White
    Write-Host ""
    
    # Preguntar si desea ver los logs
    $viewLogs = Read-Host "¿Deseas ver los logs en tiempo real? (s/n) [s]"
    if ($viewLogs -ne "n" -and $viewLogs -ne "N") {
        Write-Host ""
        Write-Host "📋 Mostrando logs (Ctrl+C para salir)..." -ForegroundColor Cyan
        docker-compose -f docker-compose.local.yml logs -f
    }
} else {
    Write-Host ""
    Write-Host "❌ Error al iniciar los microservicios" -ForegroundColor Red
    Write-Host "Ver logs con: docker-compose -f docker-compose.local.yml logs" -ForegroundColor Yellow
    exit 1
}
