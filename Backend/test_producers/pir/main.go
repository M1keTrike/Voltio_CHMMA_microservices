// PIR Motion Sensor Producer - Generador de datos de sensor de movimiento
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
	queueName = "PIR_queue"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type PIRMessage struct {
	DeviceID string `json:"deviceId"`
	Payload  struct {
		MAC            string `json:"mac"`
		MotionDetected bool   `json:"motionDetected"`
	} `json:"payload"`
}

func main() {
	log.Println("🚶 Iniciando PIR Motion Sensor Producer de Prueba...")

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
	log.Println("🔄 Publicando datos cada 20 segundos...")

	// Simular diferentes sensores PIR
	devices := []struct {
		DeviceID string
		MAC      string
		Location string
		Activity string // Nivel de actividad esperado
	}{
		{"PIR-DEV-001", "CC:DB:A7:2F:AE:B0", "Entrada Principal", "high"},
		{"PIR-DEV-002", "CC:DB:A7:2F:AE:B0", "Sala de Estar", "medium"},
		{"PIR-DEV-003", "CC:DB:A7:2F:AE:B0", "Cocina", "medium"},
		{"PIR-DEV-004", "CC:DB:A7:2F:AE:B0", "Pasillo", "low"},
		{"PIR-DEV-005", "CC:DB:A7:2F:AE:B0", "Baño", "low"},
		{"PIR-DEV-006", "CC:DB:A7:2F:AE:B0", "Exterior Trasero", "very_low"},
	}

	rand.Seed(time.Now().UnixNano())

	for {
		for _, device := range devices {
			var motionDetected bool

			// Simular patrones de movimiento realistas según hora y ubicación
			hour := time.Now().Hour()

			// Probabilidad base según la hora del día
			var baseProbability float64
			if hour >= 6 && hour <= 22 {
				// Horario activo
				baseProbability = 0.3
			} else {
				// Horario nocturno - menos actividad
				baseProbability = 0.05
			}

			// Ajustar probabilidad según el tipo de ubicación
			switch device.Activity {
			case "high":
				baseProbability *= 2.0
			case "medium":
				baseProbability *= 1.0
			case "low":
				baseProbability *= 0.5
			case "very_low":
				baseProbability *= 0.2
			}

			// Picos de actividad específicos
			if device.Location == "Cocina" && (hour == 7 || hour == 12 || hour == 19) {
				baseProbability *= 3.0 // Horas de comida
			}

			if device.Location == "Baño" && (hour >= 6 && hour <= 8) {
				baseProbability *= 2.0 // Rutina matutina
			}

			// Decidir si hay movimiento
			motionDetected = rand.Float64() < baseProbability

			msg := PIRMessage{
				DeviceID: device.DeviceID,
			}
			msg.Payload.MAC = device.MAC
			msg.Payload.MotionDetected = motionDetected

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
				status := "🟢 SIN MOVIMIENTO"
				if motionDetected {
					status = "🔴 MOVIMIENTO DETECTADO"
				}
				log.Printf("📤 [%s - %s] %s",
					device.MAC, device.Location, status)
			}
		}

		time.Sleep(20 * time.Second)
	}
}
