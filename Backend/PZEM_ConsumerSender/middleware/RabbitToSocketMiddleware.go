package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/rabbitmq/amqp091-go"
)

// --- CONFIGURACIÓN ---
const (
	// RabbitMQ
	amqpURI    = "amqp://admin:trike@52.73.74.139:5672/" // Tu URL de RabbitMQ
	queueName  = "PZEM_queue"                            // Nombre de tu cola de RabbitMQ

	// WebSocket
	wsURI = "wss://websocketvoltio.acstree.xyz/ws?topic=pzem&emitter=true" // Tu endpoint WebSocket
)

// --- ESTRUCTURAS DE DATOS ---

// Estructura para decodificar el mensaje que viene de RabbitMQ
// Solo nos interesa el payload y, dentro de él, la MAC.
type RabbitMessage struct {
	Payload struct {
		MAC string `json:"mac"`
	} `json:"payload"`
}

// Estructura para el objeto que irá DENTRO del campo "content"
type ContentPayload struct {
	MAC     string `json:"mac"`
	Message string `json:"message"` // El JSON original completo como un string
}

// Estructura final del mensaje que se enviará al WebSocket
type WebSocketMessage struct {
	Content string `json:"content"`
}

// Función de utilidad para manejar errores
func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	// --- CONEXIÓN A RABBITMQ ---
	log.Println("Conectando a RabbitMQ...")
	conn, err := amqp091.Dial(amqpURI)
	failOnError(err, "Fallo al conectar con RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Fallo al abrir un canal")
	defer ch.Close()

	// Declaramos la cola por si no existe
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	failOnError(err, "Fallo al declarar la cola")
	log.Printf("Conectado a RabbitMQ, escuchando en la cola '%s'", queueName)

	// --- CONEXIÓN AL WEBSOCKET ---
	log.Println("Conectando a WebSocket...")
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURI, nil)
	failOnError(err, "Fallo al conectar con el WebSocket")
	defer wsConn.Close()
	log.Printf("Conectado al WebSocket en %s", wsURI)

	// --- CONSUMIDOR RABBITMQ ---
	msgs, err := ch.Consume(
		queueName,
		"",      // consumer
		false,   // auto-ack -> MUY IMPORTANTE, confirmaremos manualmente
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	failOnError(err, "Fallo al registrar un consumidor")

	// Bucle para procesar mensajes
	log.Println(" [*] Esperando mensajes. Para salir, presiona CTRL+C")
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("--> Mensaje recibido de RabbitMQ: %s", d.Body)

			// 1. Decodificar el mensaje de RabbitMQ para obtener la MAC
			var rabbitMsg RabbitMessage
			err := json.Unmarshal(d.Body, &rabbitMsg)
			if err != nil {
				log.Printf("Error al decodificar el JSON de RabbitMQ: %v. Descartando mensaje.", err)
				d.Nack(false, false) // Nack para descartar el mensaje malformado
				continue
			}

			// 2. Construir el payload para el campo "content"
			contentPayload := ContentPayload{
				MAC:     rabbitMsg.Payload.MAC,
				Message: string(d.Body), // El mensaje original como string
			}
			
			// 3. Convertir ese payload a un string JSON
			contentBytes, err := json.Marshal(contentPayload)
			if err != nil {
				log.Printf("Error al codificar el content payload: %v. Descartando mensaje.", err)
				d.Nack(false, false)
				continue
			}

			// 4. Construir el mensaje final para el WebSocket
			wsMsg := WebSocketMessage{
				Content: string(contentBytes),
			}

			// 5. Enviar el mensaje al WebSocket
			log.Printf("<-- Enviando mensaje a WebSocket: %v", wsMsg)
			err = wsConn.WriteJSON(wsMsg)
			if err != nil {
				log.Printf("!!! Error al escribir en el WebSocket: %v. El programa terminará.", err)
                // En un sistema real, aquí implementarías una lógica de reconexión al WebSocket.
				close(forever) // Termina el programa si falla la conexión WS
				return
			}
			
			// 6. Confirmar que el mensaje fue procesado correctamente
			d.Ack(false)
		}
	}()

	<-forever // Bloquea el programa para que siga corriendo
}