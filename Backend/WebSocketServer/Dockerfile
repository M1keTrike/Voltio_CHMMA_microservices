# ==============================================================================
# Dockerfile - WebSocket Server
# ==============================================================================
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar aplicación con optimizaciones
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o websocket-server ./cmd/main.go

# ==============================================================================
# Imagen final
# ==============================================================================
FROM alpine:latest

# Instalar certificados CA y herramientas de red
RUN apk --no-cache add ca-certificates wget tzdata && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copiar binario compilado
COPY --from=builder /app/websocket-server .

# Cambiar ownership
RUN chown -R appuser:appuser /app

# Usar usuario no root
USER appuser

# Exponer puerto
EXPOSE 8081

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Ejecutar con variables de entorno optimizadas para Go
ENV GOGC=100 \
    GOMAXPROCS=0

CMD ["./websocket-server"]
