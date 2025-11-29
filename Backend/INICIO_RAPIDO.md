# 🚀 INICIO RÁPIDO - Voltio Backend Microservicios

## ⚠️ Pre-requisitos

**IMPORTANTE**: Este Docker Compose asume que ya tienes corriendo:

✅ **PostgreSQL** (puerto 5432) - Contenedor `postgres-local`
✅ **InfluxDB** (puerto 8086) - Contenedor `voltio-influxdb`  
✅ **RabbitMQ** (puertos 5672, 15672) - Contenedor `voltio-rabbitmq`
✅ **API Voltio** (puerto 8000) - Contenedor `voltio-api`

**Estos servicios NO se incluyen en este docker-compose.** Solo dockeriza los microservicios Go del Backend.

---

## 📦 Lo que se va a Dockerizar

Este `docker-compose.local.yml` **SOLO** levanta:

1. ✅ **WebSocket Server** (puerto 8081)
2. ✅ **Automation Engine** (sin puerto expuesto)
3. ✅ **PIR Consumer** (sin puerto expuesto)
4. ✅ **DHT22 Consumer** (sin puerto expuesto)
5. ✅ **Light Consumer** (sin puerto expuesto)
6. ✅ **PZEM Consumer** (sin puerto expuesto)
7. ✅ **Notification Consumer** (sin puerto expuesto)

---

## 🎯 Pasos para Iniciar

### 1️⃣ Verificar que tus servicios estén corriendo

```powershell
# Ver contenedores corriendo
docker ps

# Deberías ver:
# - postgres-local (puerto 5432)
# - voltio-influxdb (puerto 8086)
# - voltio-rabbitmq (puertos 5672, 15672)
# - voltio-api (puerto 8000)
```

### 2️⃣ Preparar tabla en PostgreSQL

Los microservicios Go necesitan la tabla `automation_rules` en tu PostgreSQL:

```powershell
# Conectar a tu PostgreSQL
docker exec -it postgres-local psql -U mike -d voltio_db

# Ejecutar el SQL (copiar y pegar):
```

```sql
CREATE TABLE IF NOT EXISTS automation_rules (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    trigger_device_mac VARCHAR(17) NOT NULL,
    action_device_mac VARCHAR(17) NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    trigger_metric VARCHAR(50) NOT NULL,
    comparison_operator VARCHAR(20),
    threshold_value DECIMAL(10, 2),
    action_capability_id INTEGER NOT NULL,
    action_payload VARCHAR(10) NOT NULL,
    active_time_start TIME,
    active_time_end TIME,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_automation_rules_trigger_mac
    ON automation_rules(trigger_device_mac);

CREATE INDEX IF NOT EXISTS idx_automation_rules_is_active
    ON automation_rules(is_active);
```

```powershell
# Salir de psql
\q
```

### 3️⃣ Configurar variables de entorno (opcional)

El archivo `.env.example` ya tiene las credenciales correctas. Si necesitas cambiar algo:

```powershell
# Copiar archivo de ejemplo
cp .env.example .env

# Editar si es necesario
notepad .env
```

**NOTA**: Las credenciales por defecto ya coinciden con tus servicios:

- PostgreSQL: `mike / trike / voltio_db`
- InfluxDB: `my-super-secret-auth-token / mi-org / sensores`
- RabbitMQ: `admin / trike`

### 4️⃣ Levantar los microservicios

```powershell
cd Backend

# Iniciar todos los microservicios
docker-compose -f docker-compose.local.yml up -d --build

# Ver logs en tiempo real
docker-compose -f docker-compose.local.yml logs -f

# Ver solo logs de un servicio
docker-compose -f docker-compose.local.yml logs -f automation-engine
```

### 5️⃣ Verificar que todo esté corriendo

```powershell
# Ver estado de contenedores
docker-compose -f docker-compose.local.yml ps

# Deberías ver 7 contenedores corriendo:
# - voltio-websocket (puerto 8081)
# - voltio-automation
# - voltio-pir-consumer
# - voltio-dht22-consumer
# - voltio-light-consumer
# - voltio-pzem-consumer
# - voltio-notification
```

### 6️⃣ Verificar conectividad

```powershell
# WebSocket Server
curl http://localhost:8081

# RabbitMQ Management (debería seguir funcionando)
# Abrir en navegador: http://localhost:15672

# InfluxDB (debería seguir funcionando)
# Abrir en navegador: http://localhost:8086
```

---

## 🧪 Pruebas

### Opción 1: Usar productores de prueba

```powershell
cd test_producers

# Iniciar productor PIR
go run pir_producer.go

# Iniciar productor DHT22
go run dht22_producer.go

# Iniciar todos
.\start_all_producers.ps1
```

### Opción 2: Verificar en RabbitMQ UI

1. Abrir http://localhost:15672 (admin/trike)
2. Ir a "Queues"
3. Deberías ver las colas creadas por los consumers
4. Publica un mensaje de prueba manualmente

---

## 🛑 Detener los Microservicios

```powershell
# Detener sin borrar contenedores
docker-compose -f docker-compose.local.yml stop

# Detener y borrar contenedores (NO borra tus datos de PostgreSQL/InfluxDB)
docker-compose -f docker-compose.local.yml down

# Ver qué queda corriendo (deberían quedar tus 4 servicios originales)
docker ps
```

---

## 🔧 Comandos Útiles

```powershell
# Reiniciar un servicio específico
docker-compose -f docker-compose.local.yml restart automation-engine

# Reconstruir un servicio
docker-compose -f docker-compose.local.yml up -d --build websocket-server

# Ver logs de todos los servicios
docker-compose -f docker-compose.local.yml logs -f

# Ver estadísticas de recursos
docker stats
```

---

## 📊 Arquitectura de Conexiones

```
┌──────────────────────────────────────────────────┐
│       SERVICIOS EXISTENTES (Ya corriendo)        │
├──────────────────────────────────────────────────┤
│  postgres-local (5432)                           │
│  voltio-influxdb (8086)                          │
│  voltio-rabbitmq (5672, 15672)                   │
│  voltio-api (8000)                               │
└────────┬─────────────────────────────────────────┘
         │
         │ host.docker.internal
         │
┌────────▼─────────────────────────────────────────┐
│    MICROSERVICIOS GO (Nuevo - En Docker)         │
├──────────────────────────────────────────────────┤
│  websocket-server (8081)                         │
│  automation-engine                               │
│  pir-consumer                                    │
│  dht22-consumer                                  │
│  light-consumer                                  │
│  pzem-consumer                                   │
│  notification-consumer                           │
└──────────────────────────────────────────────────┘
```

---

## ⚠️ Troubleshooting

### Problema: Consumer no conecta a RabbitMQ

**Solución**:

```powershell
# Verificar que RabbitMQ esté corriendo
docker ps | findstr rabbitmq

# Verificar logs
docker-compose -f docker-compose.local.yml logs pir-consumer
```

### Problema: Automation Engine no conecta a PostgreSQL

**Causa**: Tabla `automation_rules` no existe

**Solución**: Ejecutar el SQL del paso 2️⃣

### Problema: Consumers no escriben a InfluxDB

**Causa**: Token incorrecto

**Solución**:

1. Verificar token en InfluxDB UI: http://localhost:8086
2. Actualizar `.env` con el token correcto
3. Reiniciar: `docker-compose -f docker-compose.local.yml restart`

---

## ✅ Checklist de Verificación

Después de levantar todo:

- [ ] PostgreSQL está corriendo (docker ps)
- [ ] InfluxDB está corriendo (docker ps)
- [ ] RabbitMQ está corriendo (docker ps)
- [ ] API está corriendo (docker ps)
- [ ] Tabla `automation_rules` existe en PostgreSQL
- [ ] 7 microservicios Go están corriendo
- [ ] No hay errores en los logs
- [ ] WebSocket responde en http://localhost:8081
- [ ] RabbitMQ UI accesible en http://localhost:15672

---

## 🎯 Resumen

Este setup:

✅ **Reutiliza** tus servicios existentes (PostgreSQL, InfluxDB, RabbitMQ, API)
✅ **Solo dockeriza** los microservicios Go del Backend
✅ **Conecta** a servicios externos usando `host.docker.internal`
✅ **Mantiene** tus credenciales actuales
✅ **No duplica** servicios
✅ **Fácil de levantar** con un solo comando

**¡Listo para desarrollo!** 🚀
