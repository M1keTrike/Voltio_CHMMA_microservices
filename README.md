# Voltio CHMMA Microservices

[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A real-time IoT sensor monitoring and automation system built with Go microservices, RabbitMQ message queues, WebSockets, and time-series data storage.

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Project Structure](#project-structure)
- [Services](#services)
- [Development Guide](#development-guide)
- [Testing](#testing)
- [Configuration](#configuration)
- [Contributing](#contributing)
- [Documentation](#documentation)

## 🌟 Overview

Voltio CHMMA is an IoT sensor monitoring platform that collects, processes, and streams real-time data from various sensors including:

- **DHT22**: Temperature and humidity sensors
- **PZEM**: Electric power meters (voltage, current, power, energy)
- **PIR**: Motion detection sensors
- **Light Sensors**: Ambient light level monitoring

The system uses a microservices architecture where each sensor type has dedicated producer and consumer services that communicate through RabbitMQ message queues. Data is stored in InfluxDB for time-series analysis and streamed to clients via WebSockets for real-time monitoring.

## 🏗️ Architecture

```
┌─────────────────┐
│  IoT Sensors    │
│  (DHT22, PZEM,  │
│  PIR, Light)    │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│        Test Producers (Simulators)       │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │  DHT22  │ │  PZEM   │ │   PIR   │   │
│  │Producer │ │Producer │ │Producer │   │
│  └────┬────┘ └────┬────┘ └────┬────┘   │
└───────┼───────────┼──────────┼─────────┘
        │           │          │
        └───────────┼──────────┘
                    ▼
        ┌───────────────────────┐
        │      RabbitMQ         │
        │   Message Broker      │
        └───────────┬───────────┘
                    │
        ┌───────────┴──────────────────────┐
        ▼                                   ▼
┌──────────────────┐            ┌──────────────────┐
│  Consumer        │            │   Automation     │
│  Services        │            │   Engine         │
│                  │            │                  │
│ - DHT22          │            │ - Rule Engine    │
│ - PZEM           │            │ - Alert System   │
│ - PIR            │            │ - PostgreSQL     │
│ - Light          │            │                  │
│ - Notification   │            └──────────────────┘
└────────┬─────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌──────┐  ┌──────────┐
│InfluxDB  WebSocket  │
│Time-Series│ Server   │
│  Storage  │          │
└──────┘  └────┬─────┘
               │
               ▼
          ┌─────────┐
          │ Clients │
          │ (Web/   │
          │  Mobile)│
          └─────────┘
```

### Data Flow

1. **Producers** simulate IoT sensors and publish sensor data to RabbitMQ queues
2. **Consumer Services** subscribe to queues, process messages, and:
   - Store data in **InfluxDB** for historical analysis
   - Forward data to **WebSocket Server** for real-time streaming
   - Send alerts when timeout conditions are detected
3. **Automation Engine** consumes events and triggers automated actions based on predefined rules
4. **WebSocket Server** broadcasts real-time data to connected clients
5. **Notification Service** sends alerts to external systems via webhooks

## ✨ Features

- **Real-time Data Streaming**: WebSocket-based live sensor data delivery
- **Time-Series Storage**: InfluxDB integration for historical data analysis
- **Alert System**: Automatic alerts when sensors stop reporting (timeout detection)
- **Automation Engine**: Rule-based automation for IoT events
- **Microservices Architecture**: Independently scalable services
- **Environment-Based Configuration**: Secure credential management via environment variables
- **Multi-Sensor Support**: Extensible architecture for various sensor types
- **Test Producers**: Built-in sensor simulators for development and testing

## 📦 Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.19 or higher** - [Download](https://golang.org/dl/)
- **RabbitMQ** - Message broker ([Installation Guide](https://www.rabbitmq.com/download.html))
- **InfluxDB 2.x** - Time-series database ([Installation Guide](https://docs.influxdata.com/influxdb/v2/install/))
- **PostgreSQL** - Required for automation engine ([Installation Guide](https://www.postgresql.org/download/))
- **Git** - Version control

### Optional Tools

- **Docker & Docker Compose** - For containerized deployment
- **Caddy** - Reverse proxy for WebSocket server (included in WebSocketServer service)

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/M1keTrike/Voltio_CHMMA_microservices.git
cd Voltio_CHMMA_microservices
```

### 2. Set Up Environment Variables

Copy the example environment files and configure them with your credentials:

```bash
# Root environment file (for consumers, producers, WebSocket server)
cp .env.example .env

# Automation engine environment file
cp Backend/automation-engine/.env.example Backend/automation-engine/.env
```

Edit both `.env` files with your actual configuration. See [ENVIRONMENT_SETUP.md](./ENVIRONMENT_SETUP.md) for detailed configuration options.

**Minimum required configuration:**

```bash
# .env
RABBITMQ_URI=amqp://user:password@localhost:5672/
INFLUX_TOKEN=your-influxdb-token
INFLUX_ORG=your-org
INFLUX_BUCKET=your-bucket

# Backend/automation-engine/.env
RABBITMQ_URI=amqp://user:password@localhost:5672/
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=voltio_user
POSTGRES_PASSWORD=your-password
POSTGRES_DB=voltiodb
```

### 3. Install Dependencies

Navigate to each service directory and install Go modules:

```bash
# Install dependencies for all services
cd Backend/test_producers && go mod tidy && cd ../..
cd Backend/DHT22_ConsumerSender && go mod tidy && cd ../..
cd Backend/PZEM_ConsumerSender && go mod tidy && cd ../..
cd Backend/PIR_ConsumerSender && go mod tidy && cd ../..
cd Backend/LightSensor_ConsumerSender && go mod tidy && cd ../..
cd Backend/Notification_ConsumerSender && go mod tidy && cd ../..
cd Backend/WebSocketServer && go mod tidy && cd ../..
cd Backend/automation-engine && go mod tidy && cd ../..
```

### 4. Start Services

#### Option A: Using PowerShell Script (Windows)

```powershell
# Start all services at once
.\start_all_voltio_services.ps1
```

#### Option B: Manual Start

Start each service in a separate terminal:

```bash
# Terminal 1: WebSocket Server
cd Backend/WebSocketServer
go run cmd/main.go

# Terminal 2: Automation Engine
cd Backend/automation-engine
go run main.go

# Terminal 3-7: Consumer Services
cd Backend/DHT22_ConsumerSender
go run middleware/RabbitToSocketMiddleware.go

cd Backend/PZEM_ConsumerSender
go run middleware/RabbitToSocketMiddleware.go

cd Backend/PIR_ConsumerSender
go run middleware/RabbitToSocketMiddleware.go

cd Backend/LightSensor_ConsumerSender
go run middleware/RabbitToSocketMiddleware.go

cd Backend/Notification_ConsumerSender
go run middleware/RabbitToSocketMiddleware.go

# Terminal 8-11: Test Producers (optional, for testing)
cd Backend/test_producers/dht22
go run main.go

cd Backend/test_producers/pzem
go run main.go

cd Backend/test_producers/pir
go run main.go

cd Backend/test_producers/light
go run main.go
```

### 5. Verify Everything is Working

Check that:
- Services connect to RabbitMQ successfully
- Consumer services are writing to InfluxDB
- WebSocket server is accepting connections (default port: 8081)
- Test producers are generating and sending data

## 📁 Project Structure

```
Voltio_CHMMA_microservices/
├── .env.example                          # Environment variables template
├── .gitignore                            # Git ignore rules
├── README.md                             # This file
├── ENVIRONMENT_SETUP.md                  # Detailed environment configuration guide
├── SECURITY_MIGRATION.md                 # Security best practices and migration guide
├── start_all_voltio_services.ps1        # PowerShell script to start all services
│
└── Backend/                              # All microservices
    │
    ├── DHT22_ConsumerSender/             # DHT22 temperature/humidity consumer
    │   ├── go.mod
    │   ├── go.sum
    │   └── middleware/
    │       └── RabbitToSocketMiddleware.go
    │
    ├── PZEM_ConsumerSender/              # PZEM electric meter consumer
    │   ├── README.md
    │   ├── IMPLEMENTATION_GUIDE.md
    │   ├── go.mod
    │   ├── go.sum
    │   └── middleware/
    │       └── RabbitToSocketMiddleware.go
    │
    ├── PIR_ConsumerSender/               # PIR motion sensor consumer
    │   ├── go.mod
    │   ├── go.sum
    │   ├── main.go
    │   └── middleware/
    │       └── RabbitToSocketMiddleware.go
    │
    ├── LightSensor_ConsumerSender/       # Light sensor consumer
    │   ├── go.mod
    │   ├── go.sum
    │   └── middleware/
    │       └── RabbitToSocketMiddleware.go
    │
    ├── Notification_ConsumerSender/      # Notification/alert consumer
    │   ├── go.mod
    │   ├── go.sum
    │   └── middleware/
    │       └── RabbitToSocketMiddleware.go
    │
    ├── WebSocketServer/                  # Real-time WebSocket server
    │   ├── Caddyfile                     # Caddy reverse proxy configuration
    │   ├── go.mod
    │   ├── go.sum
    │   ├── cmd/
    │   │   └── main.go                   # Server entry point
    │   └── internal/                     # Internal packages
    │       ├── adapters/                 # External adapters
    │       ├── core/                     # Business logic
    │       ├── models/                   # Data models
    │       ├── ports/                    # Interfaces
    │       └── server/                   # Server implementation
    │
    ├── automation-engine/                # Rule-based automation engine
    │   ├── Dockerfile
    │   ├── .env.example
    │   ├── go.mod
    │   ├── go.sum
    │   ├── main.go
    │   ├── config/                       # Configuration management
    │   ├── database/                     # PostgreSQL integration
    │   ├── messaging/                    # RabbitMQ integration
    │   ├── models/                       # Data models
    │   └── rules/                        # Rule engine logic
    │
    └── test_producers/                   # Sensor simulators for testing
        ├── README.md                     # Test producers documentation
        ├── go.mod
        ├── go.sum
        ├── start_all_producers.ps1       # Start all producers script
        ├── start_single_producer.ps1     # Start individual producer script
        ├── start_all_producers.bat       # Batch script (Windows)
        ├── start_single_producer.bat     # Batch script (Windows)
        ├── dht22/                        # DHT22 test producer
        │   └── main.go
        ├── pzem/                         # PZEM test producer
        │   └── main.go
        ├── pir/                          # PIR test producer
        │   └── main.go
        └── light/                        # Light sensor test producer
            └── main.go
```

## 🔧 Services

### Consumer Services

Each consumer service follows the same pattern:
- Subscribes to a dedicated RabbitMQ queue
- Processes incoming sensor messages
- Writes data to InfluxDB for storage
- Forwards data to WebSocket server for real-time streaming
- Monitors for sensor timeouts and triggers alerts

**Available Consumers:**
- `DHT22_ConsumerSender` - Temperature and humidity data
- `PZEM_ConsumerSender` - Electrical measurements
- `PIR_ConsumerSender` - Motion detection events
- `LightSensor_ConsumerSender` - Light level measurements
- `Notification_ConsumerSender` - Alert notifications

### WebSocket Server

Provides real-time data streaming to connected clients:
- Topic-based subscription system
- Support for multiple concurrent clients
- Automatic reconnection handling
- Runs on port 8081 by default (configurable)

### Automation Engine

Rule-based automation system:
- Consumes sensor events from RabbitMQ
- Evaluates predefined rules from PostgreSQL
- Triggers automated actions based on conditions
- Supports time-based and event-based rules
- Caches rules for performance (refreshes every 5 minutes)

### Test Producers

Realistic sensor data simulators for development and testing:
- **DHT22 Producer**: Temperature (20-35°C) and humidity (40-70%) with day/night patterns
- **PZEM Producer**: Electrical data with realistic power consumption patterns
- **PIR Producer**: Motion detection with location-based probabilities
- **Light Producer**: Indoor and outdoor light levels with time-based variations

Each producer publishes data at configurable intervals to simulate real sensors.

## 💻 Development Guide

### Adding a New Sensor Type

1. **Create Producer** (in `Backend/test_producers/newsensor/`)
   - Define message structure
   - Implement data generation logic
   - Publish to dedicated RabbitMQ queue

2. **Create Consumer** (in `Backend/NewSensor_ConsumerSender/`)
   - Subscribe to the sensor's queue
   - Implement InfluxDB write logic
   - Implement WebSocket forwarding
   - Add timeout monitoring

3. **Update Configuration**
   - Add queue name to `.env.example`
   - Add WebSocket URI to `.env.example`
   - Document environment variables in `ENVIRONMENT_SETUP.md`

4. **Update Scripts**
   - Add to `start_all_voltio_services.ps1`
   - Add to test producer scripts

### Code Style

- Follow standard Go conventions and formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for complex logic
- Handle errors appropriately
- Use environment variables for configuration

### Building

Build individual services:

```bash
# Build a consumer service
cd Backend/DHT22_ConsumerSender
go build -o dht22_consumer middleware/RabbitToSocketMiddleware.go

# Build automation engine
cd Backend/automation-engine
go build -o automation-engine main.go

# Build test producer
cd Backend/test_producers/dht22
go build -o dht22_producer main.go
```

## 🧪 Testing

### Running Test Producers

Test the system using the built-in sensor simulators:

```powershell
# Windows PowerShell - Start all producers
cd Backend/test_producers
.\start_all_producers.ps1

# Or start individual producers
.\start_single_producer.ps1 dht22
.\start_single_producer.ps1 pzem
.\start_single_producer.ps1 pir
.\start_single_producer.ps1 light
```

```bash
# Linux/Mac - Start individual producer
cd Backend/test_producers/dht22
go run main.go
```

### Testing Alert System

1. Start a producer to generate data
2. Stop the producer
3. Wait for timeout period (default: 2 minutes)
4. Verify alert is sent to alerts queue
5. Check notification consumer processes the alert

### Verifying Data Flow

1. **Check RabbitMQ**: Access management UI (default: http://localhost:15672)
   - Verify queues are created
   - Monitor message rates
   - Check for message backlog

2. **Check InfluxDB**: Query data to verify storage
   ```bash
   influx query 'from(bucket:"sensores") |> range(start: -1h) |> filter(fn: (r) => r._measurement == "dht22")'
   ```

3. **Check WebSocket**: Use a WebSocket client to connect
   ```javascript
   const ws = new WebSocket('ws://localhost:8081/ws?topic=dht22');
   ws.onmessage = (event) => console.log(event.data);
   ```

## ⚙️ Configuration

Configuration is managed through environment variables. See [ENVIRONMENT_SETUP.md](./ENVIRONMENT_SETUP.md) for comprehensive configuration documentation.

### Key Configuration Files

- **`.env`** - Root level services (consumers, producers, WebSocket server)
- **`Backend/automation-engine/.env`** - Automation engine specific configuration

### Important Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `RABBITMQ_URI` | RabbitMQ connection URI | `amqp://guest:guest@localhost:5672/` |
| `INFLUX_TOKEN` | InfluxDB authentication token | (required) |
| `INFLUX_URL` | InfluxDB server URL | `http://localhost:8086` |
| `INFLUX_ORG` | InfluxDB organization | `mi-org` |
| `INFLUX_BUCKET` | InfluxDB bucket name | `sensores` |
| `WEBSOCKET_PORT` | WebSocket server port | `8081` |
| `POSTGRES_HOST` | PostgreSQL host (automation engine) | `localhost` |

## 🤝 Contributing

We welcome contributions! Here's how you can help:

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Make your changes**
   - Follow the code style guidelines
   - Add tests if applicable
   - Update documentation
4. **Commit your changes**
   ```bash
   git commit -m "Add amazing feature"
   ```
5. **Push to your branch**
   ```bash
   git push origin feature/amazing-feature
   ```
6. **Open a Pull Request**

### Development Workflow

- Write clear, descriptive commit messages
- Keep pull requests focused on a single feature/fix
- Update documentation for user-facing changes
- Ensure all services build and run successfully
- Test with the included test producers

### Reporting Issues

When reporting issues, please include:
- Go version
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Relevant log output

## 📚 Documentation

- **[ENVIRONMENT_SETUP.md](./ENVIRONMENT_SETUP.md)** - Complete guide to environment variables and configuration
- **[SECURITY_MIGRATION.md](./SECURITY_MIGRATION.md)** - Security best practices and credential management
- **[Backend/test_producers/README.md](./Backend/test_producers/README.md)** - Test producer documentation and usage
- **[Backend/PZEM_ConsumerSender/README.md](./Backend/PZEM_ConsumerSender/README.md)** - PZEM consumer implementation details

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with [Go](https://golang.org/)
- Message queuing by [RabbitMQ](https://www.rabbitmq.com/)
- Time-series storage by [InfluxDB](https://www.influxdata.com/)
- WebSocket support by [Gorilla WebSocket](https://github.com/gorilla/websocket)
- Reverse proxy by [Caddy](https://caddyserver.com/)

## 📞 Support

For questions, issues, or feature requests:
- Open an issue on [GitHub](https://github.com/M1keTrike/Voltio_CHMMA_microservices/issues)
- Check existing documentation
- Review closed issues for solutions

---

**Happy Coding! 🚀**

*Made with ❤️ by the Voltio CHMMA Team*
