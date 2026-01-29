# Test Producers - Sensor Simulators

This directory contains test producers that simulate IoT sensors, generating realistic data for all sensor types in the Voltio system. These are essential for development, testing, and demonstrating the system without physical hardware.

## 📋 Overview

Test producers simulate real IoT devices by:
- Generating realistic sensor data with time-based patterns
- Publishing messages to RabbitMQ queues
- Mimicking actual sensor behavior and reporting frequencies
- Supporting multiple device instances per sensor type

## 📁 Project Structure

```
test_producers/
├── README.md
├── go.mod
├── go.sum
├── start_all_producers.ps1      # PowerShell script to run all producers
├── start_single_producer.ps1    # PowerShell script to run a single producer
├── start_all_producers.bat      # Batch script to run all producers (Windows)
├── start_single_producer.bat    # Batch script to run a single producer (Windows)
├── dht22/
│   └── main.go                  # DHT22 Temperature/Humidity producer
├── light/
│   └── main.go                  # Light sensor producer
├── pir/
│   └── main.go                  # PIR motion sensor producer
└── pzem/
    └── main.go                  # PZEM electric meter producer
```


## 🌡️ Available Producers

### 1. DHT22 Producer (dht22/main.go)

Simulates temperature and humidity sensors commonly used in environmental monitoring.

- **Queue**: `DHT22_queue`
- **Data**: Temperature and humidity readings
- **Frequency**: Every 30 seconds
- **Simulated Devices**: 3 sensors
- **Message Format**:

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

**Realistic Patterns:**
- Temperature: 20-35°C with day/night variations
- Humidity: 40-70% with inverse correlation to temperature
- Simulates natural environmental changes

### 2. Light Sensor Producer (light/main.go)

Simulates ambient light level sensors for indoor and outdoor environments.

- **Queue**: `LightSensor_queue`
- **Data**: Light level in lux
- **Frequency**: Every 15 seconds
- **Simulated Devices**: 4 sensors (indoor and outdoor)
- **Message Format**:

```json
{
  "deviceId": "LIGHT-DEV-001",
  "payload": {
    "mac": "LIGHT-001",
    "lightLevel": 1250.5
  }
}
```

**Realistic Patterns:**
- Outdoor sensors: 0-100,000 lux (solar patterns)
- Indoor sensors: 0-1,000 lux (human activity patterns)
- Time-based variations throughout the day

### 3. PIR Motion Sensor Producer (pir/main.go)

Simulates passive infrared motion detection sensors for occupancy monitoring.

- **Queue**: `PIR_queue`
- **Data**: Motion detection (boolean)
- **Frequency**: Every 20 seconds
- **Simulated Devices**: 6 sensors in different locations
- **Message Format**:

```json
{
  "deviceId": "PIR-DEV-001",
  "payload": {
    "mac": "PIR-001",
    "motionDetected": true
  }
}
```

**Realistic Patterns:**
- Location-based motion probabilities
- Higher activity during 6 AM - 10 PM
- Different activity levels per zone (entrance, hallway, room, etc.)

### 4. PZEM Electric Meter Producer (pzem/main.go)

Simulates PZEM-004T electric power meters for energy monitoring.

- **Queue**: `PZEM_queue`
- **Data**: Complete electrical measurements
- **Frequency**: Every 10 seconds
- **Simulated Devices**: 4 meters on different circuits
- **Message Format**:

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

**Realistic Patterns:**
- Consumption patterns by load type
- Variations based on time of day
- Complete electrical measurements with power factor



## ⚙️ Configuration

All producers use environment variables for configuration:

**Environment Variable:**
- `RABBITMQ_URI` - RabbitMQ connection URI

**Default Configuration (if not set):**
- RabbitMQ URI: `amqp://guest:guest@localhost:5672/`
- This default works for local development with standard RabbitMQ installation

**Dependencies:**
- `github.com/rabbitmq/amqp091-go` - RabbitMQ Go client

### Setting Custom Configuration

Create a `.env` file in the root directory or export environment variables:

```bash
# For local RabbitMQ
export RABBITMQ_URI="amqp://guest:guest@localhost:5672/"

# Or for remote RabbitMQ
export RABBITMQ_URI="amqp://username:password@hostname:5672/"
```

## 🚀 Installation and Setup

### Prerequisites

- **Go 1.19 or higher** - [Download](https://golang.org/dl/)
- **RabbitMQ** - Running locally or accessible remotely

### Installation

```bash
# Navigate to test producers directory
cd Backend/test_producers

# Download dependencies
go mod tidy

# Verify installation
go mod verify
```

## 🎮 Running Producers

### Option 1: PowerShell Scripts (Recommended for Windows)

**Start all producers at once:**
```powershell
.\start_all_producers.ps1
```
This will open 4 separate PowerShell windows, one for each producer.

**Start a specific producer:**
```powershell
.\start_single_producer.ps1 dht22
.\start_single_producer.ps1 light
.\start_single_producer.ps1 pir
.\start_single_producer.ps1 pzem
```

### Option 2: Batch Scripts (Windows CMD)

**Start all producers at once:**
```batch
start_all_producers.bat
```

**Start a specific producer:**
```batch
start_single_producer.bat dht22
start_single_producer.bat light
start_single_producer.bat pir
start_single_producer.bat pzem
```

### Option 3: Manual Execution (All Platforms)

Start each producer in a separate terminal:

```bash
# DHT22 Producer
cd dht22
go run main.go

# Light Sensor Producer
cd light
go run main.go

# PIR Motion Sensor Producer
cd pir
go run main.go

# PZEM Electric Meter Producer
cd pzem
go run main.go
```

## 📊 Generated Data

The producers generate realistic sensor data with intelligent patterns:

### 🌡️ DHT22 (Temperature/Humidity)

- **Temperature variations** based on time of day
- **Inverse correlation** between temperature and humidity
- **Range**: 20-35°C temperature, 40-70% humidity
- **Pattern**: Warmer during afternoon, cooler at night

### 💡 Light Sensor

- **Outdoor sensors**: Solar patterns (0-100,000 lux)
  - Sunrise simulation: gradual increase
  - Midday peak: maximum brightness
  - Sunset simulation: gradual decrease
  - Night: near-zero values
- **Indoor sensors**: Human activity patterns (0-1,000 lux)
  - Higher during active hours (8 AM - 10 PM)
  - Lower during sleep hours
- **Time-based variations** throughout the day

### 🚶 PIR Motion

- **Location-based probabilities**:
  - Entrance: Higher motion probability
  - Hallways: Medium motion probability
  - Storage: Lower motion probability
- **Time-based patterns**: More activity 6 AM - 10 PM
- **Realistic randomness**: Not all sensors detect motion simultaneously
- **Different activity levels** per zone

### ⚡ PZEM Electric

- **Load-type patterns**:
  - Main circuit: High, variable consumption
  - HVAC circuit: Cyclic patterns
  - Lighting: Time-based (on at night)
  - Appliances: Intermittent usage
- **Complete measurements**:
  - Voltage: ~220V with minor variations
  - Current: Proportional to power consumption
  - Power: Calculated from voltage and current
  - Energy: Cumulative consumption
  - Frequency: ~50/60 Hz based on region
  - Power Factor: 0.85-0.98 (realistic range)
- **Time-based variations** reflecting usage patterns

## 🎯 Testing Purposes

These producers enable comprehensive testing of:

### 1. Timeout Alert System
- **Test**: Stop one or more producers
- **Expected**: Alert triggered after timeout period (2 minutes default)
- **Verifies**: Consumer timeout detection and alert generation

### 2. Data Flow Validation
- **Test**: Run producers and verify data pipeline
- **Verifies**:
  - Messages published to RabbitMQ queues
  - Consumer services receive messages
  - Data written to InfluxDB
  - Real-time streaming via WebSocket
  - Data integrity throughout pipeline

### 3. System Load Testing
- **Test**: Run multiple producers simultaneously
- **Verifies**:
  - System handles concurrent sensor data
  - No message loss under load
  - Performance metrics remain acceptable
  - Resource utilization is reasonable

### 4. Critical Alerts
- **Test**: PZEM producer can generate conditions for electrical alerts
- **Verifies**:
  - High power consumption detection
  - Voltage anomaly detection
  - Alert notification system
  - Automation engine rule triggers

### 5. Development Workflow
- **Test**: Develop new features without physical sensors
- **Benefits**:
  - Consistent, repeatable test data
  - No hardware dependencies
  - Easy to modify for edge cases
  - Quick iteration cycles

## 🛑 Stopping Producers

### PowerShell/Batch Scripts
Close the corresponding CMD or PowerShell windows that were opened.

### Manual Execution
Press `Ctrl+C` in each terminal running a producer.

**Note**: Producers will attempt to gracefully close their RabbitMQ connections before exiting.

## 📝 Log Output

Each producer displays informative logs:

### Startup Logs
```
🌡️ Starting DHT22 Test Producer...
✅ Connected to RabbitMQ - Queue: DHT22_queue
🔄 Publishing data every 30 seconds...
```

### Data Publishing Logs
```
📤 [DHT22-001] 24.5°C, 62.3%
📤 [DHT22-002] 26.1°C, 58.7%
📤 [DHT22-003] 23.8°C, 64.1%
```

### Error Logs
```
❌ Error connecting to RabbitMQ: dial tcp: connection refused
❌ Error publishing message: channel closed
```

### Log Indicators
- ✅ Success/Connection established
- 📤 Data sent successfully
- ❌ Errors (connection, publishing)
- 🔄 Status/Frequency information

## 🔧 Customization

### Modifying Data Patterns

Edit the `main.go` file in each producer directory to customize:

1. **Sending Frequency**:
   ```go
   time.Sleep(30 * time.Second) // Change to desired interval
   ```

2. **Data Ranges**:
   ```go
   baseTemp := 20.0 + rand.Float64()*15.0 // Adjust min/max values
   ```

3. **Number of Devices**:
   ```go
   devices := []struct {
       DeviceID string
       MAC      string
   }{
       {"DHT22-DEV-001", "DHT22-001"},
       {"DHT22-DEV-002", "DHT22-002"},
       // Add more devices here
   }
   ```

4. **Message Format**:
   Modify the struct definitions to add or change fields.

### Adding a New Producer

1. Create a new directory (e.g., `newsensor/`)
2. Copy structure from an existing producer
3. Modify message format and data generation logic
4. Update `start_all_producers` scripts to include new producer
5. Add environment variable configuration if needed

## 🐛 Troubleshooting

### "Failed to connect to RabbitMQ"

**Problem**: Producer can't connect to RabbitMQ  
**Solutions**:
- Verify RabbitMQ is running: `rabbitmq-server`
- Check connection URI in environment variables
- Test connection: http://localhost:15672 (management UI)
- Verify firewall isn't blocking port 5672

### "Queue declaration failed"

**Problem**: RabbitMQ won't create queue  
**Solutions**:
- Check RabbitMQ permissions
- Verify queue name doesn't conflict
- Review RabbitMQ logs for details
- Try deleting and recreating queue in management UI

### Producer starts but no messages appear

**Problem**: Producer runs but consumers don't receive messages  
**Solutions**:
- Check RabbitMQ management UI: http://localhost:15672
- Verify queue exists and has published messages
- Check queue name matches consumer configuration
- Ensure consumers are running and connected

### High CPU usage

**Problem**: Producer using excessive CPU  
**Solutions**:
- Increase sleep duration between messages
- Reduce number of simulated devices
- Optimize data generation algorithms
- Check for infinite loops in modified code

## 📚 Additional Resources

- [RabbitMQ Go Client](https://github.com/rabbitmq/amqp091-go) - Official documentation
- [RabbitMQ Tutorials](https://www.rabbitmq.com/getstarted.html) - Learn RabbitMQ basics
- [Main Project README](../../README.md) - Complete system documentation
- [Environment Setup](../../ENVIRONMENT_SETUP.md) - Configuration guide

## 🤝 Contributing

Want to improve the test producers?

1. Add more realistic data patterns
2. Create producers for new sensor types
3. Improve error handling and logging
4. Add configuration options
5. Write documentation for new features

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

---

**Happy Testing! 🚀**

*For questions or issues, please open an issue on the [GitHub repository](https://github.com/M1keTrike/Voltio_CHMMA_microservices).*
