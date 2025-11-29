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

# Compilar aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o websocket-server ./cmd/main.go

# ==============================================================================
# Imagen final
# ==============================================================================
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

# Copiar binario compilado
COPY --from=builder /app/websocket-server .

# Exponer puerto
EXPOSE 8081

# Ejecutar
CMD ["./websocket-server"]
