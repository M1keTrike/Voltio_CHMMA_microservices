// RabbitToSocketMiddleware.go - DHT22 Consumer con Sistema de Alertas
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rabbitmq/amqp091-go"
)

// --- CONFIGURACIÓN ---
const (
	amqpURI         = "amqp://admin:trike@52.73.74.139:5672/"
	queueName       = "DHT22_queue"
	alertsQueueName = "alerts-queue"
	wsURI           = "wss://websocketvoltio.acstree.xyz/ws?topic=dht22&emitter=true"

	// InfluxDB Configuration
	influxURL    = "http://52.201.107.193:8086"
	influxToken  = "lJLzxtHLHvPNgdvU9dcInGYb/qLbLxUPgrePzLd47EKCLUWBzJ+RmJkpH0f1HkmQ"
	influxOrg    = "mi-org"
	influxBucket = "sensores"

	// Timeout Configuration
	timeoutDuration = 2 * time.Minute
	checkInterval   = 30 * time.Second
)

// --- ESTRUCTURAS ---
type DHT22Message struct {
	DeviceID string `json:"deviceId"`
	Payload  struct {
		MAC         string  `json:"mac"`
		Temperature float64 `json:"temperature"`
		Humidity    float64 `json:"humidity"`
	} `json:"payload"`
}

type AlertMessage struct {
	SensorMAC  string    `json:"sensor_mac"`
	SensorType string    `json:"sensor_type"`
	AlertType  string    `json:"alert_type"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
	DeviceID   string    `json:"device_id,omitempty"`
}

type ContentPayload struct {
	MAC     string `json:"mac"`
	Message string `json:"message"`
}

type WebSocketMessage struct {
	Content string `json:"content"`
}

type DHT22Consumer struct {
	RabbitConn    *amqp091.Connection
	Channel       *amqp091.Channel
	AlertsChannel *amqp091.Channel
	WSConn        *websocket.Conn
	InfluxClient  influxdb2.Client
	WriteAPI      api.WriteAPIBlocking

	// Timeout tracking
	LastSeen      map[string]time.Time
	LastSeenMutex sync.RWMutex
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func NewDHT22Consumer() (*DHT22Consumer, error) {
	dc := &DHT22Consumer{
		LastSeen: make(map[string]time.Time),
	}

	if err := dc.setupConnections(); err != nil {
		return nil, err
	}

	// Start timeout checker goroutine
	go dc.timeoutChecker()

	return dc, nil
}

func (dc *DHT22Consumer) setupConnections() error {
	var err error

	// RabbitMQ Connection
	log.Println("Conectando a RabbitMQ...")
	dc.RabbitConn, err = amqp091.Dial(amqpURI)
	if err != nil {
		return fmt.Errorf("fallo al conectar con RabbitMQ: %v", err)
	}

	dc.Channel, err = dc.RabbitConn.Channel()
	if err != nil {
		return fmt.Errorf("fallo al abrir canal principal: %v", err)
	}

	// Separate channel for alerts
	dc.AlertsChannel, err = dc.RabbitConn.Channel()
	if err != nil {
		return fmt.Errorf("fallo al abrir canal de alertas: %v", err)
	}

	// Declarar colas
	_, err = dc.Channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("fallo al declarar cola DHT22: %v", err)
	}

	_, err = dc.AlertsChannel.QueueDeclare(alertsQueueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("fallo al declarar cola de alertas: %v", err)
	}

	// WebSocket Connection
	log.Println("Conectando a WebSocket...")
	dc.WSConn, _, err = websocket.DefaultDialer.Dial(wsURI, nil)
	if err != nil {
		return fmt.Errorf("fallo al conectar con WebSocket: %v", err)
	}

	// InfluxDB Connection
	log.Println("Configurando cliente de InfluxDB...")
	dc.InfluxClient = influxdb2.NewClient(influxURL, influxToken)
	dc.WriteAPI = dc.InfluxClient.WriteAPIBlocking(influxOrg, influxBucket)

	log.Println("✅ Todas las conexiones DHT22 listas")
	return nil
}

func (dc *DHT22Consumer) Start() error {
	log.Println("🌡️ Iniciando DHT22 Consumer con Alertas...")
	log.Printf("👂 Escuchando cola: %s", queueName)
	log.Printf("🚨 Cola de alertas: %s", alertsQueueName)

	msgs, err := dc.Channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("fallo al registrar consumidor: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("📨 [DHT22] Mensaje recibido: %s", string(d.Body))

			var dht22Msg DHT22Message
			if err := json.Unmarshal(d.Body, &dht22Msg); err != nil {
				log.Printf("❌ [DHT22] Error al decodificar JSON: %v", err)
				d.Nack(false, false)
				continue
			}

			// Procesar mensaje en paralelo
			if err := dc.processMessage(&dht22Msg, d.Body); err != nil {
				log.Printf("❌ [DHT22] Error procesando mensaje: %v", err)
				d.Nack(false, true) // Requeue para reintento
				continue
			}

			// Éxito
			d.Ack(false)
			log.Printf("✅ [DHT22] Mensaje procesado - MAC: %s, Temp: %.1f°C, Hum: %.1f%%",
				dht22Msg.Payload.MAC, dht22Msg.Payload.Temperature, dht22Msg.Payload.Humidity)
		}
	}()

	log.Println("🔄 [DHT22] Consumer iniciado. Esperando datos de sensores...")
	<-forever
	return nil
}

func (dc *DHT22Consumer) processMessage(dht22Msg *DHT22Message, originalBody []byte) error {
	var wg sync.WaitGroup
	var errors []error
	var errorsMutex sync.Mutex

	// Actualizar timestamp de último mensaje visto
	dc.updateLastSeen(dht22Msg.Payload.MAC)

	wg.Add(2) // InfluxDB + WebSocket

	// Goroutine 1: Escribir a InfluxDB
	go func() {
		defer wg.Done()
		if err := dc.writeToInfluxDB(dht22Msg); err != nil {
			errorsMutex.Lock()
			errors = append(errors, fmt.Errorf("InfluxDB error: %v", err))
			errorsMutex.Unlock()
		}
	}()

	// Goroutine 2: Publicar a WebSocket
	go func() {
		defer wg.Done()
		if err := dc.publishToWebSocket(dht22Msg, originalBody); err != nil {
			errorsMutex.Lock()
			errors = append(errors, fmt.Errorf("WebSocket error: %v", err))
			errorsMutex.Unlock()
		}
	}()

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("errores en procesamiento: %v", errors)
	}

	return nil
}

func (dc *DHT22Consumer) writeToInfluxDB(dht22Msg *DHT22Message) error {
	p := influxdb2.NewPoint("temperature_humidity_metrics",
		map[string]string{"deviceId": dht22Msg.DeviceID, "mac": dht22Msg.Payload.MAC},
		map[string]interface{}{
			"temperature": dht22Msg.Payload.Temperature,
			"humidity":    dht22Msg.Payload.Humidity,
		},
		time.Now(),
	)

	if err := dc.WriteAPI.WritePoint(context.Background(), p); err != nil {
		log.Printf("❌ [DHT22][InfluxDB] Error: %v", err)
		return err
	}

	log.Printf("✅ [DHT22][InfluxDB] Datos escritos para MAC: %s", dht22Msg.Payload.MAC)
	return nil
}

func (dc *DHT22Consumer) publishToWebSocket(dht22Msg *DHT22Message, originalBody []byte) error {
	contentPayload := ContentPayload{
		MAC:     dht22Msg.Payload.MAC,
		Message: string(originalBody),
	}

	contentBytes, _ := json.Marshal(contentPayload)
	wsMsg := WebSocketMessage{Content: string(contentBytes)}

	if err := dc.WSConn.WriteJSON(wsMsg); err != nil {
		log.Printf("❌ [DHT22][WebSocket] Error: %v", err)
		return err
	}

	log.Printf("✅ [DHT22][WebSocket] Mensaje enviado para MAC: %s", dht22Msg.Payload.MAC)
	return nil
}

func (dc *DHT22Consumer) updateLastSeen(mac string) {
	dc.LastSeenMutex.Lock()
	defer dc.LastSeenMutex.Unlock()
	dc.LastSeen[mac] = time.Now()
}

func (dc *DHT22Consumer) timeoutChecker() {
	log.Println("🕐 [DHT22] Iniciando verificador de timeouts...")
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			dc.checkTimeouts()
		}
	}
}

func (dc *DHT22Consumer) checkTimeouts() {
	dc.LastSeenMutex.RLock()
	now := time.Now()
	timeouts := make([]string, 0)

	for mac, lastSeen := range dc.LastSeen {
		if now.Sub(lastSeen) > timeoutDuration {
			timeouts = append(timeouts, mac)
		}
	}
	dc.LastSeenMutex.RUnlock()

	// Procesar timeouts encontrados
	for _, mac := range timeouts {
		log.Printf("⏰ [DHT22] TIMEOUT detectado para MAC: %s", mac)
		dc.publishTimeoutAlert(mac)

		// Remover del mapa para evitar spam de alertas
		dc.LastSeenMutex.Lock()
		delete(dc.LastSeen, mac)
		dc.LastSeenMutex.Unlock()
	}

	if len(timeouts) > 0 {
		log.Printf("🚨 [DHT22] %d timeouts procesados", len(timeouts))
	}
}

func (dc *DHT22Consumer) publishTimeoutAlert(mac string) {
	alert := AlertMessage{
		SensorMAC:  mac,
		SensorType: "dht22",
		AlertType:  "TIMEOUT",
		Message:    fmt.Sprintf("El sensor DHT22 '%s' ha dejado de reportar datos por más de %v", mac, timeoutDuration),
		Timestamp:  time.Now(),
	}

	alertBytes, err := json.Marshal(alert)
	if err != nil {
		log.Printf("❌ [DHT22] Error marshaling alerta: %v", err)
		return
	}

	err = dc.AlertsChannel.Publish(
		"",
		alertsQueueName,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        alertBytes,
		},
	)

	if err != nil {
		log.Printf("❌ [DHT22] Error publicando alerta: %v", err)
	} else {
		log.Printf("🚨 [DHT22] Alerta TIMEOUT publicada para MAC: %s", mac)
	}
}

func (dc *DHT22Consumer) Close() {
	log.Println("🔒 [DHT22] Cerrando conexiones...")
	if dc.WSConn != nil {
		dc.WSConn.Close()
	}
	if dc.Channel != nil {
		dc.Channel.Close()
	}
	if dc.AlertsChannel != nil {
		dc.AlertsChannel.Close()
	}
	if dc.RabbitConn != nil {
		dc.RabbitConn.Close()
	}
	if dc.InfluxClient != nil {
		dc.InfluxClient.Close()
	}
	log.Println("✅ [DHT22] Conexiones cerradas")
}

func main() {
	log.Println("🚀 Iniciando DHT22 Consumer...")

	consumer, err := NewDHT22Consumer()
	failOnError(err, "Error creando DHT22 consumer")
	defer consumer.Close()

	log.Fatal(consumer.Start())
}
