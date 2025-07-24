# Test Producers - Productores de Prueba

Este directorio contiene productores de prueba para generar datos de todos los tipos de sensores del sistema Voltio.

## � Estructura del Proyecto

```
test_producers/
├── README.md
├── go.mod
├── go.sum
├── start_all_producers.ps1      # Script PowerShell para ejecutar todos los productores
├── start_single_producer.ps1    # Script PowerShell para ejecutar un productor individual
├── start_all_producers.bat      # Script Batch para ejecutar todos los productores
├── start_single_producer.bat    # Script Batch para ejecutar un productor individual
├── dht22/
│   └── main.go                  # Productor DHT22 (Temperatura/Humedad)
├── light/
│   └── main.go                  # Productor Light Sensor
├── pir/
│   └── main.go                  # Productor PIR Motion
└── pzem/
    └── main.go                  # Productor PZEM Electric Meter
```

## 🌡️ Productores Disponibles

### 1. DHT22 Producer (dht22/main.go)

- **Cola**: `DHT22_queue`
- **Datos**: Temperatura y humedad
- **Frecuencia**: Cada 30 segundos
- **Dispositivos**: 3 sensores simulados
- **Formato**:

```json
{
  "deviceId": "DHT22-DEV-001",
  "payload": {
    "mac": "DHT22-001",
    "temperature": 25.5,
    "humidity": 60.2
  }
}
```

### 2. Light Sensor Producer (light/main.go)

- **Cola**: `LightSensor_queue`
- **Datos**: Nivel de luz en lux
- **Frecuencia**: Cada 15 segundos
- **Dispositivos**: 4 sensores (interior y exterior)
- **Formato**:

```json
{
  "deviceId": "LIGHT-DEV-001",
  "payload": {
    "mac": "LIGHT-001",
    "lightLevel": 1250.5
  }
}
```

### 3. PIR Motion Sensor Producer (pir/main.go)

- **Cola**: `PIR_queue`
- **Datos**: Detección de movimiento (boolean)
- **Frecuencia**: Cada 20 segundos
- **Dispositivos**: 6 sensores en diferentes ubicaciones
- **Formato**:

```json
{
  "deviceId": "PIR-DEV-001",
  "payload": {
    "mac": "PIR-001",
    "motionDetected": true
  }
}
```

### 4. PZEM Electric Meter Producer (pzem/main.go)

- **Cola**: `PZEM_queue`
- **Datos**: Medidas eléctricas completas
- **Frecuencia**: Cada 10 segundos
- **Dispositivos**: 4 medidores en diferentes circuitos
- **Formato**:

```json
{
  "deviceId": "PZEM-DEV-001",
  "payload": {
    "mac": "PZEM-001",
    "voltage": 220.5,
    "current": 5.25,
    "power": 1157.6,
    "energy": 1543.2,
    "frequency": 50.1,
    "powerFactor": 0.92
  }
}
```

## ⚙️ Configuración

Todos los productores están configurados para conectarse a:

- **RabbitMQ**: `amqp://admin:trike@52.73.74.139:5672/`
- **Usuario**: `admin`
- **Password**: `trike`
- **Dependencias**: `github.com/rabbitmq/amqp091-go`

## 🚀 Instalación y Configuración

### Prerrequisitos

- Go 1.19 o superior
- Acceso a RabbitMQ en `52.73.74.139:5672`

### Instalación

```bash
cd test_producers
go mod tidy
```

## 🎮 Ejecución

### Opción 1: Scripts Batch (Recomendado para Windows)

```batch
# Ejecuta todos los productores en ventanas separadas
start_all_producers.bat

# Ejecutar un productor específico
start_single_producer.bat dht22
start_single_producer.bat light
start_single_producer.bat pir
start_single_producer.bat pzem
```

### Opción 2: Scripts PowerShell

```powershell
# Ejecuta todos los productores en ventanas separadas
.\start_all_producers.ps1

# Ejecutar un productor específico
.\start_single_producer.ps1 dht22
.\start_single_producer.ps1 light
.\start_single_producer.ps1 pir
.\start_single_producer.ps1 pzem
```

### Opción 3: Ejecución manual

```bash
# Productor DHT22
cd dht22
go run main.go

# Productor Light Sensor
cd light
go run main.go

# Productor PIR Motion
cd pir
go run main.go

# Productor PZEM Electric
cd pzem
go run main.go
```

## 📊 Datos Generados

Los productores generan datos realistas que incluyen:

### 🌡️ DHT22 (Temperatura/Humedad)

- Variaciones de temperatura según la hora del día
- Correlación inversa entre temperatura y humedad
- Rango: 20-35°C, 40-70% humedad

### 💡 Light Sensor

- Patrones solares para sensores exteriores (0-100,000 lux)
- Patrones de uso humano para interiores (0-1,000 lux)
- Variación según horarios de actividad

### 🚶 PIR Motion

- Probabilidades de movimiento basadas en ubicación
- Patrones horarios realistas (mayor actividad 6AM-10PM)
- Diferentes niveles de actividad por zona

### ⚡ PZEM Electric

- Patrones de consumo según tipo de carga
- Variaciones por horario y tipo de dispositivo
- Medidas eléctricas completas con factor de potencia

## 🎯 Propósito de Testing

Estos productores permiten probar:

1. **Sistema de Alertas por Timeout**: Deteniendo productores para activar alertas
2. **Flujo de Datos**: Validar escritura en InfluxDB y streaming WebSocket
3. **Carga del Sistema**: Probar con múltiples sensores simultáneos
4. **Alertas CRÍTICAS**: El productor PZEM puede generar condiciones para alertas eléctricas

## 🛑 Detener Productores

- **Scripts Batch/PowerShell**: Cerrar las ventanas de CMD/PowerShell correspondientes
- **Ejecución manual**: Presionar `Ctrl+C` en cada terminal

## 📝 Logs

Cada productor muestra:

- ✅ Estado de conexión a RabbitMQ
- 📤 Datos enviados con valores en tiempo real
- ❌ Errores de conexión o publicación
- 🔄 Indicador de frecuencia de envío
