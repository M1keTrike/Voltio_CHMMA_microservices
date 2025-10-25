# Getting Started with Voltio CHMMA

This guide will help you set up and run the Voltio CHMMA microservices system from scratch. Perfect for developers new to the project or IoT systems.

## 📖 What You'll Learn

By following this guide, you will:
1. Set up all required dependencies
2. Configure the system for local development
3. Start all microservices
4. Run test producers to simulate IoT sensors
5. Verify data is flowing through the system
6. Understand the basic architecture

## ⏱️ Estimated Time

- **Quick setup (Docker)**: ~30 minutes
- **Full setup (native)**: ~1-2 hours

## 📋 Prerequisites Check

Before you begin, verify you have:

- [ ] A computer with at least 8GB RAM
- [ ] Windows 10+, macOS 10.15+, or Linux (Ubuntu 20.04+)
- [ ] Administrator/sudo access for installing software
- [ ] Internet connection for downloading dependencies
- [ ] Basic command line knowledge
- [ ] A code editor (VS Code, GoLand, or similar)

## 🛠️ Step 1: Install Required Software

### Install Go

**Windows:**
1. Download Go from https://golang.org/dl/
2. Run the installer (e.g., `go1.21.0.windows-amd64.msi`)
3. Follow the installation wizard
4. Verify installation:
   ```powershell
   go version
   ```

**macOS:**
```bash
# Using Homebrew
brew install go

# Verify installation
go version
```

**Linux (Ubuntu/Debian):**
```bash
# Download and install
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile

# Verify installation
go version
```

### Install RabbitMQ

**Windows:**
1. Install Erlang first: https://www.erlang.org/downloads
2. Download RabbitMQ: https://www.rabbitmq.com/install-windows.html
3. Run the installer
4. RabbitMQ will start automatically as a service

**macOS:**
```bash
brew install rabbitmq
brew services start rabbitmq
```

**Linux (Ubuntu/Debian):**
```bash
# Install RabbitMQ
sudo apt-get update
sudo apt-get install -y rabbitmq-server

# Start RabbitMQ
sudo systemctl enable rabbitmq-server
sudo systemctl start rabbitmq-server

# Enable management plugin
sudo rabbitmq-plugins enable rabbitmq_management
```

**Verify RabbitMQ:**
- Management UI: http://localhost:15672
- Default credentials: `guest` / `guest`

### Install InfluxDB

**Windows:**
1. Download InfluxDB 2.x from https://portal.influxdata.com/downloads/
2. Extract the zip file
3. Run `influxd.exe` to start the server
4. Open http://localhost:8086 to complete setup

**macOS:**
```bash
brew install influxdb
brew services start influxdb
```

**Linux (Ubuntu/Debian):**
```bash
# Download and install
wget https://dl.influxdata.com/influxdb/releases/influxdb2-2.7.1-amd64.deb
sudo dpkg -i influxdb2-2.7.1-amd64.deb

# Start InfluxDB
sudo systemctl enable influxdb
sudo systemctl start influxdb
```

**Set Up InfluxDB:**
1. Open http://localhost:8086
2. Click "Get Started"
3. Create initial user:
   - Username: `voltio-admin`
   - Password: (choose a strong password)
   - Organization: `voltio-org`
   - Bucket: `sensores`
4. Save your API token - you'll need it later!

### Install PostgreSQL

**Windows:**
1. Download from https://www.postgresql.org/download/windows/
2. Run the installer
3. Remember the password you set for the `postgres` user

**macOS:**
```bash
brew install postgresql
brew services start postgresql
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get update
sudo apt-get install -y postgresql postgresql-contrib
sudo systemctl enable postgresql
sudo systemctl start postgresql
```

**Set Up PostgreSQL:**
```bash
# Connect to PostgreSQL
sudo -u postgres psql

# In PostgreSQL prompt, run:
CREATE USER voltio_user WITH PASSWORD 'your_secure_password';
CREATE DATABASE voltiodb OWNER voltio_user;
GRANT ALL PRIVILEGES ON DATABASE voltiodb TO voltio_user;
\q
```

### Install Git (if not already installed)

**Windows:**
- Download from https://git-scm.com/download/win

**macOS:**
```bash
brew install git
```

**Linux:**
```bash
sudo apt-get install -y git
```

## 📥 Step 2: Clone the Repository

```bash
# Clone the repository
git clone https://github.com/M1keTrike/Voltio_CHMMA_microservices.git

# Navigate into the directory
cd Voltio_CHMMA_microservices

# Check that you're in the right place
ls -la
# You should see: Backend/, .env.example, README.md, etc.
```

## ⚙️ Step 3: Configure Environment Variables

### 3.1 Create Root .env File

```bash
# Copy the example file
cp .env.example .env
```

Edit `.env` with your favorite text editor and update these values:

```bash
# RabbitMQ Configuration
RABBITMQ_URI=amqp://guest:guest@localhost:5672/

# InfluxDB Configuration
INFLUX_URL=http://localhost:8086
INFLUX_TOKEN=YOUR_INFLUXDB_TOKEN_HERE  # From InfluxDB setup step
INFLUX_ORG=voltio-org
INFLUX_BUCKET=sensores

# Queue Names (these defaults are fine)
ALERTS_QUEUE_NAME=alerts-queue
PZEM_QUEUE_NAME=PZEM_queue
DHT22_QUEUE_NAME=DHT22_queue
LIGHT_QUEUE_NAME=LightSensor_queue
PIR_QUEUE_NAME=PIR_queue

# WebSocket URIs (for local testing, these defaults are fine)
PZEM_WEBSOCKET_URI=ws://localhost:8081/ws?topic=pzem&emitter=true
DHT22_WEBSOCKET_URI=ws://localhost:8081/ws?topic=dht22&emitter=true
LIGHT_WEBSOCKET_URI=ws://localhost:8081/ws?topic=light_sensor&emitter=true
PIR_WEBSOCKET_URI=ws://localhost:8081/ws?topic=pir&emitter=true

# Notification Service (optional for testing)
API_WEBHOOK_URL=http://localhost:8000/api/internal/notifications/service
API_WEBHOOK_TOKEN=your-webhook-token-here

# WebSocket Server Configuration
WEBSOCKET_PORT=8081
```

### 3.2 Create Automation Engine .env File

```bash
# Copy the example file
cp Backend/automation-engine/.env.example Backend/automation-engine/.env
```

Edit `Backend/automation-engine/.env`:

```bash
# RabbitMQ Configuration
RABBITMQ_URI=amqp://guest:guest@localhost:5672/

# PostgreSQL Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=voltio_user
POSTGRES_PASSWORD=your_secure_password  # From PostgreSQL setup
POSTGRES_DB=voltiodb
```

## 📦 Step 4: Install Go Dependencies

This will download all required Go packages for each service:

```bash
# Install dependencies for test producers
cd Backend/test_producers
go mod tidy
cd ../..

# Install dependencies for consumer services
cd Backend/DHT22_ConsumerSender
go mod tidy
cd ../..

cd Backend/PZEM_ConsumerSender
go mod tidy
cd ../..

cd Backend/PIR_ConsumerSender
go mod tidy
cd ../..

cd Backend/LightSensor_ConsumerSender
go mod tidy
cd ../..

cd Backend/Notification_ConsumerSender
go mod tidy
cd ../..

# Install dependencies for WebSocket server
cd Backend/WebSocketServer
go mod tidy
cd ../..

# Install dependencies for automation engine
cd Backend/automation-engine
go mod tidy
cd ../..
```

**Note:** This step might take a few minutes as Go downloads all dependencies.

## 🚀 Step 5: Start the Services

### Option A: Using PowerShell Script (Windows - Recommended)

```powershell
# This will open multiple PowerShell windows, one for each service
.\start_all_voltio_services.ps1
```

### Option B: Manual Start (All Platforms)

Open separate terminal windows for each service:

**Terminal 1 - WebSocket Server:**
```bash
cd Backend/WebSocketServer
go run cmd/main.go
```

**Terminal 2 - Automation Engine:**
```bash
cd Backend/automation-engine
go run main.go
```

**Terminal 3 - DHT22 Consumer:**
```bash
cd Backend/DHT22_ConsumerSender
go run middleware/RabbitToSocketMiddleware.go
```

**Terminal 4 - PZEM Consumer:**
```bash
cd Backend/PZEM_ConsumerSender
go run middleware/RabbitToSocketMiddleware.go
```

**Terminal 5 - PIR Consumer:**
```bash
cd Backend/PIR_ConsumerSender
go run middleware/RabbitToSocketMiddleware.go
```

**Terminal 6 - Light Sensor Consumer:**
```bash
cd Backend/LightSensor_ConsumerSender
go run middleware/RabbitToSocketMiddleware.go
```

**Terminal 7 - Notification Consumer:**
```bash
cd Backend/Notification_ConsumerSender
go run middleware/RabbitToSocketMiddleware.go
```

### What You Should See

Each service should output logs showing:
- ✅ Connection to RabbitMQ successful
- ✅ Connection to InfluxDB successful (for consumers)
- ✅ Listening for messages
- ✅ WebSocket connection ready (for consumers)

If you see errors, check:
1. Are RabbitMQ and InfluxDB running?
2. Are your `.env` files configured correctly?
3. Did you save the `.env` files after editing?

## 🧪 Step 6: Test with Simulated Sensors

Now that all services are running, let's simulate IoT sensors sending data.

Open **four more terminal windows** for the test producers:

**Terminal 8 - DHT22 Test Producer:**
```bash
cd Backend/test_producers/dht22
go run main.go
```

**Terminal 9 - PZEM Test Producer:**
```bash
cd Backend/test_producers/pzem
go run main.go
```

**Terminal 10 - PIR Test Producer:**
```bash
cd Backend/test_producers/pir
go run main.go
```

**Terminal 11 - Light Sensor Test Producer:**
```bash
cd Backend/test_producers/light
go run main.go
```

### What You Should See

**In Producer Windows:**
- 📤 Sending temperature/humidity data
- 📤 Sending electrical measurements
- 📤 Sending motion detection events
- 📤 Sending light level readings

**In Consumer Windows:**
- 📥 Receiving messages from RabbitMQ
- 💾 Writing to InfluxDB
- 📡 Forwarding to WebSocket clients

## ✅ Step 7: Verify Everything is Working

### Check RabbitMQ

1. Open http://localhost:15672
2. Login with `guest` / `guest`
3. Click on "Queues" tab
4. You should see queues like:
   - `DHT22_queue`
   - `PZEM_queue`
   - `PIR_queue`
   - `LightSensor_queue`
   - `alerts-queue`
5. Messages should be flowing (check "Message rates" column)

### Check InfluxDB

1. Open http://localhost:8086
2. Login with your credentials
3. Click "Data Explorer" (left sidebar)
4. Select bucket: `sensores`
5. You should see measurements like:
   - `dht22` (temperature, humidity)
   - `pzem` (voltage, current, power)
   - `pir` (motion events)
   - `light_sensor` (light levels)

### Check WebSocket Server

You can test WebSocket connections using a browser's developer console:

```javascript
// Open browser console (F12) and run:
const ws = new WebSocket('ws://localhost:8081/ws?topic=dht22');
ws.onmessage = (event) => {
    console.log('Received:', JSON.parse(event.data));
};
ws.onopen = () => console.log('Connected to WebSocket');
ws.onerror = (error) => console.log('Error:', error);
```

You should see sensor data streaming in real-time!

## 🎉 Success!

Congratulations! You now have the complete Voltio CHMMA system running locally. Here's what you've accomplished:

✅ Installed all required dependencies  
✅ Configured the system with environment variables  
✅ Started all microservices  
✅ Simulated IoT sensors sending data  
✅ Verified data flows through the entire system  

## 🔍 What's Next?

Now that you have the system running, you can:

1. **Explore the Code**
   - Look at how producers generate sensor data
   - Understand how consumers process messages
   - Study the WebSocket server implementation

2. **Modify Test Data**
   - Change sensor values in test producers
   - Adjust sending frequencies
   - Add new sensor devices

3. **Build Something New**
   - Add a new sensor type
   - Create a custom consumer
   - Build a web dashboard to visualize data

4. **Read More Documentation**
   - [README.md](./README.md) - Complete project overview
   - [CONTRIBUTING.md](./CONTRIBUTING.md) - Contribution guidelines
   - [ENVIRONMENT_SETUP.md](./ENVIRONMENT_SETUP.md) - Detailed configuration

## 🐛 Troubleshooting

### "Failed to connect to RabbitMQ"

**Problem:** Service can't connect to RabbitMQ  
**Solution:**
1. Check RabbitMQ is running: http://localhost:15672
2. Verify `RABBITMQ_URI` in `.env` file
3. Try default credentials: `amqp://guest:guest@localhost:5672/`

### "InfluxDB authentication failed"

**Problem:** Invalid or missing InfluxDB token  
**Solution:**
1. Open InfluxDB UI: http://localhost:8086
2. Generate a new API token: Settings → Tokens → Generate Token
3. Copy token to `INFLUX_TOKEN` in `.env`
4. Restart consumer services

### "Queue not found" or "Queue declaration failed"

**Problem:** RabbitMQ queue doesn't exist  
**Solution:**
- Producers automatically create queues when they start
- Make sure at least one producer is running
- Check RabbitMQ management UI to verify queue exists

### "Address already in use" (Port 8081)

**Problem:** WebSocket server port is already taken  
**Solution:**
1. Stop any other process using port 8081
2. Or change `WEBSOCKET_PORT` in `.env` to a different port (e.g., 8082)
3. Update consumer WebSocket URIs to match new port

### Service compiles but doesn't start

**Problem:** Missing or incorrect environment variables  
**Solution:**
1. Verify `.env` files exist in correct locations
2. Check all required variables are set
3. Review [ENVIRONMENT_SETUP.md](./ENVIRONMENT_SETUP.md) for requirements
4. Restart terminal/shell to pick up new environment variables

## 💬 Need Help?

If you're stuck:

1. **Check the logs** - Error messages usually point to the problem
2. **Review documentation** - README.md and ENVIRONMENT_SETUP.md
3. **Search issues** - Someone may have had the same problem
4. **Ask for help** - Open an issue on GitHub

## 📚 Learning Resources

- **Go Basics:** https://tour.golang.org/
- **RabbitMQ Tutorials:** https://www.rabbitmq.com/getstarted.html
- **InfluxDB Guide:** https://docs.influxdata.com/influxdb/v2/get-started/
- **WebSocket Basics:** https://javascript.info/websocket

---

**Welcome to the Voltio CHMMA community! Happy coding! 🚀**
