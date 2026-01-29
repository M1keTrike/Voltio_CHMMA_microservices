# PZEM Consumer Service

This service consumes electrical measurement data from PZEM-004T power meters via RabbitMQ, stores it in InfluxDB for time-series analysis, and forwards real-time data to WebSocket clients.

## 📋 Overview

The PZEM Consumer Service is part of the Voltio CHMMA microservices ecosystem. It handles electrical monitoring data including:
- Voltage (V)
- Current (A)
- Power (W)
- Energy consumption (kWh)
- Frequency (Hz)
- Power Factor

## 🏗️ Architecture

```
IoT Devices/Producers
        ↓
    RabbitMQ Queue
   (PZEM_queue)
        ↓
   PZEM Consumer
        ↓
    ┌───┴───┐
    ↓       ↓
InfluxDB  WebSocket
(Storage) (Real-time)
```

## 🚀 Quick Start

### Prerequisites

- Go 1.19 or higher
- RabbitMQ server running
- InfluxDB 2.x running
- Environment variables configured

### Running the Service

```bash
cd Backend/PZEM_ConsumerSender
go run middleware/RabbitToSocketMiddleware.go
```

## ⚙️ Configuration

The service uses environment variables from the root `.env` file:

```bash
# RabbitMQ Configuration
RABBITMQ_URI=amqp://user:password@localhost:5672/

# Queue Names
PZEM_QUEUE_NAME=PZEM_queue
ALERTS_QUEUE_NAME=alerts-queue

# InfluxDB Configuration
INFLUX_URL=http://localhost:8086
INFLUX_TOKEN=your-token-here
INFLUX_ORG=your-org
INFLUX_BUCKET=sensores

# WebSocket Configuration
PZEM_WEBSOCKET_URI=ws://localhost:8081/ws?topic=pzem&emitter=true
```

See [ENVIRONMENT_SETUP.md](../../ENVIRONMENT_SETUP.md) for complete configuration options.

## 📊 Message Format

The consumer expects messages in this JSON format:

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

## 🔍 Features

### 1. RabbitMQ Message Consumption
- Subscribes to `PZEM_queue`
- Automatically acknowledges processed messages
- Handles reconnection on connection loss

### 2. InfluxDB Storage
- Writes measurements to time-series database
- Tags by device MAC address and sensor type
- Stores all electrical parameters as fields
- Enables historical data analysis

### 3. WebSocket Forwarding
- Real-time data streaming to connected clients
- Automatic reconnection handling
- Topic-based subscription (topic: `pzem`)

### 4. Timeout Monitoring
- Monitors for device timeouts (default: 2 minutes)
- Sends alerts when devices stop reporting
- Publishes alerts to dedicated queue
- Helps detect device failures

## 📝 Logs

The service provides detailed logging:

### Startup
```
✅ Connected to RabbitMQ - Queue: PZEM_queue
✅ Connected to InfluxDB
✅ Connected to WebSocket server
```

### Normal Operation
```
📥 Received message from PZEM-001
💾 Written to InfluxDB: voltage=220.5V, current=5.25A, power=1157.6W
📡 Forwarded to WebSocket clients
```

### Alerts
```
⚠️ Timeout detected for device PZEM-001 (last seen: 2m ago)
📤 Alert sent to alerts-queue
```

## 🧪 Testing

### With Test Producer

1. Start the PZEM consumer service
2. Run the PZEM test producer:
   ```bash
   cd Backend/test_producers/pzem
   go run main.go
   ```
3. Verify logs show message flow

### Verify Data Storage

Query InfluxDB to check stored data:

```bash
influx query 'from(bucket:"sensores") 
  |> range(start: -1h) 
  |> filter(fn: (r) => r._measurement == "pzem")'
```

### Test Timeout Alert

1. Start consumer and producer
2. Stop the producer (Ctrl+C)
3. Wait 2+ minutes
4. Verify alert is generated in logs
5. Check alerts-queue in RabbitMQ management UI

## 🛠️ Development

### Building

```bash
cd Backend/PZEM_ConsumerSender
go build -o pzem_consumer middleware/RabbitToSocketMiddleware.go
```

### Dependencies

```bash
go mod download
go mod verify
```

## 🐛 Troubleshooting

### Connection Issues

**RabbitMQ Connection Failed**
- Verify RabbitMQ is running
- Check `RABBITMQ_URI` in `.env`
- Ensure queue exists or service has permission to create it

**InfluxDB Connection Failed**
- Verify InfluxDB is running
- Check `INFLUX_TOKEN` and `INFLUX_URL` in `.env`
- Ensure token has write permissions to bucket

**WebSocket Connection Failed**
- Verify WebSocket server is running
- Check `PZEM_WEBSOCKET_URI` in `.env`
- Ensure WebSocket server is accessible

### Data Issues

**Messages Not Consumed**
- Check RabbitMQ queue has messages
- Verify queue name matches configuration
- Review consumer logs for errors

**Data Not in InfluxDB**
- Verify InfluxDB connection is successful
- Check bucket and organization names
- Review write permissions for token

**WebSocket Not Receiving Data**
- Connect a test client to WebSocket
- Verify topic matches (`pzem`)
- Check WebSocket server logs

## 📚 Related Documentation

- [Main README](../../README.md) - Project overview
- [ENVIRONMENT_SETUP.md](../../ENVIRONMENT_SETUP.md) - Configuration guide
- [Test Producers](../test_producers/README.md) - Testing with simulators

## 🤝 Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for contribution guidelines.

---

For questions or issues, please open an issue on the [GitHub repository](https://github.com/M1keTrike/Voltio_CHMMA_microservices).
