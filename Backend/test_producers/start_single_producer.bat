@echo off
if "%1"=="" (
    echo ❌ Error: Debes especificar el tipo de productor
    echo.
    echo Uso: %0 ^<tipo_productor^>
    echo.
    echo Tipos disponibles:
    echo   dht22  - 🌡️ DHT22 (Temperatura/Humedad)
    echo   light  - 💡 Light Sensor
    echo   pir    - 🚶 PIR Motion Sensor
    echo   pzem   - ⚡ PZEM Electric Meter
    echo.
    echo Ejemplo: %0 dht22
    pause
    exit /b 1
)

set PRODUCER=%1

if not exist "%~dp0%PRODUCER%" (
    echo ❌ Error: No se encontró el directorio del productor '%PRODUCER%'
    echo.
    echo Verifica que el productor exista en: %~dp0%PRODUCER%
    pause
    exit /b 1
)

if "%PRODUCER%"=="dht22" (
    set ICON=🌡️
    set NAME=DHT22 (Temperatura/Humedad)
) else if "%PRODUCER%"=="light" (
    set ICON=💡
    set NAME=Light Sensor
) else if "%PRODUCER%"=="pir" (
    set ICON=🚶
    set NAME=PIR Motion Sensor
) else if "%PRODUCER%"=="pzem" (
    set ICON=⚡
    set NAME=PZEM Electric Meter
) else (
    echo ❌ Error: Tipo de productor '%PRODUCER%' no válido
    pause
    exit /b 1
)

echo %ICON% Iniciando %NAME%...
echo.
echo 📂 Directorio: %~dp0%PRODUCER%
echo 🚀 Ejecutando: go run main.go
echo 🛑 Presiona Ctrl+C para detener el productor
echo.

cd /d "%~dp0%PRODUCER%"
go run main.go
