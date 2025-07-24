# ===================================
# DOCKERFILE UNIFICADO BACKEND VOLTIO
# ===================================
# Contiene todos los servicios del backend:
# - PZEM Consumer
# - DHT22 Consumer  
# - PIR Consumer
# - Light Sensor Consumer
# - Notification Consumer
# - Automation Engine

# --- Etapa 1: Builder ---
FROM golang:1.23.4-alpine AS builder

# Instalar dependencias necesarias
RUN apk add --no-cache git

# Crear directorio de trabajo
WORKDIR /app

# === PZEM CONSUMER ===
WORKDIR /app/pzem
COPY PZEM_ConsumerSender/go.mod PZEM_ConsumerSender/go.sum ./
RUN go mod download
COPY PZEM_ConsumerSender/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /pzem-consumer ./middleware/RabbitToSocketMiddleware_NEW.go

# === DHT22 CONSUMER ===
WORKDIR /app/dht22
COPY DHT22_ConsumerSender/go.mod DHT22_ConsumerSender/go.sum ./
RUN go mod download
COPY DHT22_ConsumerSender/ ./
# Usar el archivo NEW si existe, sino el normal
RUN if [ -f ./middleware/RabbitToSocketMiddleware_NEW.go ]; then \
        CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /dht22-consumer ./middleware/RabbitToSocketMiddleware_NEW.go; \
    else \
        CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /dht22-consumer ./middleware/RabbitToSocketMiddleware.go; \
    fi

# === PIR CONSUMER ===
WORKDIR /app/pir
COPY PIR_ConsumerSender/go.mod PIR_ConsumerSender/go.sum ./
RUN go mod download
COPY PIR_ConsumerSender/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /pir-consumer ./middleware/RabbitToSocketMiddleware.go

# === LIGHT SENSOR CONSUMER ===
WORKDIR /app/light
COPY LightSensor_ConsumerSender/go.mod LightSensor_ConsumerSender/go.sum ./
RUN go mod download
COPY LightSensor_ConsumerSender/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /light-consumer ./middleware/RabbitToSocketMiddleware.go

# === NOTIFICATION CONSUMER ===
WORKDIR /app/notification
COPY Notification_ConsumerSender/go.mod Notification_ConsumerSender/go.sum ./
RUN go mod download
COPY Notification_ConsumerSender/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /notification-consumer ./middleware/RabbitToSocketMiddleware.go

# === AUTOMATION ENGINE ===
WORKDIR /app/automation
COPY automation-engine/go.mod automation-engine/go.sum ./
RUN go mod download
COPY automation-engine/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /automation-engine ./main.go

# --- Etapa 2: Final ---
FROM alpine:latest

# Instalar supervisor para gestionar múltiples procesos
RUN apk add --no-cache supervisor

# Crear directorio de trabajo
WORKDIR /app

# Copiar todos los binarios compilados
COPY --from=builder /pzem-consumer ./pzem-consumer
COPY --from=builder /dht22-consumer ./dht22-consumer
COPY --from=builder /pir-consumer ./pir-consumer
COPY --from=builder /light-consumer ./light-consumer
COPY --from=builder /notification-consumer ./notification-consumer
COPY --from=builder /automation-engine ./automation-engine

# Hacer todos los binarios ejecutables
RUN chmod +x ./pzem-consumer ./dht22-consumer ./pir-consumer ./light-consumer ./notification-consumer ./automation-engine

# Crear configuración de supervisor
RUN mkdir -p /etc/supervisor/conf.d

# Configuración de supervisor para todos los servicios
COPY <<EOF /etc/supervisor/conf.d/supervisord.conf
[supervisord]
nodaemon=true
user=root
logfile=/var/log/supervisor/supervisord.log
pidfile=/var/run/supervisord.pid

[program:pzem-consumer]
command=/app/pzem-consumer
directory=/app
autostart=true
autorestart=true
stderr_logfile=/var/log/pzem-consumer.err.log
stdout_logfile=/var/log/pzem-consumer.out.log
user=root
priority=100

[program:dht22-consumer]
command=/app/dht22-consumer
directory=/app
autostart=true
autorestart=true
stderr_logfile=/var/log/dht22-consumer.err.log
stdout_logfile=/var/log/dht22-consumer.out.log
user=root
priority=200

[program:pir-consumer]
command=/app/pir-consumer
directory=/app
autostart=true
autorestart=true
stderr_logfile=/var/log/pir-consumer.err.log
stdout_logfile=/var/log/pir-consumer.out.log
user=root
priority=300

[program:light-consumer]
command=/app/light-consumer
directory=/app
autostart=true
autorestart=true
stderr_logfile=/var/log/light-consumer.err.log
stdout_logfile=/var/log/light-consumer.out.log
user=root
priority=400

[program:notification-consumer]
command=/app/notification-consumer
directory=/app
autostart=true
autorestart=true
stderr_logfile=/var/log/notification-consumer.err.log
stdout_logfile=/var/log/notification-consumer.out.log
user=root
priority=500

[program:automation-engine]
command=/app/automation-engine
directory=/app
autostart=true
autorestart=true
stderr_logfile=/var/log/automation-engine.err.log
stdout_logfile=/var/log/automation-engine.out.log
user=root
priority=600
EOF

# Crear directorio de logs
RUN mkdir -p /var/log/supervisor

# Variables de entorno por defecto (se pueden sobreescribir)
ENV RABBITMQ_URI="amqp://guest:guest@rabbitmq:5672/"
ENV INFLUXDB_URL="http://influxdb:8086"
ENV INFLUXDB_TOKEN="lJLzxtHLHvPNgdvU9dcInGYb/qLbLxUPgrePzLd47EKCLUWBzJ+RmJkpH0f1HkmQ"
ENV INFLUXDB_ORG="mi-org"
ENV INFLUXDB_BUCKET="sensores"
ENV POSTGRES_HOST="13.222.89.227"
ENV POSTGRES_PORT="5432"
ENV POSTGRES_DB="voltiodb"
ENV POSTGRES_USER="chmma"
ENV POSTGRES_PASSWORD="HSQCx3Ajt4p^aJGC"

# Comando para iniciar supervisor
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
