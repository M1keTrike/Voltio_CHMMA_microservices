# 🚀 Backend Voltio - Docker Unificado

## 📋 Descripción

Este sistema Docker contiene **todos los servicios del backend** del sistema Voltio con flexibilidad para ejecutar consumers individuales o todos juntos. Se ha simplificado removiendo WebSocket Server y Test Producers para optimizar el despliegue.

### 🏗️ Servicios Incluidos

| Servicio | Descripción | Base de Datos |
|----------|-------------|--------------|
| **PZEM Consumer** | Procesa datos de medidores eléctricos | InfluxDB |
| **DHT22 Consumer** | Procesa datos de temperatura/humedad | InfluxDB |
| **PIR Consumer** | Procesa datos de sensores de movimiento | InfluxDB |
| **Light Consumer** | Procesa datos de sensores de luz | InfluxDB |
| **Notification Consumer** | Maneja alertas y notificaciones | PostgreSQL |
| **Automation Engine** | Motor de reglas de automatización | PostgreSQL |

### 🔗 Infraestructura
- **RabbitMQ**: Cola de mensajes (Container)
- **InfluxDB**: Base de datos de series temporales (Container)  
- **PostgreSQL**: Base de datos relacional (Externa: 13.222.89.227:5432)

## 🚀 Inicio Rápido

### 1. Subir archivos con FileZilla

Sube toda la carpeta `Backend/` a tu servidor usando FileZilla.

### 2. Ejecutar en el servidor

```bash
# Dar permisos de ejecución
chmod +x voltio-manager.sh

# Ver opciones disponibles
./voltio-manager.sh help

# Ejecutar todos los consumers
./voltio-manager.sh all

# O ejecutar un consumer específico
./voltio-manager.sh single pzem
```

### 3. Verificar funcionamiento

```bash
# Ver estado de contenedores
./voltio-manager.sh status

# Ver logs en tiempo real
./voltio-manager.sh logs

## 🐳 Opciones de Ejecución

### Opción 1: Todos los Consumers Juntos (Recomendado)
```bash
# Ejecutar todos los consumers en un solo contenedor
./voltio-manager.sh all

# Ver logs de todos
./voltio-manager.sh logs

# Detener todos
./voltio-manager.sh stop
```

### Opción 2: Consumers Individuales
```bash
# Ejecutar solo PZEM consumer
./voltio-manager.sh single pzem

# Ejecutar solo DHT22 consumer
./voltio-manager.sh single dht22

# Ejecutar solo PIR consumer
./voltio-manager.sh single pir

# Ejecutar solo Light consumer
./voltio-manager.sh single light

# Ejecutar solo Notification consumer
./voltio-manager.sh single notification

# Ejecutar solo Automation Engine
./voltio-manager.sh single automation
```

### Ver Logs Específicos
```bash
# Logs de un consumer específico
./voltio-manager.sh logs pzem
./voltio-manager.sh logs dht22

# Logs de todos los services
./voltio-manager.sh logs
```

## 🔧 Uso Directo con Docker Compose

### Para todos los consumers
```bash
docker-compose up -d              # Iniciar todos
docker-compose down               # Detener todos
docker-compose logs -f            # Ver logs todos
```

### Para consumers individuales
```bash
# Iniciar solo PZEM
docker-compose -f docker-compose.single.yml --profile pzem up -d

# Iniciar solo DHT22
docker-compose -f docker-compose.single.yml --profile dht22 up -d

# Ver logs específicos
docker-compose -f docker-compose.single.yml logs -f pzem-consumer
```

## 🔧 Configuración

### Variables de Entorno

| Variable | Valor por Defecto | Descripción |
|----------|-------------------|-------------|
| `RABBITMQ_URI` | `amqp://guest:guest@rabbitmq:5672/` | Conexión RabbitMQ |
| `INFLUXDB_URL` | `http://influxdb:8086` | URL InfluxDB |
| `INFLUXDB_TOKEN` | `lJLzxtHLHvPNgdvU9dcInGYb/...` | Token InfluxDB |
| `INFLUXDB_ORG` | `mi-org` | Organización InfluxDB |
| `INFLUXDB_BUCKET` | `sensores` | Bucket InfluxDB |
| `POSTGRES_HOST` | `13.222.89.227` | Host PostgreSQL Externa |
| `POSTGRES_DB` | `voltiodb` | Base de datos PostgreSQL Externa |

### Personalizar configuración

Edita el archivo `docker-compose.yml` y modifica las variables de entorno:

```yaml
environment:
  - RABBITMQ_URI=amqp://tu_usuario:tu_password@tu_host:5672/
  - INFLUXDB_URL=http://tu_influxdb:8086
  # ... otras variables
```

## 📊 Monitoreo

### Acceso a servicios

- **WebSocket**: `ws://localhost:8081/ws`
- **RabbitMQ Management**: `http://localhost:15672` (guest/guest)
- **InfluxDB**: `http://localhost:8086` (admin/adminpassword)
- **PostgreSQL**: `13.222.89.227:5432` (chmma/HSQCx3Ajt4p^aJGC)

### Verificar salud de servicios

```bash
# Estado general
docker-compose ps

# Recursos utilizados
docker stats

# Logs de errores
docker-compose logs voltio-backend | grep -i error
```

## 🔍 Troubleshooting

### Problema: Servicios no se conectan

```bash
# Verificar red Docker
docker network ls
docker network inspect backend_voltio-network

# Reiniciar servicios dependientes
docker-compose restart rabbitmq influxdb postgres
docker-compose restart voltio-backend
```

### Problema: Falta memoria

```bash
# Limpiar contenedores no utilizados
docker system prune -f

# Ver uso de recursos
docker stats --no-stream
```

### Problema: Puertos ocupados

```bash
# Verificar puertos en uso
netstat -tulpn | grep :8081
netstat -tulpn | grep :5672

# Cambiar puertos en docker-compose.yml si es necesario
```

## 📁 Estructura del Proyecto

```
Backend/
├── Dockerfile                 # Docker unificado
├── docker-compose.yml         # Configuración completa
├── start-voltio.sh           # Script de inicio
├── README-DOCKER.md          # Esta documentación
├── WebSocketServer/          # Código WebSocket Server
├── PZEM_ConsumerSender/      # Consumer PZEM
├── DHT22_ConsumerSender/     # Consumer DHT22
├── PIR_ConsumerSender/       # Consumer PIR
├── LightSensor_ConsumerSender/ # Consumer Light
├── Notification_ConsumerSender/ # Consumer Notifications
└── automation-engine/        # Motor de automatización
```

## 🚨 Notas de Seguridad

1. **Cambiar passwords por defecto** en producción
2. **Usar certificados SSL/TLS** para conexiones externas
3. **Configurar firewall** para permitir solo puertos necesarios
4. **Realizar backups regulares** de datos InfluxDB

## 📞 Soporte

- Para problemas con Docker: Verificar logs con `docker-compose logs`
- Para problemas de conectividad: Verificar variables de entorno
- Para problemas de rendimiento: Verificar recursos con `docker stats`

## 🎯 Pruebas

Para probar el sistema, puedes usar los productores de prueba:

```bash
# En otra terminal/máquina
cd test_producers/
go run dht22/main.go  # Enviar datos DHT22
go run pzem/main.go   # Enviar datos PZEM
# etc...
```

---

**¡Sistema Voltio listo para producción! 🎉**
