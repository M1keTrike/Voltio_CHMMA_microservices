@echo off
echo 🚀 Iniciando todos los productores de prueba...
echo.

echo 🌡️ Iniciando DHT22 Producer...
start "DHT22 Producer" cmd /k "cd /d %~dp0dht22 && echo 🌡️ DHT22 Producer && go run main.go"

timeout /t 2 /nobreak >nul

echo 💡 Iniciando Light Sensor Producer...
start "Light Sensor Producer" cmd /k "cd /d %~dp0light && echo 💡 Light Sensor Producer && go run main.go"

timeout /t 2 /nobreak >nul

echo 🚶 Iniciando PIR Motion Producer...
start "PIR Motion Producer" cmd /k "cd /d %~dp0pir && echo 🚶 PIR Motion Producer && go run main.go"

timeout /t 2 /nobreak >nul

echo ⚡ Iniciando PZEM Electric Producer...
start "PZEM Electric Producer" cmd /k "cd /d %~dp0pzem && echo ⚡ PZEM Electric Producer && go run main.go"

echo.
echo ✅ Todos los productores han sido iniciados en ventanas separadas
echo 💡 Tip: Cada productor se ejecuta en su propia ventana de CMD
echo 🛑 Para detener todos los productores, cierra las ventanas de CMD correspondientes
echo.
pause
