# Voltio Services - Deployment Guide

Este repositorio contiene todos los servicios de Voltio (consumers y WebSocket server) empaquetados en un único contenedor Docker para facilitar el deployment.

## 🏗️ Arquitectura

El sistema incluye:
- **WebSocket Server**: Servidor de comunicación en tiempo real (Puerto 8081)
- **PZEM Consumer**: Procesa datos de consumo eléctrico
- **DHT22 Consumer**: Procesa datos de temperatura y humedad
- **Light Sensor Consumer**: Procesa datos de sensor de luz
- **PIR Consumer**: Procesa datos de sensor de movimiento
- **Notification Consumer**: Maneja alertas del sistema

Todos los servicios se ejecutan en un único contenedor usando Supervisor para gestionar los procesos.

## 🚀 Deployment en AWS EC2 Ubuntu

### Prerrequisitos

1. **Instancia EC2** con Ubuntu 20.04 o superior
2. **Security Group** configurado para permitir:
   - Puerto 22 (SSH)
   - Puerto 8081 (WebSocket Server)
3. **Par de claves** para acceso SSH

### Paso 1: Conectar a la instancia EC2

```bash
ssh -i "tu-clave.pem" ubuntu@tu-ec2-ip
```

### Paso 2: Deployment automático

```bash
# Descargar y ejecutar el script de deployment
curl -fsSL https://raw.githubusercontent.com/M1keTrike/Voltio_CHMMA/main/deploy-aws.sh -o deploy-aws.sh
chmod +x deploy-aws.sh
./deploy-aws.sh
```

### Paso 3: Configurar variables de entorno

```bash
# Editar archivo de configuración
nano /opt/voltio/Voltio_CHMMA/.env
```

Configurar las siguientes variables según tu setup:

```env
# RabbitMQ Configuration
RABBITMQ_URI=amqp://admin:trike@tu-rabbitmq-server:5672/

# InfluxDB Configuration
INFLUX_URL=http://tu-influxdb-server:8086
INFLUX_TOKEN=tu-influx-token
INFLUX_ORG=tu-organizacion
INFLUX_BUCKET=tu-bucket

# WebSocket URIs (si usas un servidor WebSocket externo)
PZEM_WEBSOCKET_URI=wss://tu-websocket-server/ws?topic=pzem&emitter=true
DHT22_WEBSOCKET_URI=wss://tu-websocket-server/ws?topic=dht22&emitter=true
LIGHT_WEBSOCKET_URI=wss://tu-websocket-server/ws?topic=light&emitter=true
PIR_WEBSOCKET_URI=wss://tu-websocket-server/ws?topic=pir&emitter=true
```

### Paso 4: Iniciar los servicios

```bash
# Aplicar cambios en el grupo docker
newgrp docker

# Ir al directorio del proyecto
cd /opt/voltio/Voltio_CHMMA

# Iniciar servicios
sudo systemctl start voltio-services

# Verificar estado
sudo systemctl status voltio-services
```

### Paso 5: Verificar deployment

```bash
# Ver logs en tiempo real
docker-compose logs -f

# Verificar que todos los servicios estén ejecutándose
docker-compose ps

# Verificar conectividad del WebSocket
curl http://localhost:8081/health
```

## 🖥️ Desarrollo Local

### Windows (PowerShell)

```powershell
# Clonar repositorio
git clone https://github.com/M1keTrike/Voltio_CHMMA.git
cd Voltio_CHMMA

# Configurar variables de entorno
Copy-Item .env.example .env
# Editar .env con tu configuración

# Iniciar servicios
.\start-voltio.ps1

# Ver logs
.\start-voltio.ps1 -Logs

# Parar servicios
.\start-voltio.ps1 -Stop
```

### Linux/macOS

```bash
# Clonar repositorio
git clone https://github.com/M1keTrike/Voltio_CHMMA.git
cd Voltio_CHMMA

# Configurar variables de entorno
cp .env.example .env
# Editar .env con tu configuración

# Iniciar servicios
docker-compose up -d

# Ver logs
docker-compose logs -f

# Parar servicios
docker-compose down
```

## 🔧 Configuración Avanzada

### Variables de Entorno Principales

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `RABBITMQ_URI` | URI de conexión a RabbitMQ | `amqp://user:pass@host:5672/` |
| `INFLUX_URL` | URL del servidor InfluxDB | `http://influxdb:8086` |
| `INFLUX_TOKEN` | Token de autenticación InfluxDB | `your-token-here` |
| `WEBSOCKET_PORT` | Puerto del servidor WebSocket | `8081` |

### Personalizar Configuración Supervisor

El archivo de configuración de Supervisor se genera automáticamente, pero puedes modificarlo editando el Dockerfile en la sección de configuración.

### Logs y Monitoreo

```bash
# Ver logs de un servicio específico
docker-compose exec voltio-services supervisorctl tail -f pzem-consumer

# Ver todos los procesos
docker-compose exec voltio-services supervisorctl status

# Reiniciar un servicio específico
docker-compose exec voltio-services supervisorctl restart pzem-consumer
```

## 🚨 Troubleshooting

### Servicios no inician

1. Verificar logs: `docker-compose logs`
2. Verificar configuración: `cat .env`
3. Verificar conectividad a RabbitMQ e InfluxDB

### Error de conectividad

1. Verificar Security Groups en AWS
2. Verificar variables de entorno
3. Verificar que RabbitMQ e InfluxDB estén accesibles

### Problemas de memoria

```bash
# Verificar uso de recursos
docker stats

# Si es necesario, aumentar el tamaño de la instancia EC2
```

## 📊 Monitoreo

### Health Checks

El contenedor incluye health checks automáticos:

```bash
# Verificar salud del contenedor
docker-compose ps

# El estado debería mostrar "healthy"
```

### Logs Estructurados

Los logs se guardan en el directorio `logs/` y se organizan por servicio:

- `logs/pzem-consumer.out.log`
- `logs/dht22-consumer.out.log`
- `logs/websocket-server.out.log`
- etc.

## 🔄 Actualizaciones

```bash
# Actualizar código
cd /opt/voltio/Voltio_CHMMA
git pull

# Reconstruir y reiniciar
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## 🛠️ Comandos Útiles

```bash
# Ver estado de todos los servicios
docker-compose ps

# Seguir logs en tiempo real
docker-compose logs -f

# Reiniciar todos los servicios
docker-compose restart

# Escalar servicios (si necesario)
docker-compose up -d --scale voltio-services=2

# Limpiar recursos no utilizados
docker system prune -f
```

## 📞 Soporte

Para reportar problemas o solicitar ayuda:

1. Revisar los logs: `docker-compose logs`
2. Verificar la configuración: `cat .env`
3. Crear un issue en el repositorio con la información relevante

## 📝 Notas de Seguridad

- Cambiar las credenciales por defecto
- Usar HTTPS/WSS en producción
- Configurar firewall apropiadamente
- Mantener Docker actualizado
- Usar secrets management para tokens sensibles
