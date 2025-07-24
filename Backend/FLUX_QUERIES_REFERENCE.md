# 📊 Consultas InfluxDB Avanzadas - Sistema Voltio

## 🔍 **Consultas Flux Preparadas para Copy-Paste**

### ⚡ **ENERGÍA - Medidores PZEM**

#### 1. **Consumo en tiempo real (último valor)**

```flux
from(bucket: "sensores")
  |> range(start: -5m)
  |> filter(fn: (r) => r._measurement == "energy_metrics")
  |> filter(fn: (r) => r.mac == "PZEM-001")
  |> last()
```

#### 2. **Potencia promedio por hora (últimas 24h)**

```flux
from(bucket: "sensores")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "energy_metrics")
  |> filter(fn: (r) => r._field == "power")
  |> filter(fn: (r) => r.mac == "PZEM-001")
  |> aggregateWindow(every: 1h, fn: mean, createEmpty: false)
  |> yield(name: "hourly_average_power")
```

#### 3. **Consumo total de energía por dispositivo (última semana)**

```flux
from(bucket: "sensores")
  |> range(start: -7d)
  |> filter(fn: (r) => r._measurement == "energy_metrics")
  |> filter(fn: (r) => r._field == "energy")
  |> group(columns: ["mac"])
  |> max()
  |> group()
  |> sort(columns: ["_value"], desc: true)
```

#### 4. **Detección de picos de consumo (>2000W)**

```flux
from(bucket: "sensores")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "energy_metrics")
  |> filter(fn: (r) => r._field == "power")
  |> filter(fn: (r) => r._value > 2000.0)
  |> group(columns: ["mac"])
  |> count()
```

#### 5. **Factor de potencia bajo (<0.85)**

```flux
from(bucket: "sensores")
  |> range(start: -1h)
  |> filter(fn: (r) => r._measurement == "energy_metrics")
  |> filter(fn: (r) => r._field == "powerFactor")
  |> filter(fn: (r) => r._value < 0.85)
  |> sort(columns: ["_time"], desc: true)
```

### 🌡️ **AMBIENTE - Sensores DHT22**

#### 6. **Temperatura actual de todos los sensores**

```flux
from(bucket: "sensores")
  |> range(start: -5m)
  |> filter(fn: (r) => r._measurement == "temperature_humidity_metrics")
  |> filter(fn: (r) => r._field == "temperature")
  |> last()
  |> group()
```

#### 7. **Promedio de humedad por día (último mes)**

```flux
from(bucket: "sensores")
  |> range(start: -30d)
  |> filter(fn: (r) => r._measurement == "temperature_humidity_metrics")
  |> filter(fn: (r) => r._field == "humidity")
  |> aggregateWindow(every: 1d, fn: mean, createEmpty: false)
  |> group(columns: ["mac"])
```

#### 8. **Máximas y mínimas temperaturas diarias**

```flux
// Máximas
max_temp = from(bucket: "sensores")
  |> range(start: -7d)
  |> filter(fn: (r) => r._measurement == "temperature_humidity_metrics")
  |> filter(fn: (r) => r._field == "temperature")
  |> aggregateWindow(every: 1d, fn: max, createEmpty: false)
  |> set(key: "type", value: "max")

// Mínimas
min_temp = from(bucket: "sensores")
  |> range(start: -7d)
  |> filter(fn: (r) => r._measurement == "temperature_humidity_metrics")
  |> filter(fn: (r) => r._field == "temperature")
  |> aggregateWindow(every: 1d, fn: min, createEmpty: false)
  |> set(key: "type", value: "min")

union(tables: [max_temp, min_temp])
  |> sort(columns: ["_time"])
```

#### 9. **Alertas de temperatura extrema (<15°C o >35°C)**

```flux
from(bucket: "sensores")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "temperature_humidity_metrics")
  |> filter(fn: (r) => r._field == "temperature")
  |> filter(fn: (r) => r._value < 15.0 or r._value > 35.0)
  |> sort(columns: ["_time"], desc: true)
```

### 🚶 **MOVIMIENTO - Sensores PIR**

#### 10. **Actividad por zona (últimas 2 horas)**

```flux
from(bucket: "sensores")
  |> range(start: -2h)
  |> filter(fn: (r) => r._measurement == "motion_sensor_metrics")
  |> filter(fn: (r) => r._field == "motion_detected")
  |> filter(fn: (r) => r._value == true)
  |> group(columns: ["mac"])
  |> count()
  |> group()
  |> sort(columns: ["_value"], desc: true)
```

#### 11. **Patrón de movimiento por horas del día**

```flux
from(bucket: "sensores")
  |> range(start: -7d)
  |> filter(fn: (r) => r._measurement == "motion_sensor_metrics")
  |> filter(fn: (r) => r._field == "motion_detected")
  |> filter(fn: (r) => r._value == true)
  |> map(fn: (r) => ({r with hour: uint(v: date.hour(t: r._time))}))
  |> group(columns: ["hour"])
  |> count()
  |> group()
  |> sort(columns: ["hour"])
```

#### 12. **Zonas sin actividad (últimas 4 horas)**

```flux
// Todas las MACs conocidas
all_macs = from(bucket: "sensores")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "motion_sensor_metrics")
  |> group(columns: ["mac"])
  |> distinct(column: "mac")

// MACs con actividad reciente
active_macs = from(bucket: "sensores")
  |> range(start: -4h)
  |> filter(fn: (r) => r._measurement == "motion_sensor_metrics")
  |> filter(fn: (r) => r._field == "motion_detected")
  |> filter(fn: (r) => r._value == true)
  |> group(columns: ["mac"])
  |> distinct(column: "mac")

// Diferencia (zonas inactivas)
join(
  tables: {all: all_macs, active: active_macs},
  on: ["mac"],
  method: "left"
)
|> filter(fn: (r) => not exists r._value_active)
```

### 💡 **LUZ - Sensores de Luminosidad**

#### 13. **Niveles de luz actuales todas las zonas**

```flux
from(bucket: "sensores")
  |> range(start: -5m)
  |> filter(fn: (r) => r._measurement == "light_sensor_metrics")
  |> filter(fn: (r) => r._field == "light_level")
  |> last()
  |> group()
  |> sort(columns: ["_value"], desc: true)
```

#### 14. **Patrón diario de iluminación (promedio por hora)**

```flux
from(bucket: "sensores")
  |> range(start: -7d)
  |> filter(fn: (r) => r._measurement == "light_sensor_metrics")
  |> filter(fn: (r) => r._field == "light_level")
  |> map(fn: (r) => ({r with hour: uint(v: date.hour(t: r._time))}))
  |> group(columns: ["hour", "mac"])
  |> mean()
  |> group(columns: ["hour"])
  |> mean()
  |> sort(columns: ["hour"])
```

#### 15. **Detección de luz artificial (>100 lux entre 19:00-06:00)**

```flux
from(bucket: "sensores")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "light_sensor_metrics")
  |> filter(fn: (r) => r._field == "light_level")
  |> filter(fn: (r) => r._value > 100.0)
  |> map(fn: (r) => ({r with hour: uint(v: date.hour(t: r._time))}))
  |> filter(fn: (r) => r.hour >= 19 or r.hour <= 6)
  |> sort(columns: ["_time"], desc: true)
```

---

## 📈 **Consultas de Análisis Avanzado**

### 16. **Correlación Temperatura vs Consumo Eléctrico**

```flux
// Temperatura promedio por hora
temp = from(bucket: "sensores")
  |> range(start: -7d)
  |> filter(fn: (r) => r._measurement == "temperature_humidity_metrics")
  |> filter(fn: (r) => r._field == "temperature")
  |> aggregateWindow(every: 1h, fn: mean, createEmpty: false)
  |> set(key: "_field", value: "avg_temperature")

// Consumo promedio por hora
power = from(bucket: "sensores")
  |> range(start: -7d)
  |> filter(fn: (r) => r._measurement == "energy_metrics")
  |> filter(fn: (r) => r._field == "power")
  |> aggregateWindow(every: 1h, fn: mean, createEmpty: false)
  |> set(key: "_field", value: "avg_power")

// Combinar datos
union(tables: [temp, power])
  |> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
  |> sort(columns: ["_time"])
```

### 17. **Eficiencia Energética por Zona**

```flux
from(bucket: "sensores")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "energy_metrics")
  |> filter(fn: (r) => r._field == "power" or r._field == "powerFactor")
  |> pivot(rowKey: ["_time", "mac"], columnKey: ["_field"], valueColumn: "_value")
  |> map(fn: (r) => ({r with efficiency: r.power * r.powerFactor}))
  |> group(columns: ["mac"])
  |> mean(column: "efficiency")
  |> sort(columns: ["efficiency"], desc: true)
```

### 18. **Resumen de Actividad General del Sistema**

```flux
// Conteo de eventos por tipo de sensor
energy_count = from(bucket: "sensores")
  |> range(start: -1h)
  |> filter(fn: (r) => r._measurement == "energy_metrics")
  |> count()
  |> set(key: "sensor_type", value: "energy")

environment_count = from(bucket: "sensores")
  |> range(start: -1h)
  |> filter(fn: (r) => r._measurement == "temperature_humidity_metrics")
  |> count()
  |> set(key: "sensor_type", value: "environment")

motion_count = from(bucket: "sensores")
  |> range(start: -1h)
  |> filter(fn: (r) => r._measurement == "motion_sensor_metrics")
  |> count()
  |> set(key: "sensor_type", value: "motion")

light_count = from(bucket: "sensores")
  |> range(start: -1h)
  |> filter(fn: (r) => r._measurement == "light_sensor_metrics")
  |> count()
  |> set(key: "sensor_type", value: "light")

union(tables: [energy_count, environment_count, motion_count, light_count])
  |> group(columns: ["sensor_type"])
  |> sum()
```

### 19. **Detección de Sensores Inactivos (sin datos >30 min)**

```flux
import "date"

// Tiempo límite (30 minutos atrás)
threshold = date.sub(from: now(), d: 30m)

// Últimos datos por sensor
last_data = from(bucket: "sensores")
  |> range(start: -2h)
  |> group(columns: ["_measurement", "mac"])
  |> last()
  |> filter(fn: (r) => r._time < threshold)
  |> group()
  |> sort(columns: ["_time"])
```

### 20. **Consumo Total y Costo Estimado (últimas 24h)**

```flux
// Asumiendo costo de 0.15 USD por kWh
from(bucket: "sensores")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "energy_metrics")
  |> filter(fn: (r) => r._field == "energy")
  |> group(columns: ["mac"])
  |> max()
  |> group()
  |> sum()
  |> map(fn: (r) => ({r with estimated_cost: r._value * 0.15}))
```

---

## 🛠️ **Consultas de Mantenimiento**

### 21. **Limpieza de datos antiguos (>90 días)**

```flux
// ⚠️ PRECAUCIÓN: Esta consulta ELIMINA datos
from(bucket: "sensores")
  |> range(start: -90d, stop: -89d)
  |> drop()
```

### 22. **Verificación de integridad de datos**

```flux
from(bucket: "sensores")
  |> range(start: -1h)
  |> group(columns: ["_measurement"])
  |> count()
  |> group()
  |> sort(columns: ["_measurement"])
```

Estas consultas te dan una base sólida para extraer insights valiosos del sistema IoT Voltio.
