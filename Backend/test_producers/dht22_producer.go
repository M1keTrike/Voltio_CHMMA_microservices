// dht22_producer.go - Productor de prueba para sensor DHT22
package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const (
	amqpURI    = "amqp://admin:trike@52.73.74.139:5672/"
	exchange   = "amq.topic"
	routingKey = "dht22.data.events"
)

type DHT22Message struct {
	MAC         string  `json:"mac"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

func main() {
	log.Println("🌡️ Iniciando DHT22 Producer de Prueba...")

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

	// Simular diferentes sensores DHT22
	sensors := []string{
		"CC:DB:A7:2F:AE:B0",
	}

	rand.Seed(time.Now().UnixNano())

	for {
		for _, sensorMAC := range sensors {
			// Generar datos aleatorios realistas
			temperature := 15.0 + rand.Float64()*20.0 // 15-35°C
			humidity := 30.0 + rand.Float64()*40.0    // 30-70%

			msg := DHT22Message{
				MAC:         sensorMAC,
				Temperature: temperature,
				Humidity:    humidity,
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
				log.Printf("📤 [%s] Temp: %.1f°C, Hum: %.1f%%",
					sensorMAC, temperature, humidity)
			}
		}

		time.Sleep(10 * time.Second)
	}
}
