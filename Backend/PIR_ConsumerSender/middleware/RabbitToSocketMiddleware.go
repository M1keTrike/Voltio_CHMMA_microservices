// RabbitToSocketMiddleware.go - PIR Consumer con Sistema de Alertas
package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rabbitmq/amqp091-go"
)

// --- CONFIGURACIÓN ---
var (
	amqpURI         = getEnv("RABBITMQ_URI", "amqp://admin:trike@52.73.74.139:5672/")
	queueName       = "PIR_queue"
	alertsQueueName = "alerts-queue"
	wsURI           = getEnv("PIR_WEBSOCKET_URI", "wss://voltiows.acstree.xyz/ws?topic=pir&emitter=true")

	// InfluxDB Configuration
	influxURL    = getEnv("INFLUXDB_URL", "http://52.201.107.193:8086")
	influxToken  = getEnv("INFLUXDB_TOKEN", "lJLzxtHLHvPNgdvU9dcInGYb/qLbLxUPgrePzLd47EKCLUWBzJ+RmJkpH0f1HkmQ")
	influxOrg    = getEnv("INFLUXDB_ORG", "mi-org")
	influxBucket = getEnv("INFLUXDB_BUCKET", "sensores")

	// Timeout Configuration
	timeoutDuration = 2 * time.Minute
	checkInterval   = 30 * time.Second
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// --- ESTRUCTURAS ---
type PIRMessage struct {
	SensorMAC  string    `json:"sensor_mac"`
	SensorType string    `json:"sensor_type"`
	Timestamp  time.Time `json:"timestamp"`
	Data       struct {
		Motion bool `json:"motion"`
	} `json:"data"`
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

type PIRConsumer struct {
	RabbitConn    *amqp091.Connection
	Channel       *amqp091.Channel
	AlertsChannel *amqp091.Channel
	WSConn        *websocket.Conn
	InfluxClient  influxdb2.Client
	WriteAPI      api.WriteAPI

	// Timeout tracking
	LastSeen      map[string]time.Time
	LastSeenMutex sync.RWMutex
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func NewPIRConsumer() (*PIRConsumer, error) {
	pc := &PIRConsumer{
		LastSeen: make(map[string]time.Time),
	}

	if err := pc.setupConnections(); err != nil {
		return nil, err
	}

	// Start timeout checker goroutine
	go pc.timeoutChecker()

	return pc, nil
}

func (pc *PIRConsumer) setupConnections() error {
	var err error

	// RabbitMQ Connection
	log.Println("Conectando a RabbitMQ...")
	pc.RabbitConn, err = amqp091.Dial(amqpURI)
	if err != nil {
		return fmt.Errorf("fallo al conectar con RabbitMQ: %v", err)
	}

	pc.Channel, err = pc.RabbitConn.Channel()
	if err != nil {
		return fmt.Errorf("fallo al abrir canal principal: %v", err)
	}

	// Separate channel for alerts
	pc.AlertsChannel, err = pc.RabbitConn.Channel()
	if err != nil {
		return fmt.Errorf("fallo al abrir canal de alertas: %v", err)
	}

	// Declarar colas
	_, err = pc.Channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("fallo al declarar cola PIR: %v", err)
	}

	_, err = pc.AlertsChannel.QueueDeclare(alertsQueueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("fallo al declarar cola de alertas: %v", err)
	}

	// WebSocket Connection with TLS support
	log.Println("Conectando a WebSocket...")
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false, // Validar certificados SSL
		},
		HandshakeTimeout: 10 * time.Second,
	}
	pc.WSConn, _, err = dialer.Dial(wsURI, http.Header{})
	if err != nil {
		return fmt.Errorf("fallo al conectar con WebSocket: %v", err)
	}

	// InfluxDB Connection
	log.Println("Configurando cliente de InfluxDB...")
	pc.InfluxClient = influxdb2.NewClient(influxURL, influxToken)
	pc.WriteAPI = pc.InfluxClient.WriteAPI(influxOrg, influxBucket)

	log.Println("✅ Todas las conexiones PIR listas")
	return nil
}

func (pc *PIRConsumer) Start() error {
	log.Println("🚶 Iniciando PIR Consumer con Alertas...")
	log.Printf("👂 Escuchando cola: %s", queueName)
	log.Printf("🚨 Cola de alertas: %s", alertsQueueName)

	msgs, err := pc.Channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("fallo al registrar consumidor: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("📨 [PIR] Mensaje recibido: %s", string(d.Body))

			var pirMsg PIRMessage
			if err := json.Unmarshal(d.Body, &pirMsg); err != nil {
				log.Printf("❌ [PIR] Error al decodificar JSON: %v", err)
				d.Nack(false, false)
				continue
			}

			// Procesar mensaje en paralelo
			if err := pc.processMessage(&pirMsg, d.Body); err != nil {
				log.Printf("❌ [PIR] Error procesando mensaje: %v", err)
				d.Nack(false, true) // Requeue para reintento
				continue
			}

			// Éxito
			d.Ack(false)
			log.Printf("✅ [PIR] Mensaje procesado - MAC: %s, Movimiento: %t",
				pirMsg.SensorMAC, pirMsg.Data.Motion)
		}
	}()

	log.Println("🔄 [PIR] Consumer iniciado. Esperando datos de sensores...")
	<-forever
	return nil
}

func (pc *PIRConsumer) processMessage(pirMsg *PIRMessage, originalBody []byte) error {
	var wg sync.WaitGroup
	var errors []error
	var errorsMutex sync.Mutex

	// Actualizar timestamp de último mensaje visto
	pc.updateLastSeen(pirMsg.SensorMAC)

	wg.Add(2) // InfluxDB + WebSocket

	// Goroutine 1: Escribir a InfluxDB
	go func() {
		defer wg.Done()
		if err := pc.writeToInfluxDB(pirMsg); err != nil {
			errorsMutex.Lock()
			errors = append(errors, fmt.Errorf("InfluxDB error: %v", err))
			errorsMutex.Unlock()
		}
	}()

	// Goroutine 2: Publicar a WebSocket
	go func() {
		defer wg.Done()
		if err := pc.publishToWebSocket(pirMsg, originalBody); err != nil {
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

func (pc *PIRConsumer) writeToInfluxDB(pirMsg *PIRMessage) error {
	p := influxdb2.NewPoint("motion_sensor_metrics",
		map[string]string{"mac": pirMsg.SensorMAC},
		map[string]interface{}{
			"motion_detected": pirMsg.Data.Motion,
		},
		time.Now(),
	)

	pc.WriteAPI.WritePoint(p)
	pc.WriteAPI.Flush()

	log.Printf("✅ [PIR][InfluxDB] Datos escritos para MAC: %s", pirMsg.SensorMAC)
	return nil
}

func (pc *PIRConsumer) publishToWebSocket(pirMsg *PIRMessage, originalBody []byte) error {
	contentPayload := ContentPayload{
		MAC:     pirMsg.SensorMAC,
		Message: string(originalBody),
	}

	contentBytes, _ := json.Marshal(contentPayload)
	wsMsg := WebSocketMessage{Content: string(contentBytes)}

	if err := pc.WSConn.WriteJSON(wsMsg); err != nil {
		log.Printf("❌ [PIR][WebSocket] Error: %v", err)
		return err
	}

	log.Printf("✅ [PIR][WebSocket] Mensaje enviado para MAC: %s", pirMsg.SensorMAC)
	return nil
}

func (pc *PIRConsumer) updateLastSeen(mac string) {
	pc.LastSeenMutex.Lock()
	defer pc.LastSeenMutex.Unlock()
	pc.LastSeen[mac] = time.Now()
}

func (pc *PIRConsumer) timeoutChecker() {
	log.Println("🕐 [PIR] Iniciando verificador de timeouts...")
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pc.checkTimeouts()
		}
	}
}

func (pc *PIRConsumer) checkTimeouts() {
	pc.LastSeenMutex.RLock()
	now := time.Now()
	timeouts := make([]string, 0)

	for mac, lastSeen := range pc.LastSeen {
		if now.Sub(lastSeen) > timeoutDuration {
			timeouts = append(timeouts, mac)
		}
	}
	pc.LastSeenMutex.RUnlock()

	// Procesar timeouts encontrados
	for _, mac := range timeouts {
		log.Printf("⏰ [PIR] TIMEOUT detectado para MAC: %s", mac)
		pc.publishTimeoutAlert(mac)

		// Remover del mapa para evitar spam de alertas
		pc.LastSeenMutex.Lock()
		delete(pc.LastSeen, mac)
		pc.LastSeenMutex.Unlock()
	}

	if len(timeouts) > 0 {
		log.Printf("🚨 [PIR] %d timeouts procesados", len(timeouts))
	}
}

func (pc *PIRConsumer) publishTimeoutAlert(mac string) {
	alert := AlertMessage{
		SensorMAC:  mac,
		SensorType: "pir",
		AlertType:  "TIMEOUT",
		Message:    fmt.Sprintf("El sensor PIR '%s' ha dejado de reportar datos por más de %v", mac, timeoutDuration),
		Timestamp:  time.Now(),
	}

	alertBytes, err := json.Marshal(alert)
	if err != nil {
		log.Printf("❌ [PIR] Error marshaling alerta: %v", err)
		return
	}

	err = pc.AlertsChannel.Publish(
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
		log.Printf("❌ [PIR] Error publicando alerta: %v", err)
	} else {
		log.Printf("🚨 [PIR] Alerta TIMEOUT publicada para MAC: %s", mac)
	}
}

func (pc *PIRConsumer) Close() {
	log.Println("🔒 [PIR] Cerrando conexiones...")
	if pc.WSConn != nil {
		pc.WSConn.Close()
	}
	if pc.Channel != nil {
		pc.Channel.Close()
	}
	if pc.AlertsChannel != nil {
		pc.AlertsChannel.Close()
	}
	if pc.RabbitConn != nil {
		pc.RabbitConn.Close()
	}
	if pc.InfluxClient != nil {
		pc.InfluxClient.Close()
	}
	log.Println("✅ [PIR] Conexiones cerradas")
}

func main() {
	log.Println("🚀 Iniciando PIR Consumer...")

	consumer, err := NewPIRConsumer()
	failOnError(err, "Error creando PIR consumer")
	defer consumer.Close()

	log.Fatal(consumer.Start())
}
