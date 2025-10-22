// pzem_producer.go - Productor de prueba para medidor PZEM
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
	routingKey = "pzem.data.events"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type PZEMMessage struct {
	DeviceID string `json:"deviceId"`
	Payload  struct {
		MAC         string  `json:"mac"`
		Voltage     float64 `json:"voltage"`
		Current     float64 `json:"current"`
		Power       float64 `json:"power"`
		Energy      float64 `json:"energy"`
		Frequency   float64 `json:"frequency"`
		PowerFactor float64 `json:"powerFactor"`
	} `json:"payload"`
}

func main() {
	log.Println("⚡ Iniciando PZEM Producer de Prueba...")

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

	// Simular diferentes medidores PZEM
	devices := []struct {
		DeviceID string
		MAC      string
	}{
		{"PZEM-DEV-001", "CC:DB:A7:2F:AE:B0"},
	}

	rand.Seed(time.Now().UnixNano())
	var totalEnergy float64 = 1000.0 // Energía acumulada inicial

	for {
		for _, device := range devices {
			// Generar datos eléctricos realistas
			voltage := 220.0 + (rand.Float64()-0.5)*10.0 // 215-225V
			current := rand.Float64() * 5.0              // 0-5A
			power := voltage * current                   // Potencia calculada
			frequency := 49.8 + rand.Float64()*0.4       // 49.8-50.2Hz
			powerFactor := 0.8 + rand.Float64()*0.2      // 0.8-1.0
			totalEnergy += power * (10.0 / 3600.0)       // Incrementar energía (10s en horas)

			msg := PZEMMessage{
				DeviceID: device.DeviceID,
			}
			msg.Payload.MAC = device.MAC
			msg.Payload.Voltage = voltage
			msg.Payload.Current = current
			msg.Payload.Power = power
			msg.Payload.Energy = totalEnergy
			msg.Payload.Frequency = frequency
			msg.Payload.PowerFactor = powerFactor

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
				log.Printf("📤 [%s] %.1fV, %.2fA, %.1fW, %.1fkWh",
					device.MAC, voltage, current, power, totalEnergy/1000.0)
			}
		}

		time.Sleep(10 * time.Second)
	}
}
