# Script para ejecutar un solo productor
# Uso: .\start_single_producer.ps1 <nombre_del_productor>
# Ejemplo: .\start_single_producer.ps1 dht22

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("dht22", "light", "pir", "pzem")]
    [string]$Producer
)

# Mapeo de productores a iconos y nombres
$ProducerInfo = @{
    "dht22" = @{ Name = "DHT22 (Temperatura/Humedad)"; Icon = "🌡️" }
    "light" = @{ Name = "Light Sensor"; Icon = "💡" }
    "pir"   = @{ Name = "PIR Motion Sensor"; Icon = "🚶" }
    "pzem"  = @{ Name = "PZEM Electric Meter"; Icon = "⚡" }
}

$Info = $ProducerInfo[$Producer]
$ProducerPath = Join-Path (Get-Location) $Producer

Write-Host "$($Info.Icon) Iniciando $($Info.Name)..." -ForegroundColor Green

# Verificar que el directorio existe
if (!(Test-Path $ProducerPath)) {
    Write-Host "❌ Error: No se encontró el directorio $ProducerPath" -ForegroundColor Red
    exit 1
}

# Cambiar al directorio del productor y ejecutar
Set-Location $ProducerPath
Write-Host "📂 Directorio: $ProducerPath" -ForegroundColor Cyan
Write-Host "🚀 Ejecutando: go run main.go" -ForegroundColor Yellow
Write-Host "🛑 Presiona Ctrl+C para detener el productor`n" -ForegroundColor Red

# Ejecutar el productor
go run main.go
