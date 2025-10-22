// DHT22 Producer - Generador de datos de temperatura y humedad
package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const (
	queueName = "DHT22_queue"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type DHT22Message struct {
	DeviceID string `json:"deviceId"`
	Payload  struct {
		MAC         string  `json:"mac"`
		Temperature float64 `json:"temperature"`
		Humidity    float64 `json:"humidity"`
	} `json:"payload"`
}

func main() {
	log.Println("🌡️ Iniciando DHT22 Producer de Prueba...")

	// Conectar a RabbitMQ
	amqpURI := getEnv("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/")
	conn, err := amqp091.Dial(amqpURI)
	if err != nil {
		log.Fatalf("Error conectando a RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error abriendo canal: %v", err)
	}
	defer ch.Close()

	// Declarar cola
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Error declarando cola: %v", err)
	}

	log.Printf("✅ Conectado a RabbitMQ - Cola: %s", queueName)
	log.Println("🔄 Publicando datos cada 30 segundos...")

	// Simular diferentes sensores DHT22
	devices := []struct {
		DeviceID string
		MAC      string
	}{
		{"DHT22-DEV-001", "DHT22-001"},
		{"DHT22-DEV-002", "DHT22-002"},
		{"DHT22-DEV-003", "DHT22-003"},
	}

	rand.Seed(time.Now().UnixNano())

	for {
		for _, device := range devices {
			// Generar datos realistas de temperatura y humedad
			baseTemp := 20.0 + rand.Float64()*15.0     // 20-35°C
			baseHumidity := 40.0 + rand.Float64()*30.0 // 40-70%

			// Agregar variación temporal (simulando día/noche)
			hour := time.Now().Hour()
			tempVariation := 5.0 * ((float64(hour) - 12.0) / 12.0)
			temperature := baseTemp + tempVariation + (rand.Float64()-0.5)*2.0

			// La humedad tiende a ser inversa a la temperatura
			humidity := baseHumidity - (temperature-25.0)*0.5 + (rand.Float64()-0.5)*5.0

			msg := DHT22Message{
				DeviceID: device.DeviceID,
			}
			msg.Payload.MAC = device.MAC
			msg.Payload.Temperature = temperature
			msg.Payload.Humidity = humidity

			msgBytes, err := json.Marshal(msg)
			if err != nil {
				log.Printf("❌ Error marshaling mensaje: %v", err)
				continue
			}

			err = ch.Publish(
				"",        // exchange
				queueName, // routing key
				false,     // mandatory
				false,     // immediate
				amqp091.Publishing{
					ContentType: "application/json",
					Body:        msgBytes,
				},
			)

			if err != nil {
				log.Printf("❌ Error publicando mensaje: %v", err)
			} else {
				log.Printf("📤 [%s] %.1f°C, %.1f%%",
					device.MAC, temperature, humidity)
			}
		}

		time.Sleep(30 * time.Second)
	}
}
