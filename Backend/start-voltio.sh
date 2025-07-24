#!/bin/bash

# ==========================================
# SCRIPT DE INICIO SISTEMA VOLTIO - DOCKER
# ==========================================
# Para uso con FileZilla y servidores remotos

echo "🚀 Iniciando Sistema Voltio - Backend Completo..."

# Verificar si Docker está instalado
if ! command -v docker &> /dev/null; then
    echo "❌ Docker no está instalado. Instalando Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    sudo usermod -aG docker $USER
    echo "✅ Docker instalado. Por favor, reinicia la sesión y ejecuta el script nuevamente."
    exit 1
fi

# Verificar si Docker Compose está instalado
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose no está instalado. Instalando..."
    sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    echo "✅ Docker Compose instalado."
fi

# Detener y limpiar contenedores existentes si existen
echo "🧹 Limpiando contenedores existentes..."
docker-compose down --remove-orphans --volumes 2>/dev/null || true

# Limpiar imágenes antiguas
echo "🗑️ Eliminando imágenes antiguas..."
docker system prune -f

# Construir y levantar todos los servicios
echo "🔨 Construyendo imagen del backend..."
docker-compose build --no-cache

echo "🚀 Iniciando todos los servicios..."
docker-compose up -d

# Esperar a que los servicios estén listos
echo "⏳ Esperando a que los servicios estén listos..."
sleep 30

# Verificar estado de los servicios
echo "📊 Estado de los servicios:"
docker-compose ps

# Mostrar logs de inicio
echo "📝 Logs de inicio del backend:"
docker-compose logs voltio-backend --tail=20

# Información de conexión
echo ""
echo "✅ ¡Sistema Voltio iniciado correctamente!"
echo ""
echo "🔗 URLs de acceso:"
echo "   - WebSocket Server: ws://localhost:8081/ws"
echo "   - RabbitMQ Management: http://localhost:15672 (guest/guest)"
echo "   - InfluxDB: http://localhost:8086 (admin/adminpassword)"
echo "   - PostgreSQL Externa: 13.222.89.227:5432 (chmma/HSQCx3Ajt4p^aJGC)"
echo ""
echo "📋 Comandos útiles:"
echo "   - Ver logs: docker-compose logs -f"
echo "   - Detener: docker-compose down"
echo "   - Reiniciar: docker-compose restart"
echo "   - Estado: docker-compose ps"
echo ""
echo "🎯 El sistema está listo para recibir datos de los productores!"
echo "   Puedes conectar tus dispositivos IoT o ejecutar los test producers."
