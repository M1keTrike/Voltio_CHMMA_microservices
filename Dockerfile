# Voltio Complete Services Dockerfile
# Este Dockerfile construye y ejecuta todos los consumers y el WebSocket server

# ============================================
# Stage 1: Build all Go applications
# ============================================
FROM golang:1.23.4-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy all go.mod and go.sum files first for better caching
COPY Backend/PZEM_ConsumerSender/go.mod Backend/PZEM_ConsumerSender/go.sum ./PZEM_ConsumerSender/
COPY Backend/WebSocketServer/go.mod Backend/WebSocketServer/go.sum ./WebSocketServer/
COPY Backend/test_producers/go.mod Backend/test_producers/go.sum ./test_producers/
COPY Backend/automation-engine/go.mod Backend/automation-engine/go.sum ./automation-engine/
COPY Backend/DHT22_ConsumerSender/go.mod Backend/DHT22_ConsumerSender/go.sum ./DHT22_ConsumerSender/
COPY Backend/LightSensor_ConsumerSender/go.mod Backend/LightSensor_ConsumerSender/go.sum ./LightSensor_ConsumerSender/
COPY Backend/Notification_ConsumerSender/go.mod Backend/Notification_ConsumerSender/go.sum ./Notification_ConsumerSender/
COPY Backend/PIR_ConsumerSender/go.mod Backend/PIR_ConsumerSender/go.sum ./PIR_ConsumerSender/

# Download dependencies for all modules
RUN cd PZEM_ConsumerSender && go mod download
RUN cd WebSocketServer && go mod download
RUN cd automation-engine && go mod download
RUN cd DHT22_ConsumerSender && go mod download
RUN cd LightSensor_ConsumerSender && go mod download
RUN cd Notification_ConsumerSender && go mod download
RUN cd PIR_ConsumerSender && go mod download

# Copy source code
COPY Backend/ ./

# Build all applications
RUN cd PZEM_ConsumerSender/consumers/pzem_consumer && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/bin/pzem-consumer .
RUN cd PZEM_ConsumerSender/consumers/dht22-consumer && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/bin/dht22-consumer .
RUN cd PZEM_ConsumerSender/consumers/light-sensor-consumer && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/bin/light-sensor-consumer .
RUN cd PZEM_ConsumerSender/consumers/pir-consumer && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/bin/pir-consumer .
RUN cd PZEM_ConsumerSender/consumers/notification-consumer && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/bin/notification-consumer .
RUN cd WebSocketServer/cmd && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/bin/websocket-server .

# ============================================
# Stage 2: Create final runtime image
# ============================================
FROM alpine:latest

# Install ca-certificates for HTTPS requests and supervisor for process management
RUN apk --no-cache add ca-certificates supervisor

WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /build/bin/* ./

# Create supervisor configuration
RUN mkdir -p /etc/supervisor/conf.d /var/log/supervisor

# Create supervisor configuration file
COPY <<EOF /etc/supervisor/conf.d/supervisord.conf
[supervisord]
nodaemon=true
user=root
logfile=/var/log/supervisor/supervisord.log
pidfile=/var/run/supervisord.pid

[program:websocket-server]
command=/app/websocket-server
autostart=true
autorestart=true
stderr_logfile=/var/log/supervisor/websocket-server.err.log
stdout_logfile=/var/log/supervisor/websocket-server.out.log
environment=PORT=8081

[program:notification-consumer]
command=/app/notification-consumer
autostart=true
autorestart=true
stderr_logfile=/var/log/supervisor/notification-consumer.err.log
stdout_logfile=/var/log/supervisor/notification-consumer.out.log
environment=RABBITMQ_URI="%(ENV_RABBITMQ_URI)s",ALERTS_QUEUE_NAME="%(ENV_ALERTS_QUEUE_NAME)s",API_WEBHOOK_URL="%(ENV_API_WEBHOOK_URL)s",API_WEBHOOK_TOKEN="%(ENV_API_WEBHOOK_TOKEN)s"

[program:pzem-consumer]
command=/app/pzem-consumer
autostart=true
autorestart=true
stderr_logfile=/var/log/supervisor/pzem-consumer.err.log
stdout_logfile=/var/log/supervisor/pzem-consumer.out.log
environment=RABBITMQ_URI="%(ENV_RABBITMQ_URI)s",QUEUE_NAME="%(ENV_PZEM_QUEUE_NAME)s",WEBSOCKET_URI="%(ENV_PZEM_WEBSOCKET_URI)s",INFLUX_URL="%(ENV_INFLUX_URL)s",INFLUX_TOKEN="%(ENV_INFLUX_TOKEN)s",INFLUX_ORG="%(ENV_INFLUX_ORG)s",INFLUX_BUCKET="%(ENV_INFLUX_BUCKET)s",CONSUMER_TYPE="pzem",TOPIC_NAME="pzem",TIMEOUT_SECONDS="300",ALERTS_QUEUE_NAME="%(ENV_ALERTS_QUEUE_NAME)s"

[program:dht22-consumer]
command=/app/dht22-consumer
autostart=true
autorestart=true
stderr_logfile=/var/log/supervisor/dht22-consumer.err.log
stdout_logfile=/var/log/supervisor/dht22-consumer.out.log
environment=RABBITMQ_URI="%(ENV_RABBITMQ_URI)s",QUEUE_NAME="%(ENV_DHT22_QUEUE_NAME)s",WEBSOCKET_URI="%(ENV_DHT22_WEBSOCKET_URI)s",INFLUX_URL="%(ENV_INFLUX_URL)s",INFLUX_TOKEN="%(ENV_INFLUX_TOKEN)s",INFLUX_ORG="%(ENV_INFLUX_ORG)s",INFLUX_BUCKET="%(ENV_INFLUX_BUCKET)s",CONSUMER_TYPE="dht22",TOPIC_NAME="dht22",TIMEOUT_SECONDS="300",ALERTS_QUEUE_NAME="%(ENV_ALERTS_QUEUE_NAME)s"

[program:light-sensor-consumer]
command=/app/light-sensor-consumer
autostart=true
autorestart=true
stderr_logfile=/var/log/supervisor/light-sensor-consumer.err.log
stdout_logfile=/var/log/supervisor/light-sensor-consumer.out.log
environment=RABBITMQ_URI="%(ENV_RABBITMQ_URI)s",QUEUE_NAME="%(ENV_LIGHT_QUEUE_NAME)s",WEBSOCKET_URI="%(ENV_LIGHT_WEBSOCKET_URI)s",INFLUX_URL="%(ENV_INFLUX_URL)s",INFLUX_TOKEN="%(ENV_INFLUX_TOKEN)s",INFLUX_ORG="%(ENV_INFLUX_ORG)s",INFLUX_BUCKET="%(ENV_INFLUX_BUCKET)s",CONSUMER_TYPE="light-sensor",TOPIC_NAME="light",TIMEOUT_SECONDS="300",ALERTS_QUEUE_NAME="%(ENV_ALERTS_QUEUE_NAME)s"

[program:pir-consumer]
command=/app/pir-consumer
autostart=true
autorestart=true
stderr_logfile=/var/log/supervisor/pir-consumer.err.log
stdout_logfile=/var/log/supervisor/pir-consumer.out.log
environment=RABBITMQ_URI="%(ENV_RABBITMQ_URI)s",QUEUE_NAME="%(ENV_PIR_QUEUE_NAME)s",WEBSOCKET_URI="%(ENV_PIR_WEBSOCKET_URI)s",INFLUX_URL="%(ENV_INFLUX_URL)s",INFLUX_TOKEN="%(ENV_INFLUX_TOKEN)s",INFLUX_ORG="%(ENV_INFLUX_ORG)s",INFLUX_BUCKET="%(ENV_INFLUX_BUCKET)s",CONSUMER_TYPE="pir",TOPIC_NAME="pir",TIMEOUT_SECONDS="300",ALERTS_QUEUE_NAME="%(ENV_ALERTS_QUEUE_NAME)s"
EOF

# Expose WebSocket server port
EXPOSE 8081

# Create health check script
COPY <<EOF /app/healthcheck.sh
#!/bin/sh
# Check if all processes are running
supervisorctl status | grep -q "RUNNING" || exit 1
# Check if WebSocket server is responding
wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1
exit 0
EOF

RUN chmod +x /app/healthcheck.sh

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD /app/healthcheck.sh

# Start supervisor
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
