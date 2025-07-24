#!/bin/bash

# ==========================================
# SCRIPT DE GESTIÓN CONSUMERS VOLTIO
# ==========================================
# Permite ejecutar consumers individuales o todos juntos

show_help() {
    echo "🚀 Gestión de Consumers Voltio"
    echo ""
    echo "Uso: $0 <comando> [consumer]"
    echo ""
    echo "📋 Comandos:"
    echo "  all                    - Ejecutar todos los consumers juntos"
    echo "  single <consumer>      - Ejecutar un consumer individual"
    echo "  stop                   - Detener todos los consumers"
    echo "  logs [consumer]        - Ver logs (todos o específico)"
    echo "  status                 - Ver estado de containers"
    echo "  clean                  - Limpiar containers y volúmenes"
    echo ""
    echo "📦 Consumers disponibles:"
    echo "  pzem          - Medidores eléctricos PZEM"
    echo "  dht22         - Sensores temperatura/humedad DHT22"
    echo "  pir           - Sensores de movimiento PIR"
    echo "  light         - Sensores de luz"
    echo "  notification  - Gestor de notificaciones"
    echo "  automation    - Motor de automatización"
    echo ""
    echo "💡 Ejemplos:"
    echo "  $0 all                    # Ejecutar todos los consumers"
    echo "  $0 single pzem           # Solo consumer PZEM"
    echo "  $0 single dht22          # Solo consumer DHT22"
    echo "  $0 logs pzem             # Ver logs del consumer PZEM"
    echo "  $0 stop                  # Detener todo"
}

# Verificar si Docker está instalado
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo "❌ Docker no está instalado"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        echo "❌ Docker Compose no está instalado"
        exit 1
    fi
}

# Ejecutar todos los consumers
run_all() {
    echo "🚀 Iniciando todos los consumers..."
    docker-compose up -d
    echo "✅ Todos los consumers iniciados"
    docker-compose ps
}

# Ejecutar consumer individual
run_single() {
    local consumer=$1
    
    if [[ ! "$consumer" =~ ^(pzem|dht22|pir|light|notification|automation)$ ]]; then
        echo "❌ Consumer no válido: $consumer"
        echo "📦 Consumers disponibles: pzem, dht22, pir, light, notification, automation"
        exit 1
    fi
    
    echo "🚀 Iniciando consumer: $consumer"
    docker-compose -f docker-compose.single.yml --profile $consumer up -d
    echo "✅ Consumer $consumer iniciado"
    docker-compose -f docker-compose.single.yml ps
}

# Detener todos los services
stop_all() {
    echo "🛑 Deteniendo todos los services..."
    docker-compose down
    docker-compose -f docker-compose.single.yml down
    echo "✅ Todos los services detenidos"
}

# Ver logs
show_logs() {
    local consumer=$1
    
    if [ -z "$consumer" ]; then
        echo "📝 Mostrando logs de todos los services..."
        docker-compose logs -f
    else
        if [[ "$consumer" =~ ^(pzem|dht22|pir|light|notification|automation)$ ]]; then
            echo "📝 Mostrando logs del consumer: $consumer"
            docker-compose -f docker-compose.single.yml logs -f ${consumer}-consumer
        elif [ "$consumer" = "voltio-backend" ]; then
            echo "📝 Mostrando logs del backend unificado..."
            docker-compose logs -f voltio-backend
        else
            echo "❌ Service no válido: $consumer"
            exit 1
        fi
    fi
}

# Ver estado
show_status() {
    echo "📊 Estado de containers:"
    echo ""
    echo "🔄 Containers unificados:"
    docker-compose ps
    echo ""
    echo "🔄 Containers individuales:"
    docker-compose -f docker-compose.single.yml ps
}

# Limpiar todo
clean_all() {
    echo "🧹 Limpiando containers y volúmenes..."
    docker-compose down --volumes --remove-orphans
    docker-compose -f docker-compose.single.yml down --volumes --remove-orphans
    docker system prune -f
    echo "✅ Limpieza completada"
}

# Main
check_docker

case "$1" in
    "all")
        run_all
        ;;
    "single")
        if [ -z "$2" ]; then
            echo "❌ Especifica el consumer a ejecutar"
            show_help
            exit 1
        fi
        run_single "$2"
        ;;
    "stop")
        stop_all
        ;;
    "logs")
        show_logs "$2"
        ;;
    "status")
        show_status
        ;;
    "clean")
        clean_all
        ;;
    "help"|"-h"|"--help"|"")
        show_help
        ;;
    *)
        echo "❌ Comando no reconocido: $1"
        show_help
        exit 1
        ;;
esac
