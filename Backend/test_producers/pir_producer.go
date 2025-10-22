// pir_producer.go - Productor de prueba para sensor PIR
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
	exchange   = "amq.topic"
	routingKey = "pir.data.events"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type PIRMessage struct {
	MAC            string `json:"mac"`
	MotionDetected bool   `json:"motion_detected"`
}

func main() {
	log.Println("🚶 Iniciando PIR Sensor Producer de Prueba...")

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

	log.Printf("✅ Conectado a RabbitMQ - Exchange: %s, RoutingKey: %s", exchange, routingKey)
	log.Println("🔄 Publicando datos cada 10 segundos...")

	// Simular diferentes sensores PIR
	sensors := []string{
		"CC:DB:A7:2F:AE:B0",
	}

	rand.Seed(time.Now().UnixNano())

	for {
		for _, sensorMAC := range sensors {
			// Generar detección de movimiento aleatoria (70% sin movimiento, 30% con movimiento)
			motionDetected := rand.Float64() < 0.3

			msg := PIRMessage{
				MAC:            sensorMAC,
				MotionDetected: motionDetected,
			}

			msgBytes, err := json.Marshal(msg)
			if err != nil {
				log.Printf("❌ Error marshaling mensaje: %v", err)
				continue
			}

			err = ch.Publish(
				exchange,   // exchange
				routingKey, // routing key
				false,      // mandatory
				false,      // immediate
				amqp091.Publishing{
					ContentType: "application/json",
					Body:        msgBytes,
				},
			)

			if err != nil {
				log.Printf("❌ Error publicando mensaje: %v", err)
			} else {
				motionStatus := "❌ Sin movimiento"
				if motionDetected {
					motionStatus = "✅ Movimiento detectado"
				}
				log.Printf("📤 [%s] %s", sensorMAC, motionStatus)
			}
		}

		time.Sleep(10 * time.Second)
	}
}
