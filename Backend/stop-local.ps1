# ==============================================================================
# Script PowerShell para Detener Microservicios Go de Voltio Backend
# ==============================================================================

Write-Host "🛑 Deteniendo Voltio Backend Microservicios Go..." -ForegroundColor Yellow
Write-Host ""
Write-Host "⚠️  NOTA: Solo se detendrán los microservicios Go." -ForegroundColor Cyan
Write-Host "   PostgreSQL, InfluxDB, RabbitMQ y API seguirán corriendo." -ForegroundColor Cyan
Write-Host ""

docker-compose -f docker-compose.local.yml down

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Microservicios Go detenidos correctamente" -ForegroundColor Green
    Write-Host ""
    Write-Host "🔗 Servicios que siguen corriendo:" -ForegroundColor Cyan
    
    # Verificar qué servicios siguen corriendo
    $containers = docker ps --format "{{.Names}}" 2>$null
    
    if ($containers -match "postgres-local") {
        Write-Host "  ✅ PostgreSQL (postgres-local)" -ForegroundColor White
    }
    if ($containers -match "voltio-influxdb") {
        Write-Host "  ✅ InfluxDB (voltio-influxdb)" -ForegroundColor White
    }
    if ($containers -match "voltio-rabbitmq") {
        Write-Host "  ✅ RabbitMQ (voltio-rabbitmq)" -ForegroundColor White
    }
    if ($containers -match "voltio-api") {
        Write-Host "  ✅ API Voltio (voltio-api)" -ForegroundColor White
    }
    
} else {
    Write-Host ""
    Write-Host "❌ Error al detener los microservicios" -ForegroundColor Red
    exit 1
}
