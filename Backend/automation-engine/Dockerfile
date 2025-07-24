# Dockerfile para automation-engine
FROM golang:1.23-alpine
WORKDIR /app
COPY . .
RUN go build -o automation-engine .
CMD ["./automation-engine"]
