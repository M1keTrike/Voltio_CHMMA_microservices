# deploy.ps1 - Script de despliegue para consumidores Voltio (Windows)

param(
    [Parameter(Position=0)]
    [string]$Action = "help",
    
    [Parameter(Position=1)]
    [string]$Service = "",
    
    [Parameter(Position=2)]
    [int]$Scale = 1
)

# Función para logging con colores
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    
    switch ($Level) {
        "INFO" { Write-Host "[$timestamp] $Message" -ForegroundColor Green }
        "WARN" { Write-Host "[$timestamp] WARNING: $Message" -ForegroundColor Yellow }
        "ERROR" { Write-Host "[$timestamp] ERROR: $Message" -ForegroundColor Red }
    }
}

# Función para mostrar ayuda
function Show-Help {
    Write-Host "Script de despliegue para consumidores Voltio" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Uso: .\deploy.ps1 [ACCIÓN] [PARÁMETROS]" -ForegroundColor White
    Write-Host ""
    Write-Host "Acciones:" -ForegroundColor White
    Write-Host "  build         Construir todas las imágenes Docker"
    Write-Host "  start         Iniciar todos los consumidores"
    Write-Host "  stop          Parar todos los consumidores"
    Write-Host "  restart       Reiniciar todos los consumidores"
    Write-Host "  status        Mostrar estado de los contenedores"
    Write-Host "  logs [SERVICIO]  Mostrar logs (opcional: servicio específico)"
    Write-Host "  scale SERVICIO N  Escalar servicio a N instancias"
    Write-Host "  clean         Limpiar contenedores e imágenes no usadas"
    Write-Host "  help          Mostrar esta ayuda"
    Write-Host ""
    Write-Host "Servicios disponibles:" -ForegroundColor White
    Write-Host "  pzem-consumer, pir-consumer, dht-consumer, ldr-consumer"
    Write-Host "  ir-in-consumer, ir-out-consumer, alerts-consumer"
    Write-Host ""
    Write-Host "Ejemplos:" -ForegroundColor White
    Write-Host "  .\deploy.ps1 build"
    Write-Host "  .\deploy.ps1 start"
    Write-Host "  .\deploy.ps1 logs pzem-consumer"
    Write-Host "  .\deploy.ps1 scale pzem-consumer 3"
}

# Verificar prerrequisitos
function Test-Prerequisites {
    # Verificar Docker
    try {
        docker info | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Log "Docker no está ejecutándose" "ERROR"
            exit 1
        }
    }
    catch {
        Write-Log "Docker no está instalado o no está en PATH" "ERROR"
        exit 1
    }
    
    # Verificar docker-compose
    try {
        docker-compose version | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Log "docker-compose no está disponible" "ERROR"
            exit 1
        }
    }
    catch {
        Write-Log "docker-compose no está instalado o no está en PATH" "ERROR"
        exit 1
    }
}

# Construir imágenes
function Build-Images {
    Write-Log "Construyendo imágenes Docker..."
    
    docker-compose build --no-cache
    
    if ($LASTEXITCODE -eq 0) {
        Write-Log "Imágenes construidas exitosamente"
    } else {
        Write-Log "Error al construir imágenes" "ERROR"
        exit 1
    }
}

# Iniciar servicios
function Start-Services {
    Write-Log "Iniciando servicios..."
    
    docker-compose up -d
    
    if ($LASTEXITCODE -eq 0) {
        Write-Log "Servicios iniciados exitosamente"
        Start-Sleep -Seconds 5
        Show-Status
    } else {
        Write-Log "Error al iniciar servicios" "ERROR"
        exit 1
    }
}

# Parar servicios
function Stop-Services {
    Write-Log "Parando servicios..."
    
    docker-compose down
    
    if ($LASTEXITCODE -eq 0) {
        Write-Log "Servicios parados exitosamente"
    } else {
        Write-Log "Error al parar servicios" "ERROR"
        exit 1
    }
}

# Reiniciar servicios
function Restart-Services {
    Write-Log "Reiniciando servicios..."
    Stop-Services
    Start-Sleep -Seconds 2
    Start-Services
}

# Mostrar estado
function Show-Status {
    Write-Log "Estado de los contenedores:"
    docker-compose ps
    Write-Host ""
    
    Write-Log "Uso de recursos:"
    $containers = docker-compose ps -q
    if ($containers) {
        docker stats --no-stream --format "table {{.Container}}`t{{.CPUPerc}}`t{{.MemUsage}}`t{{.NetIO}}" $containers
    }
}

# Mostrar logs
function Show-Logs {
    param([string]$ServiceName)
    
    if ($ServiceName) {
        Write-Log "Mostrando logs de $ServiceName..."
        docker-compose logs -f $ServiceName
    } else {
        Write-Log "Mostrando logs de todos los servicios..."
        docker-compose logs -f
    }
}

# Escalar servicio
function Scale-Service {
    param([string]$ServiceName, [int]$Instances)
    
    if (-not $ServiceName -or $Instances -le 0) {
        Write-Log "Uso: .\deploy.ps1 scale SERVICIO NUMERO" "ERROR"
        exit 1
    }
    
    Write-Log "Escalando $ServiceName a $Instances instancias..."
    docker-compose up -d --scale "$ServiceName=$Instances"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Log "Servicio $ServiceName escalado exitosamente"
        Show-Status
    } else {
        Write-Log "Error al escalar servicio $ServiceName" "ERROR"
        exit 1
    }
}

# Limpiar sistema
function Clean-System {
    Write-Log "Esta operación eliminará contenedores e imágenes no utilizadas" "WARN"
    $response = Read-Host "¿Continuar? (y/N)"
    
    if ($response -eq "y" -or $response -eq "Y") {
        Write-Log "Limpiando sistema..."
        
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
        
        Write-Log "Sistema limpiado exitosamente"
    } else {
        Write-Log "Operación cancelada"
    }
}

# Verificar prerrequisitos
Test-Prerequisites

# Procesar acción
switch ($Action.ToLower()) {
    "build" {
        Build-Images
    }
    "start" {
        Start-Services
    }
    "stop" {
        Stop-Services
    }
    "restart" {
        Restart-Services
    }
    "status" {
        Show-Status
    }
    "logs" {
        Show-Logs -ServiceName $Service
    }
    "scale" {
        Scale-Service -ServiceName $Service -Instances $Scale
    }
    "clean" {
        Clean-System
    }
    "help" {
        Show-Help
    }
    default {
        Write-Log "Acción no reconocida: $Action" "ERROR"
        Write-Host ""
        Show-Help
        exit 1
    }
}
