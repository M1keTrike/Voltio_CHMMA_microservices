// RabbitToSocketMiddleware.go (Versión final)

package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/rabbitmq/amqp091-go"
)

// --- CONFIGURACIÓN ---
const (
	amqpURI   = "amqp://admin:trike@52.73.74.139:5672/"
	queueName = "PZEM_queue"
	wsURI     = "wss://websocketvoltio.acstree.xyz/ws?topic=pzem&emitter=true"

	// <<< CONFIGURACIÓN DE INFLUXDB ACTUALIZADA >>>
	influxURL    = "http://52.201.107.193:8086" // IP externa de tu InfluxDB
	influxToken  = "lJLzxtHLHvPNgdvU9dcInGYb/qLbLxUPgrePzLd47EKCLUWBzJ+RmJkpH0f1HkmQ"
	influxOrg    = "mi-org"
	influxBucket = "sensores"
)

// --- ESTRUCTURAS ---
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
type ContentPayload struct {
	MAC     string `json:"mac"`
	Message string `json:"message"`
}
type WebSocketMessage struct {
	Content string `json:"content"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	log.Println("Iniciando PZEM Middleware v2.0 (WS + InfluxDB)...")

	// --- CONEXIONES INICIALES ---
	log.Println("Conectando a RabbitMQ...")
	rabbitConn, err := amqp091.Dial(amqpURI)
	failOnError(err, "Fallo al conectar con RabbitMQ")
	defer rabbitConn.Close()
	ch, err := rabbitConn.Channel()
	failOnError(err, "Fallo al abrir un canal")
	defer ch.Close()
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	failOnError(err, "Fallo al declarar la cola")

	log.Println("Conectando a WebSocket...")
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURI, nil)
	failOnError(err, "Fallo al conectar con el WebSocket")
	defer wsConn.Close()

	log.Println("Configurando cliente de InfluxDB...")
	influxClient := influxdb2.NewClient(influxURL, influxToken)
	defer influxClient.Close()
	writeAPI := influxClient.WriteAPIBlocking(influxOrg, influxBucket)

	log.Println("--- Todas las conexiones listas. Esperando mensajes. ---")

	// --- CONSUMIDOR ---
	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	failOnError(err, "Fallo al registrar un consumidor")

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			log.Printf("--> Mensaje recibido: %s", string(d.Body)[:100]) // Log truncado

			var pzemMsg PZEMMessage
			if err := json.Unmarshal(d.Body, &pzemMsg); err != nil {
				log.Printf("Error al decodificar JSON: %v. Descartando.", err)
				d.Nack(false, false)
				continue
			}

			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				contentPayload := ContentPayload{MAC: pzemMsg.Payload.MAC, Message: string(d.Body)}
				contentBytes, _ := json.Marshal(contentPayload)
				wsMsg := WebSocketMessage{Content: string(contentBytes)}
				if err := wsConn.WriteJSON(wsMsg); err != nil {
					log.Printf("!!! [WS] Error: %v", err)
				} else {
					log.Println("OK! [WS] Mensaje enviado.")
				}
			}()

			go func() {
				defer wg.Done()
				p := influxdb2.NewPoint("energy_metrics",
					map[string]string{"deviceId": pzemMsg.DeviceID, "mac": pzemMsg.Payload.MAC},
					map[string]interface{}{
						"voltage": pzemMsg.Payload.Voltage, "current": pzemMsg.Payload.Current,
						"power": pzemMsg.Payload.Power, "energy": pzemMsg.Payload.Energy,
						"frequency": pzemMsg.Payload.Frequency, "powerFactor": pzemMsg.Payload.PowerFactor,
					},
					time.Now(),
				)
				if err := writeAPI.WritePoint(context.Background(), p); err != nil {
					log.Printf("!!! [InfluxDB] Error: %v", err)
				} else {
					log.Println("OK! [InfluxDB] Punto escrito.")
				}
			}()

			wg.Wait()
			d.Ack(false)
			log.Println("--- Mensaje procesado y confirmado ---")
		}
	}()
	<-forever
}
