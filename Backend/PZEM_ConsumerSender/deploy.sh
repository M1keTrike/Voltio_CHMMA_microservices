#!/bin/bash

# deploy.sh - Script de despliegue para consumidores Voltio

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Función para logging
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

# Función para mostrar ayuda
show_help() {
    echo "Script de despliegue para consumidores Voltio"
    echo ""
    echo "Uso: $0 [OPCIÓN]"
    echo ""
    echo "Opciones:"
    echo "  build         Construir todas las imágenes Docker"
    echo "  start         Iniciar todos los consumidores"
    echo "  stop          Parar todos los consumidores"
    echo "  restart       Reiniciar todos los consumidores"
    echo "  status        Mostrar estado de los contenedores"
    echo "  logs [SERVICIO]  Mostrar logs (opcional: servicio específico)"
    echo "  scale SERVICIO N  Escalar servicio a N instancias"
    echo "  clean         Limpiar contenedores e imágenes no usadas"
    echo "  help          Mostrar esta ayuda"
    echo ""
    echo "Servicios disponibles:"
    echo "  pzem-consumer, pir-consumer, dht-consumer, ldr-consumer"
    echo "  ir-in-consumer, ir-out-consumer, alerts-consumer"
}

# Función para construir imágenes
build_images() {
    log "Construyendo imágenes Docker..."
    
    docker-compose build --no-cache
    
    if [ $? -eq 0 ]; then
        log "Imágenes construidas exitosamente"
    else
        error "Error al construir imágenes"
        exit 1
    fi
}

# Función para iniciar servicios
start_services() {
    log "Iniciando servicios..."
    
    docker-compose up -d
    
    if [ $? -eq 0 ]; then
        log "Servicios iniciados exitosamente"
        sleep 5
        show_status
    else
        error "Error al iniciar servicios"
        exit 1
    fi
}

# Función para parar servicios
stop_services() {
    log "Parando servicios..."
    
    docker-compose down
    
    if [ $? -eq 0 ]; then
        log "Servicios parados exitosamente"
    else
        error "Error al parar servicios"
        exit 1
    fi
}

# Función para reiniciar servicios
restart_services() {
    log "Reiniciando servicios..."
    stop_services
    sleep 2
    start_services
}

# Función para mostrar estado
show_status() {
    log "Estado de los contenedores:"
    docker-compose ps
    echo ""
    log "Uso de recursos:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}" $(docker-compose ps -q) 2>/dev/null || true
}

# Función para mostrar logs
show_logs() {
    if [ -n "$1" ]; then
        log "Mostrando logs de $1..."
        docker-compose logs -f "$1"
    else
        log "Mostrando logs de todos los servicios..."
        docker-compose logs -f
    fi
}

# Función para escalar servicios
scale_service() {
    if [ -z "$1" ] || [ -z "$2" ]; then
        error "Uso: $0 scale SERVICIO NUMERO"
        exit 1
    fi
    
    log "Escalando $1 a $2 instancias..."
    docker-compose up -d --scale "$1=$2"
    
    if [ $? -eq 0 ]; then
        log "Servicio $1 escalado exitosamente"
        show_status
    else
        error "Error al escalar servicio $1"
        exit 1
    fi
}

# Función para limpiar sistema
clean_system() {
    warn "Esta operación eliminará contenedores e imágenes no utilizadas"
    read -p "¿Continuar? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log "Limpiando sistema..."
        
        # Parar servicios
        docker-compose down
        
        # Limpiar contenedores parados
        docker container prune -f
        
        # Limpiar imágenes no usadas
        docker image prune -f
        
        # Limpiar volúmenes no usados
        docker volume prune -f
        
        # Limpiar redes no usadas
        docker network prune -f
        
        log "Sistema limpiado exitosamente"
    else
        log "Operación cancelada"
    fi
}

# Verificar que docker-compose esté disponible
if ! command -v docker-compose &> /dev/null; then
    error "docker-compose no está instalado o no está en PATH"
    exit 1
fi

# Verificar que Docker esté ejecutándose
if ! docker info &> /dev/null; then
    error "Docker no está ejecutándose"
    exit 1
fi

# Procesar argumentos
case "${1:-help}" in
    build)
        build_images
        ;;
    start)
        start_services
        ;;
    stop)
        stop_services
        ;;
    restart)
        restart_services
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs "$2"
        ;;
    scale)
        scale_service "$2" "$3"
        ;;
    clean)
        clean_system
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        error "Opción no reconocida: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
