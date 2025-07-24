# --- Etapa 1: Builder ---
# Usamos una imagen oficial de Go como base para compilar nuestra aplicación.
# 'alpine' es una versión ligera que hace que esta etapa sea más rápida.
FROM golang:1.22-alpine AS builder

# Establecemos el directorio de trabajo dentro del contenedor.
WORKDIR /app

# Copiamos los archivos de dependencias.
# Esto aprovecha el caché de Docker: si no cambian, no se vuelven a descargar las dependencias.
COPY go.mod go.sum ./
RUN go mod download

# Copiamos el resto del código fuente de nuestra aplicación.
# Solo copiamos la carpeta 'middleware' que contiene nuestro código.
COPY middleware/ ./middleware/

# Compilamos la aplicación.
# CGO_ENABLED=0 crea un binario estático que no depende de librerías de C del sistema.
# -ldflags "-w -s" reduce el tamaño del binario eliminando información de depuración.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /pzem-middleware ./middleware/RabbitToSocketMiddleware.go


# --- Etapa 2: Final ---
# Usamos una imagen base mínima. Alpine es una excelente opción por su pequeño tamaño.
# ¡IMPORTANTE! Usamos 'alpine' en lugar de 'scratch' porque nuestra aplicación necesita
# certificados SSL/TLS para conectarse a 'wss://', y Alpine los incluye.
FROM alpine:latest

# Establecemos el directorio de trabajo.
WORKDIR /app

# Copiamos ÚNICAMENTE el binario compilado desde la etapa 'builder'.
# No copiamos el código fuente ni el compilador de Go.
COPY --from=builder /pzem-middleware .

# Establecemos el comando que se ejecutará cuando el contenedor se inicie.
# Simplemente corre nuestro binario.
CMD ["./pzem-middleware"]