# 🔌 Ejemplos de Implementación - Endpoints Backend Sistema Voltio

## 🚀 **Implementación Go (Gin Framework)**

### **Estructura del Proyecto:**

```
api/
├── main.go
├── handlers/
│   ├── energy.go
│   ├── environment.go
│   ├── motion.go
│   └── light.go
├── models/
│   └── responses.go
├── services/
│   └── influxdb.go
└── middleware/
    └── cors.go
```

### **main.go**

```go
package main

import (
    "github.com/gin-gonic/gin"
    "api/handlers"
    "api/services"
    "api/middleware"
)

func main() {
    // Inicializar InfluxDB
    influxService := services.NewInfluxDBService()
    defer influxService.Close()

    r := gin.Default()
    r.Use(middleware.CORS())

    // Grupos de rutas
    v1 := r.Group("/api/v1")
    {
        energy := v1.Group("/energy")
        {
            energy.GET("/current", handlers.GetCurrentEnergy(influxService))
            energy.GET("/history", handlers.GetEnergyHistory(influxService))
            energy.GET("/consumption", handlers.GetEnergyConsumption(influxService))
            energy.GET("/devices", handlers.GetEnergyDevices(influxService))
        }

        environment := v1.Group("/environment")
        {
            environment.GET("/current", handlers.GetCurrentEnvironment(influxService))
            environment.GET("/temperature", handlers.GetTemperatureHistory(influxService))
            environment.GET("/humidity", handlers.GetHumidityHistory(influxService))
            environment.GET("/averages", handlers.GetEnvironmentAverages(influxService))
        }

        motion := v1.Group("/motion")
        {
            motion.GET("/current", handlers.GetCurrentMotion(influxService))
            motion.GET("/events", handlers.GetMotionEvents(influxService))
            motion.GET("/activity", handlers.GetMotionActivity(influxService))
            motion.GET("/zones", handlers.GetAllMotionZones(influxService))
        }

        light := v1.Group("/light")
        {
            light.GET("/current", handlers.GetCurrentLight(influxService))
            light.GET("/history", handlers.GetLightHistory(influxService))
            light.GET("/averages", handlers.GetLightAverages(influxService))
            light.GET("/zones", handlers.GetAllLightZones(influxService))
        }
    }

    r.Run(":8080")
}
```

### **services/influxdb.go**

```go
package services

import (
    "context"
    "fmt"
    "time"

    influxdb2 "github.com/influxdata/influxdb-client-go/v2"
    "github.com/influxdata/influxdb-client-go/v2/api"
)

type InfluxDBService struct {
    client   influxdb2.Client
    queryAPI api.QueryAPI
}

func NewInfluxDBService() *InfluxDBService {
    client := influxdb2.NewClient(
        "http://52.201.107.193:8086",
        "lJLzxtHLHvPNgdvU9dcInGYb/qLbLxUPgrePzLd47EKCLUWBzJ+RmJkpH0f1HkmQ",
    )

    return &InfluxDBService{
        client:   client,
        queryAPI: client.QueryAPI("mi-org"),
    }
}

func (s *InfluxDBService) Close() {
    s.client.Close()
}

// Ejemplo: Obtener último dato de energía
func (s *InfluxDBService) GetLatestEnergyData(mac string) (map[string]interface{}, error) {
    query := fmt.Sprintf(`
        from(bucket: "sensores")
        |> range(start: -1h)
        |> filter(fn: (r) => r._measurement == "energy_metrics")
        |> filter(fn: (r) => r.mac == "%s")
        |> last()
    `, mac)

    result, err := s.queryAPI.Query(context.Background(), query)
    if err != nil {
        return nil, err
    }

    data := make(map[string]interface{})

    for result.Next() {
        record := result.Record()
        data[record.Field()] = record.Value()
        data["mac"] = record.ValueByKey("mac")
        data["deviceId"] = record.ValueByKey("deviceId")
        data["timestamp"] = record.Time()
    }

    return data, nil
}

// Ejemplo: Histórico de temperatura
func (s *InfluxDBService) GetTemperatureHistory(mac string, from, to time.Time) ([]map[string]interface{}, error) {
    query := fmt.Sprintf(`
        from(bucket: "sensores")
        |> range(start: %s, stop: %s)
        |> filter(fn: (r) => r._measurement == "temperature_humidity_metrics")
        |> filter(fn: (r) => r._field == "temperature")
        |> filter(fn: (r) => r.mac == "%s")
        |> sort(columns: ["_time"])
    `, from.Format(time.RFC3339), to.Format(time.RFC3339), mac)

    result, err := s.queryAPI.Query(context.Background(), query)
    if err != nil {
        return nil, err
    }

    var data []map[string]interface{}

    for result.Next() {
        record := result.Record()
        point := map[string]interface{}{
            "mac":         record.ValueByKey("mac"),
            "deviceId":    record.ValueByKey("deviceId"),
            "temperature": record.Value(),
            "timestamp":   record.Time(),
        }
        data = append(data, point)
    }

    return data, nil
}
```

### **handlers/energy.go**

```go
package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "api/services"
)

func GetCurrentEnergy(influxService *services.InfluxDBService) gin.HandlerFunc {
    return func(c *gin.Context) {
        mac := c.Query("mac")
        if mac == "" {
            c.JSON(http.StatusBadRequest, gin.H{
                "status": "error",
                "message": "MAC parameter is required",
            })
            return
        }

        data, err := influxService.GetLatestEnergyData(mac)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "status": "error",
                "message": err.Error(),
            })
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "status": "success",
            "data": data,
        })
    }
}

func GetEnergyHistory(influxService *services.InfluxDBService) gin.HandlerFunc {
    return func(c *gin.Context) {
        mac := c.Query("mac")
        fromStr := c.Query("from")
        toStr := c.Query("to")

        if mac == "" || fromStr == "" || toStr == "" {
            c.JSON(http.StatusBadRequest, gin.H{
                "status": "error",
                "message": "mac, from, and to parameters are required",
            })
            return
        }

        from, err := time.Parse(time.RFC3339, fromStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "status": "error",
                "message": "Invalid 'from' timestamp format. Use RFC3339",
            })
            return
        }

        to, err := time.Parse(time.RFC3339, toStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "status": "error",
                "message": "Invalid 'to' timestamp format. Use RFC3339",
            })
            return
        }

        data, err := influxService.GetEnergyHistory(mac, from, to)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "status": "error",
                "message": err.Error(),
            })
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "status": "success",
            "data": data,
            "count": len(data),
        })
    }
}
```

---

## 🐍 **Implementación Python (FastAPI)**

### **main.py**

```python
from fastapi import FastAPI, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware
from influxdb_client import InfluxDBClient
from datetime import datetime
from typing import Optional, List
import json

app = FastAPI(title="Voltio IoT API", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# InfluxDB Configuration
INFLUX_URL = "http://52.201.107.193:8086"
INFLUX_TOKEN = "lJLzxtHLHvPNgdvU9dcInGYb/qLbLxUPgrePzLd47EKCLUWBzJ+RmJkpH0f1HkmQ"
INFLUX_ORG = "mi-org"
INFLUX_BUCKET = "sensores"

client = InfluxDBClient(url=INFLUX_URL, token=INFLUX_TOKEN, org=INFLUX_ORG)
query_api = client.query_api()

@app.get("/api/v1/energy/current")
async def get_current_energy(mac: str = Query(..., description="MAC address of device")):
    """Obtener datos actuales de energía por MAC"""
    query = f'''
    from(bucket: "{INFLUX_BUCKET}")
        |> range(start: -1h)
        |> filter(fn: (r) => r._measurement == "energy_metrics")
        |> filter(fn: (r) => r.mac == "{mac}")
        |> last()
    '''

    try:
        result = query_api.query(query)
        data = {}

        for table in result:
            for record in table.records:
                data[record.get_field()] = record.get_value()
                data["mac"] = record.values.get("mac")
                data["deviceId"] = record.values.get("deviceId")
                data["timestamp"] = record.get_time().isoformat()

        if not data:
            raise HTTPException(status_code=404, detail="No data found for the specified MAC")

        return {"status": "success", "data": data}

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/v1/environment/current")
async def get_current_environment(mac: str = Query(..., description="MAC address of device")):
    """Obtener datos ambientales actuales por MAC"""
    query = f'''
    from(bucket: "{INFLUX_BUCKET}")
        |> range(start: -1h)
        |> filter(fn: (r) => r._measurement == "temperature_humidity_metrics")
        |> filter(fn: (r) => r.mac == "{mac}")
        |> last()
    '''

    try:
        result = query_api.query(query)
        data = {}

        for table in result:
            for record in table.records:
                data[record.get_field()] = record.get_value()
                data["mac"] = record.values.get("mac")
                data["deviceId"] = record.values.get("deviceId")
                data["timestamp"] = record.get_time().isoformat()

        if not data:
            raise HTTPException(status_code=404, detail="No data found for the specified MAC")

        return {"status": "success", "data": data}

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/v1/motion/events")
async def get_motion_events(
    mac: str = Query(..., description="MAC address of device"),
    from_time: str = Query(..., description="Start time (RFC3339)", alias="from"),
    to_time: str = Query(..., description="End time (RFC3339)", alias="to")
):
    """Obtener eventos de movimiento en un rango de tiempo"""
    query = f'''
    from(bucket: "{INFLUX_BUCKET}")
        |> range(start: {from_time}, stop: {to_time})
        |> filter(fn: (r) => r._measurement == "motion_sensor_metrics")
        |> filter(fn: (r) => r._field == "motion_detected")
        |> filter(fn: (r) => r.mac == "{mac}")
        |> filter(fn: (r) => r._value == true)
        |> sort(columns: ["_time"])
    '''

    try:
        result = query_api.query(query)
        events = []

        for table in result:
            for record in table.records:
                event = {
                    "mac": record.values.get("mac"),
                    "motion_detected": record.get_value(),
                    "timestamp": record.get_time().isoformat()
                }
                events.append(event)

        return {
            "status": "success",
            "data": events,
            "count": len(events)
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/v1/light/averages")
async def get_light_averages(period: str = Query("hour", description="Aggregation period: hour, day")):
    """Obtener promedios de luz por período"""

    window_period = "1h" if period == "hour" else "1d"

    query = f'''
    from(bucket: "{INFLUX_BUCKET}")
        |> range(start: -24h)
        |> filter(fn: (r) => r._measurement == "light_sensor_metrics")
        |> filter(fn: (r) => r._field == "light_level")
        |> aggregateWindow(every: {window_period}, fn: mean)
        |> group(columns: ["mac"])
    '''

    try:
        result = query_api.query(query)
        averages = []

        for table in result:
            for record in table.records:
                avg = {
                    "mac": record.values.get("mac"),
                    "average_light_level": record.get_value(),
                    "period": period,
                    "timestamp": record.get_time().isoformat()
                }
                averages.append(avg)

        return {
            "status": "success",
            "data": averages,
            "count": len(averages)
        }

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

---

## 🌐 **Implementación Node.js (Express)**

### **app.js**

```javascript
const express = require("express");
const cors = require("cors");
const { InfluxDB } = require("@influxdata/influxdb-client");

const app = express();
app.use(cors());
app.use(express.json());

// InfluxDB Configuration
const url = "http://52.201.107.193:8086";
const token =
  "lJLzxtHLHvPNgdvU9dcInGYb/qLbLxUPgrePzLd47EKCLUWBzJ+RmJkpH0f1HkmQ";
const org = "mi-org";
const bucket = "sensores";

const influxDB = new InfluxDB({ url, token });
const queryApi = influxDB.getQueryApi(org);

// Energy Endpoints
app.get("/api/v1/energy/current", async (req, res) => {
  const { mac } = req.query;

  if (!mac) {
    return res.status(400).json({
      status: "error",
      message: "MAC parameter is required",
    });
  }

  const query = `
        from(bucket: "${bucket}")
        |> range(start: -1h)
        |> filter(fn: (r) => r._measurement == "energy_metrics")
        |> filter(fn: (r) => r.mac == "${mac}")
        |> last()
    `;

  try {
    const data = {};

    await queryApi.queryRows(query, {
      next(row, tableMeta) {
        const o = tableMeta.toObject(row);
        data[o._field] = o._value;
        data.mac = o.mac;
        data.deviceId = o.deviceId;
        data.timestamp = o._time;
      },
      error(error) {
        console.error(error);
        res.status(500).json({
          status: "error",
          message: error.message,
        });
      },
      complete() {
        res.json({
          status: "success",
          data: data,
        });
      },
    });
  } catch (error) {
    res.status(500).json({
      status: "error",
      message: error.message,
    });
  }
});

app.get("/api/v1/environment/temperature", async (req, res) => {
  const { mac, from, to } = req.query;

  if (!mac || !from || !to) {
    return res.status(400).json({
      status: "error",
      message: "mac, from, and to parameters are required",
    });
  }

  const query = `
        from(bucket: "${bucket}")
        |> range(start: ${from}, stop: ${to})
        |> filter(fn: (r) => r._measurement == "temperature_humidity_metrics")
        |> filter(fn: (r) => r._field == "temperature")
        |> filter(fn: (r) => r.mac == "${mac}")
        |> sort(columns: ["_time"])
    `;

  try {
    const data = [];

    await queryApi.queryRows(query, {
      next(row, tableMeta) {
        const o = tableMeta.toObject(row);
        data.push({
          mac: o.mac,
          deviceId: o.deviceId,
          temperature: o._value,
          timestamp: o._time,
        });
      },
      error(error) {
        console.error(error);
        res.status(500).json({
          status: "error",
          message: error.message,
        });
      },
      complete() {
        res.json({
          status: "success",
          data: data,
          count: data.length,
        });
      },
    });
  } catch (error) {
    res.status(500).json({
      status: "error",
      message: error.message,
    });
  }
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`API running on port ${PORT}`);
});
```

---

## 📋 **Docker Compose para Despliegue**

### **docker-compose.yml**

```yaml
version: "3.8"

services:
  voltio-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - INFLUX_URL=http://52.201.107.193:8086
      - INFLUX_TOKEN=lJLzxtHLHvPNgdvU9dcInGYb/qLbLxUPgrePzLd47EKCLUWBzJ+RmJkpH0f1HkmQ
      - INFLUX_ORG=mi-org
      - INFLUX_BUCKET=sensores
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - voltio-api
    restart: unless-stopped
```

Esta guía proporciona implementaciones completas en 3 lenguajes diferentes para crear APIs que consuman los datos del sistema Voltio desde InfluxDB.
