#!/bin/bash

# ==========================================
# SCRIPT DE VERIFICACIÓN PRE-BUILD
# ==========================================
# Verifica que todos los archivos necesarios existan

echo "🔍 Verificando archivos necesarios para Docker build..."

# Función para verificar archivos
check_file() {
    if [ -f "$1" ]; then
        echo "✅ $1"
        return 0
    else
        echo "❌ $1 - FALTA"
        return 1
    fi
}

# Función para verificar directorios
check_dir() {
    if [ -d "$1" ]; then
        echo "✅ $1/"
        return 0
    else
        echo "❌ $1/ - FALTA"
        return 1
    fi
}

errors=0

echo ""
echo "📁 Verificando estructura de directorios..."

# Verificar directorios principales
check_dir "WebSocketServer" || ((errors++))
check_dir "PZEM_ConsumerSender" || ((errors++))
check_dir "DHT22_ConsumerSender" || ((errors++))
check_dir "PIR_ConsumerSender" || ((errors++))
check_dir "LightSensor_ConsumerSender" || ((errors++))
check_dir "Notification_ConsumerSender" || ((errors++))
check_dir "automation-engine" || ((errors++))

echo ""
echo "📋 Verificando archivos go.mod..."

# Verificar go.mod files
check_file "WebSocketServer/go.mod" || ((errors++))
check_file "PZEM_ConsumerSender/go.mod" || ((errors++))
check_file "DHT22_ConsumerSender/go.mod" || ((errors++))
check_file "PIR_ConsumerSender/go.mod" || ((errors++))
check_file "LightSensor_ConsumerSender/go.mod" || ((errors++))
check_file "Notification_ConsumerSender/go.mod" || ((errors++))
check_file "automation-engine/go.mod" || ((errors++))

echo ""
echo "🚀 Verificando archivos principales..."

# WebSocket Server
check_file "WebSocketServer/cmd/main.go" || ((errors++))

# PZEM Consumer
if [ -f "PZEM_ConsumerSender/middleware/RabbitToSocketMiddleware_NEW.go" ]; then
    check_file "PZEM_ConsumerSender/middleware/RabbitToSocketMiddleware_NEW.go"
else
    check_file "PZEM_ConsumerSender/middleware/RabbitToSocketMiddleware.go" || ((errors++))
fi

# DHT22 Consumer
if [ -f "DHT22_ConsumerSender/middleware/RabbitToSocketMiddleware_NEW.go" ]; then
    check_file "DHT22_ConsumerSender/middleware/RabbitToSocketMiddleware_NEW.go"
else
    check_file "DHT22_ConsumerSender/middleware/RabbitToSocketMiddleware.go" || ((errors++))
fi

# PIR Consumer
check_file "PIR_ConsumerSender/middleware/RabbitToSocketMiddleware.go" || ((errors++))

# Light Consumer
check_file "LightSensor_ConsumerSender/middleware/RabbitToSocketMiddleware.go" || ((errors++))

# Notification Consumer
check_file "Notification_ConsumerSender/middleware/RabbitToSocketMiddleware.go" || ((errors++))

# Automation Engine
check_file "automation-engine/main.go" || ((errors++))

echo ""
echo "🐳 Verificando archivos Docker..."

# Docker files
check_file "Dockerfile" || ((errors++))
check_file "docker-compose.yml" || ((errors++))

echo ""
echo "==============================================="

if [ $errors -eq 0 ]; then
    echo "✅ ¡Todos los archivos están presentes!"
    echo "🚀 Puedes proceder con el build de Docker."
    echo ""
    echo "💡 Comandos para continuar:"
    echo "   docker-compose build"
    echo "   docker-compose up -d"
    exit 0
else
    echo "❌ Se encontraron $errors errores."
    echo "🔧 Por favor, verifica los archivos faltantes antes de continuar."
    exit 1
fi
