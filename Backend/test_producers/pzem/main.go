// PZEM Producer - Generador de datos de medidor eléctrico PZEM
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
	queueName = "PZEM_queue"
)

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
	log.Println("🔄 Publicando datos cada 10 segundos...")

	// Simular diferentes medidores PZEM
	devices := []struct {
		DeviceID    string
		MAC         string
		Location    string
		LoadType    string  // Tipo de carga eléctrica
		BaseLoad    float64 // Carga base en watts
		TotalEnergy float64 // Energía acumulada inicial
	}{
		{"PZEM-DEV-001", "CC:DB:A7:2F:AE:B0", "Casa Principal", "residential", 1500.0, 1000.0},
		{"PZEM-DEV-002", "CC:DB:A7:2F:AE:B0", "Aire Acondicionado", "hvac", 3000.0, 2500.0},
		{"PZEM-DEV-003", "CC:DB:A7:2F:AE:B0", "Iluminación", "lighting", 500.0, 800.0},
		{"PZEM-DEV-004", "CC:DB:A7:2F:AE:B0", "Oficina", "office", 800.0, 1200.0},
	}

	rand.Seed(time.Now().UnixNano())

	for {
		for i, device := range devices {
			// Generar voltaje con pequeñas variaciones (210-230V típico en México)
			voltage := 220.0 + (rand.Float64()-0.5)*15.0

			// Simular patrones de consumo según la hora y tipo de carga
			hour := time.Now().Hour()
			var loadFactor float64 = 1.0

			switch device.LoadType {
			case "residential":
				// Casa: mayor consumo en mañana y noche
				if hour >= 6 && hour <= 9 || hour >= 18 && hour <= 23 {
					loadFactor = 0.8 + rand.Float64()*0.4 // 80-120%
				} else {
					loadFactor = 0.3 + rand.Float64()*0.3 // 30-60%
				}
			case "hvac":
				// Aire acondicionado: mayor uso en horas calurosas
				if hour >= 12 && hour <= 18 {
					loadFactor = 0.9 + rand.Float64()*0.2 // 90-110%
				} else {
					loadFactor = 0.2 + rand.Float64()*0.3 // 20-50%
				}
			case "lighting":
				// Iluminación: mayor uso en la noche
				if hour >= 18 || hour <= 6 {
					loadFactor = 0.7 + rand.Float64()*0.3 // 70-100%
				} else {
					loadFactor = 0.1 + rand.Float64()*0.2 // 10-30%
				}
			case "office":
				// Oficina: uso durante horario laboral
				if hour >= 8 && hour <= 18 {
					loadFactor = 0.6 + rand.Float64()*0.4 // 60-100%
				} else {
					loadFactor = 0.1 + rand.Float64()*0.2 // 10-30%
				}
			}

			// Calcular potencia actual
			power := device.BaseLoad * loadFactor

			// Calcular corriente (P = V * I)
			current := power / voltage

			// Generar frecuencia típica (49.5-50.5 Hz)
			frequency := 50.0 + (rand.Float64()-0.5)*1.0

			// Factor de potencia típico (0.85-0.95 para cargas mixtas)
			powerFactor := 0.85 + rand.Float64()*0.1

			// Incrementar energía acumulada (kWh)
			energyIncrement := power * (10.0 / 3600.0 / 1000.0) // 10 segundos a kWh
			devices[i].TotalEnergy += energyIncrement

			msg := PZEMMessage{
				DeviceID: device.DeviceID,
			}
			msg.Payload.MAC = device.MAC
			msg.Payload.Voltage = voltage
			msg.Payload.Current = current
			msg.Payload.Power = power
			msg.Payload.Energy = devices[i].TotalEnergy
			msg.Payload.Frequency = frequency
			msg.Payload.PowerFactor = powerFactor

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
				// Formatear el log con información útil
				log.Printf("📤 [%s - %s] %.1fV, %.2fA, %.0fW, %.2fkWh (PF: %.2f)",
					device.MAC, device.Location, voltage, current, power,
					devices[i].TotalEnergy, powerFactor)
			}
		}

		time.Sleep(10 * time.Second)
	}
}
