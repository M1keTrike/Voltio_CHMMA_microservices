// Light Sensor Producer - Generador de datos de sensor de luz
package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const (
	amqpURI   = "amqp://admin:trike@52.73.74.139:5672/"
	queueName = "LightSensor_queue"
)

type LightMessage struct {
	DeviceID string `json:"deviceId"`
	Payload  struct {
		MAC        string  `json:"mac"`
		LightLevel float64 `json:"lightLevel"`
	} `json:"payload"`
}

func main() {
	log.Println("💡 Iniciando Light Sensor Producer de Prueba...")

	// Conectar a RabbitMQ
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
	log.Println("🔄 Publicando datos cada 15 segundos...")

	// Simular diferentes sensores de luz
	devices := []struct {
			DeviceID string
			MAC      string
			Location string
		}{
			{"LIGHT-DEV-001", "CC:DB:A7:2F:AE:B0", "Sala"},
			{"LIGHT-DEV-002", "CC:DB:A7:2F:AE:B0", "Cocina"},
			{"LIGHT-DEV-003", "CC:DB:A7:2F:AE:B0", "Oficina"},
			{"LIGHT-DEV-004", "CC:DB:A7:2F:AE:B0", "Exterior"},
	}

	rand.Seed(time.Now().UnixNano())

	for {
		for _, device := range devices {
			var lightLevel float64

			// Simular patrones de luz realistas según la hora
			hour := time.Now().Hour()

			if device.Location == "Exterior" {
				// Luz exterior - sigue patrón solar
				if hour >= 6 && hour <= 18 {
					// Día: luz alta con variaciones
					baseLux := 50000.0 // Lux base para día
					if hour >= 11 && hour <= 15 {
						baseLux = 100000.0 // Mediodía más brillante
					}
					lightLevel = baseLux + (rand.Float64()-0.5)*20000.0
				} else {
					// Noche: luz muy baja (luna, farolas)
					lightLevel = rand.Float64() * 50.0
				}
			} else {
				// Luz interior - depende de actividad humana
				if hour >= 7 && hour <= 23 {
					// Horario activo: luces artificiales
					lightLevel = 300.0 + rand.Float64()*700.0 // 300-1000 lux
				} else {
					// Horario de descanso: luces apagadas o tenues
					lightLevel = rand.Float64() * 50.0 // 0-50 lux
				}
			}

			// Agregar ruido aleatorio pequeño
			lightLevel += (rand.Float64() - 0.5) * 10.0
			if lightLevel < 0 {
				lightLevel = 0
			}

			msg := LightMessage{
				DeviceID: device.DeviceID,
			}
			msg.Payload.MAC = device.MAC
			msg.Payload.LightLevel = lightLevel

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
				log.Printf("📤 [%s - %s] %.0f lux",
					device.MAC, device.Location, lightLevel)
			}
		}

		time.Sleep(15 * time.Second)
	}
}
