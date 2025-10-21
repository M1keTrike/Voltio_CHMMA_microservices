# Script de PowerShell para ejecutar todos los productores de prueba
# Ejecutar desde el directorio test_producers

Write-Host "🚀 Iniciando todos los productores de prueba..." -ForegroundColor Green

# Función para ejecutar un productor en una nueva ventana de PowerShell
function Start-Producer {
    param(
        [string]$ProducerName,
        [string]$ProducerPath,
        [string]$Icon
    )
    
    Write-Host "$Icon Iniciando $ProducerName..." -ForegroundColor Cyan
    
    # Ejecutar en una nueva ventana de PowerShell
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$ProducerPath'; Write-Host '$Icon $ProducerName Producer' -ForegroundColor Yellow; go run main.go"
}

# Obtener la ruta base
$BasePath = Get-Location

# Iniciar cada productor en su propia ventana
Start-Producer "DHT22 (Temp/Humedad)" "$BasePath\dht22" "🌡️"
Start-Sleep -Seconds 2

Start-Producer "Light Sensor" "$BasePath\light" "💡"
Start-Sleep -Seconds 2

Start-Producer "PIR Motion" "$BasePath\pir" "🚶"
Start-Sleep -Seconds 2

Start-Producer "PZEM Electric" "$BasePath\pzem" "⚡"

Write-Host "`n✅ Todos los productores han sido iniciados en ventanas separadas" -ForegroundColor Green
Write-Host "💡 Tip: Cada productor se ejecuta en su propia ventana de PowerShell" -ForegroundColor Yellow
Write-Host "🛑 Para detener todos los productores, cierra las ventanas de PowerShell correspondientes" -ForegroundColor Red

# Mantener la ventana principal abierta
Write-Host "`nPresiona cualquier tecla para salir..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
