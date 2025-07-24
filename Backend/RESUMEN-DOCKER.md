# 🐳 RESUMEN: Docker Backend Voltio Unificado

## ✅ Archivos Creados

He creado un sistema Docker completo que incluye **TODOS** los servicios del backend en un solo contenedor:

### 📄 Archivos principales:
- **`Dockerfile`** - Contiene todos los servicios del backend
- **`docker-compose.yml`** - Configuración completa con dependencias
- **`init-db.sql`** - Inicialización automática de PostgreSQL
- **`start-voltio.sh`** - Script de inicio automático
- **`README-DOCKER.md`** - Documentación completa
- **`verify-files.sh`** - Verificación de archivos
- **`.dockerignore`** - Optimización del build

## 🚀 Servicios Incluidos en el Docker

| Servicio | Función |
|----------|---------|
| **WebSocket Server** | Comunicación en tiempo real (Puerto 8081) |
| **PZEM Consumer** | Procesa datos de medidores eléctricos |
| **DHT22 Consumer** | Procesa datos de temperatura/humedad |
| **PIR Consumer** | Procesa datos de sensores de movimiento |
| **Light Consumer** | Procesa datos de sensores de luz |
| **Notification Consumer** | Maneja alertas y notificaciones |
| **Automation Engine** | Motor de reglas de automatización |

## 📦 Servicios de Infraestructura

El `docker-compose.yml` incluye:
- **RabbitMQ** (con Management UI)
- **InfluxDB** (con configuración automática)
- **Conexión a PostgreSQL Externa** (tu base de datos en 13.222.89.227)

## 🔧 Cómo usar con FileZilla

### 1. Subir archivos
Sube toda la carpeta `Backend/` a tu servidor usando FileZilla.

### 2. Conectar por SSH y ejecutar
```bash
# Dar permisos
chmod +x start-voltio.sh verify-files.sh

# Verificar archivos (opcional)
./verify-files.sh

# Iniciar sistema completo
./start-voltio.sh
```

### 3. Verificar funcionamiento
```bash
# Ver estado
docker-compose ps

# Ver logs
docker-compose logs -f voltio-backend
```

## 🌐 URLs de Acceso

Una vez iniciado, tendrás acceso a:
- **WebSocket**: `ws://tu-servidor:8081/ws`
- **RabbitMQ Management**: `http://tu-servidor:15672` (guest/guest)
- **InfluxDB**: `http://tu-servidor:8086` (admin/adminpassword)
- **PostgreSQL Externa**: `13.222.89.227:5432` (chmma/HSQCx3Ajt4p^aJGC)

## 🎯 Beneficios de este enfoque

✅ **Un solo contenedor** para todo el backend
✅ **Gestión simplificada** con supervisor
✅ **Configuración automática** de dependencias
✅ **Reinicio automático** de servicios
✅ **Logs centralizados**
✅ **Escalabilidad fácil**

## 📋 Comandos útiles

```bash
# Iniciar todo
docker-compose up -d

# Ver estado
docker-compose ps

# Ver logs específicos
docker-compose logs -f voltio-backend
docker-compose logs -f rabbitmq
docker-compose logs -f influxdb

# Reiniciar servicios
docker-compose restart voltio-backend

# Detener todo
docker-compose down

# Limpiar y reiniciar
docker-compose down --volumes
docker-compose up -d --build
```

## 🔧 Personalización

Para modificar configuraciones, edita las variables de entorno en `docker-compose.yml`:

```yaml
environment:
  - RABBITMQ_URI=amqp://tu_usuario:tu_password@rabbitmq:5672/
  - INFLUXDB_URL=http://influxdb:8086
  - INFLUXDB_TOKEN=tu_token
  # ... otras configuraciones
```

## ⚡ Listo para Producción

El sistema está configurado para:
- ✅ Reinicio automático de contenedores
- ✅ Persistencia de datos en volúmenes
- ✅ Health checks para servicios críticos
- ✅ Logs estructurados y accesibles
- ✅ Red interna segura entre servicios

---

**¡Ya tienes todo listo para subir con FileZilla y ejecutar! 🎉**

El sistema completo se levantará automáticamente y estará listo para recibir datos de tus dispositivos IoT.
