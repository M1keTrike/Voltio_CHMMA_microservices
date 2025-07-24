# Implementación de Microservicios Consumidores Voltio

## ✅ Implementación Completada Siguiendo las Especificaciones

Este sistema implementa exactamente los microservicios solicitados en las instrucciones, siguiendo la plantilla del `pzem-consumer` existente.

### 📋 Microservicios Implementados

#### **Consumidores de Datos de Sensores:**

1. **`dht22-consumer`** - Procesa datos de sensores DHT22

   - Cola: `dht22-data-queue`
   - Formato JSON: `{"mac": "...", "temperature": 25.5, "humidity": 60.2}`
   - InfluxDB measurement: `dht22_metrics`

2. **`light-sensor-consumer`** - Procesa datos de sensores de luz

   - Cola: `light-sensor-data-queue`
   - Formato JSON: `{"mac": "...", "light_level": 850.5}`
   - InfluxDB measurement: `light_sensor_metrics`

3. **`pir-consumer`** - Procesa datos de sensores PIR
   - Cola: `pir-data-queue`
   - Formato JSON: `{"mac": "...", "motion_detected": true}`
   - InfluxDB measurement: `pir_sensor_metrics`

#### **Consumidor de Notificaciones:**

4. **`notification-consumer`** - Maneja alertas de timeout
   - Cola: `alerts-queue`
   - Envía alertas via webhook a la API principal
   - Endpoint: `http://api-service:8000/api/internal/notifications`

### 🔧 Arquitectura Implementada (Siguiendo las Especificaciones)

Cada **consumer de datos de sensores** implementa exactamente las **3 operaciones clave** especificadas:

```go
func (consumer *SensorConsumer) processMessage(sensorData *SensorMessage, originalBody []byte) error {
    var wg sync.WaitGroup
    wg.Add(3)

    // 1. Escribir en InfluxDB
    go func() {
        defer wg.Done()
        influxErr = consumer.writeToInfluxDB(sensorData)
    }()

    // 2. Publicar en WebSocket
    go func() {
        defer wg.Done()
        wsErr = consumer.publishToWebSocket(sensorData, originalBody)
    }()

    // 3. Actualizar mapa de timeouts
    go func() {
        defer wg.Done()
        consumer.updateTimeoutMap(sensorData.MAC)
    }()

    wg.Wait()
    return nil
}
```

### ⏰ Sistema de Timeouts Implementado

Cada consumer mantiene un **mapa interno** `map[mac]time.Time` y ejecuta una **goroutine de verificación** que:

1. **Revisa periódicamente** (cada 30 segundos) el mapa de timeouts
2. **Detecta dispositivos** que no han enviado datos en `TIMEOUT_SECONDS` (default: 300s)
3. **Publica alertas** en la cola `alerts-queue` con el formato:
   ```json
   {
     "mac": "device_mac_address",
     "error_type": "timeout",
     "message": "Sensor X has not sent data for 5m0s",
     "timestamp": "2025-01-15T10:30:00Z"
   }
   ```

### 📡 Consumer de Notificaciones (Sistema Simple)

El `notification-consumer` implementa la lógica especificada:

```go
// Por cada mensaje de alerts-queue:
1. Parsea el JSON del mensaje de alerta
2. Obtiene la URL del webhook desde API_WEBHOOK_URL
3. Realiza HTTP POST al endpoint de la API principal
4. Gestión de respuestas:
   - Status 200-299: Envía ACK (mensaje eliminado)
   - Status 4xx/5xx o error de red: NO envía ACK (mensaje reintentado)
```

### 🐳 Dockerfiles y Configuración

Cada microservicio tiene su **Dockerfile específico** siguiendo el patrón:

- `Dockerfile.dht22`
- `Dockerfile.light-sensor`
- `Dockerfile.pir-consumer`
- `Dockerfile.notification`

### 📊 Variables de Entorno por Consumer

#### **Consumers de Datos de Sensores:**

```bash
# Conexiones
RABBITMQ_URI=amqp://admin:trike@52.73.74.139:5672/
QUEUE_NAME=dht22-data-queue  # Específico para cada sensor
WEBSOCKET_URI=wss://websocketvoltio.acstree.xyz/ws?topic=dht22&emitter=true
INFLUX_URL=http://52.201.107.193:8086
INFLUX_TOKEN=lJLzxtHLHvPNgdvU9dcInGYb/qLbLxUPgrePzLd47EKCLUWBzJ+RmJkpH0f1HkmQ
INFLUX_ORG=mi-org
INFLUX_BUCKET=sensores

# Configuración específica
CONSUMER_TYPE=dht22
TOPIC_NAME=dht22
TIMEOUT_SECONDS=300        # Como especificado
ALERTS_QUEUE_NAME=alerts-queue
```

#### **Notification Consumer:**

```bash
RABBITMQ_URI=amqp://admin:trike@52.73.74.139:5672/
ALERTS_QUEUE_NAME=alerts-queue
API_WEBHOOK_URL=http://api-service:8000/api/internal/notifications  # Como especificado
API_WEBHOOK_TOKEN=optional_bearer_token
```

### 🚀 Despliegue

#### **Opción 1: Todos los servicios**

```bash
docker-compose -f docker-compose-new.yml up -d
```

#### **Opción 2: Servicios individuales**

```bash
# DHT22 Consumer
docker-compose -f docker-compose-new.yml up -d dht22-consumer

# Light Sensor Consumer
docker-compose -f docker-compose-new.yml up -d light-sensor-consumer

# PIR Consumer
docker-compose -f docker-compose-new.yml up -d pir-consumer

# Notification Consumer
docker-compose -f docker-compose-new.yml up -d notification-consumer
```

#### **Verificar estado:**

```bash
docker-compose -f docker-compose-new.yml ps
docker-compose -f docker-compose-new.yml logs -f [servicio]
```

### 📋 Colas de RabbitMQ Configuradas

| Consumer              | Cola RabbitMQ             | Descripción                              |
| --------------------- | ------------------------- | ---------------------------------------- |
| dht22-consumer        | `dht22-data-queue`        | Datos de temperatura y humedad           |
| light-sensor-consumer | `light-sensor-data-queue` | Datos de nivel de luz                    |
| pir-consumer          | `pir-data-queue`          | Eventos de detección de movimiento       |
| notification-consumer | `alerts-queue`            | Alertas de timeout de todos los sensores |

### 🎯 Measurements de InfluxDB

| Consumer              | Measurement            | Campos                    |
| --------------------- | ---------------------- | ------------------------- |
| dht22-consumer        | `dht22_metrics`        | `temperature`, `humidity` |
| light-sensor-consumer | `light_sensor_metrics` | `light_level`             |
| pir-consumer          | `pir_sensor_metrics`   | `motion_detected`         |

### ✅ Cumplimiento de Especificaciones

- ✅ **Fase 1**: Análisis de arquitectura del pzem-consumer ✓
- ✅ **Fase 2A**: dht22-consumer implementado ✓
- ✅ **Fase 2B**: light-sensor-consumer implementado ✓
- ✅ **Fase 2C**: pir-consumer implementado ✓
- ✅ **Fase 3**: notification-consumer implementado ✓
- ✅ **Fase 4**: docker-compose.yml actualizado ✓
- ✅ **Variables de entorno**: Configuración completa ✓
- ✅ **Gestión de timeouts**: Implementada con goroutines ✓
- ✅ **Sistema de alertas**: Webhooks con gestión de ACK/NACK ✓

### 📞 Listo para Producción

El sistema está completamente implementado siguiendo las especificaciones exactas. Cada microservicio es independiente, escalable y sigue el patrón establecido del `pzem-consumer` original.
