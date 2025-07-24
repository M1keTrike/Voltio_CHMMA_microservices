// light_producer.go - Productor de prueba para sensor de luz
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
	exchange  = "amq.topic"
	routingKey = "light.data.events"
)

type LightSensorMessage struct {
	MAC        string  `json:"mac"`
	LightLevel float64 `json:"light_level"`
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

	log.Printf("✅ Conectado a RabbitMQ - Exchange: %s, RoutingKey: %s", exchange, routingKey)
	log.Println("🔄 Publicando datos cada 10 segundos...")

	// Simular diferentes sensores de luz
	sensors := []string{
		"CC:DB:A7:2F:AE:B0",
	}

	rand.Seed(time.Now().UnixNano())

	for {
		for _, sensorMAC := range sensors {
			// Generar nivel de luz aleatorio (0-1000 lux)
			lightLevel := rand.Float64() * 1000.0

			msg := LightSensorMessage{
				MAC:        sensorMAC,
				LightLevel: lightLevel,
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
				log.Printf("📤 [%s] Luz: %.1f lux", sensorMAC, lightLevel)
			}
		}

		time.Sleep(10 * time.Second)
	}
}
