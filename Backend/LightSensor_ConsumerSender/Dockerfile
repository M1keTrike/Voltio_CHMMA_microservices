# Etapa 1: Compilación
# Utiliza una imagen oficial de Go como base para compilar la aplicación.
FROM golang:1.19-alpine AS builder

# Establece el directorio de trabajo dentro del contenedor.
WORKDIR /app

# Copia los archivos de gestión de dependencias.
COPY go.mod go.sum ./
# Descarga las dependencias.
RUN go mod download

# Copia todo el código fuente del proyecto.
COPY . .

# --- AJUSTE CLAVE AQUÍ ---
# Compila el archivo específico que contiene tu función main().
# El ejecutable se llamará 'main' y se guardará en la raíz del contenedor.
RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./middleware/RabbitToSocketMiddleware.go

# --- Etapa 2: Producción ---
# Utiliza una imagen base mínima y ligera para la ejecución.
FROM alpine:latest

# Copia el binario compilado de la etapa anterior.
COPY --from=builder /main /main

# (Opcional) Expone un puerto si tu aplicación es un servidor web.
# EXPOSE 8080

# Define el comando que se ejecutará al iniciar el contenedor.
# Como el binario se llama 'main', este comando no necesita cambios.
CMD ["/main"]