// RabbitToSocketMiddleware.go - PZEM Consumer con Sistema de Alertas
package main

import (
	"context"
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
	// Variables de entorno con valores por defecto
	amqpURI         = getEnv("RABBITMQ_URI", "amqp://admin:trike@52.73.74.139:5672/")
	queueName       = "PZEM_queue"
	alertsQueueName = "alerts-queue"
	wsURI           = getEnv("PZEM_WEBSOCKET_URI", "wss://voltiows.acstree.xyz/ws?topic=pzem&emitter=true")

	// InfluxDB Configuration
	influxURL    = getEnv("INFLUXDB_URL", "http://52.201.107.193:8086")
	influxToken  = getEnv("INFLUXDB_TOKEN", "lJLzxtHLHvPNgdvU9dcInGYb/qLbLxUPgrePzLd47EKCLUWBzJ+RmJkpH0f1HkmQ")
	influxOrg    = getEnv("INFLUXDB_ORG", "mi-org")
	influxBucket = getEnv("INFLUXDB_BUCKET", "sensores")

	// Timeout Configuration - Más tiempo para PZEM (equipos más estables)
	timeoutDuration = 5 * time.Minute  // 5 minutos vs 2 minutos de sensores
	checkInterval   = 60 * time.Second // Check cada minuto
)

// getEnv obtiene variable de entorno con valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// --- ESTRUCTURAS ---
type PZEMMessage struct {
	Payload struct {
		MAC         string  `json:"mac"`
		Voltage     float64 `json:"voltage"`
		Current     float64 `json:"current"`
		Power       float64 `json:"power"`
		Energy      float64 `json:"energy"`
		Frequency   float64 `json:"frequency"`
		PowerFactor float64 `json:"powerFactor"`
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

type PZEMConsumer struct {
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

func NewPZEMConsumer() (*PZEMConsumer, error) {
	pc := &PZEMConsumer{
		LastSeen: make(map[string]time.Time),
	}

	if err := pc.setupConnections(); err != nil {
		return nil, err
	}

	// Start timeout checker goroutine
	go pc.timeoutChecker()

	return pc, nil
}

func (pc *PZEMConsumer) setupConnections() error {
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
		return fmt.Errorf("fallo al declarar cola PZEM: %v", err)
	}

	_, err = pc.AlertsChannel.QueueDeclare(alertsQueueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("fallo al declarar cola de alertas: %v", err)
	}

	// WebSocket Connection with TLS support
	log.Println("Conectando a WebSocket...")
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
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
	pc.WriteAPI = pc.InfluxClient.WriteAPIBlocking(influxOrg, influxBucket)

	log.Println("✅ Todas las conexiones PZEM listas")
	return nil
}

func (pc *PZEMConsumer) Start() error {
	log.Println("⚡ Iniciando PZEM Consumer con Sistema de Alertas CRÍTICAS...")
	log.Printf("👂 Escuchando cola: %s", queueName)
	log.Printf("🚨 Cola de alertas: %s", alertsQueueName)
	log.Printf("⏰ Timeout configurado: %v (CRÍTICO para energía eléctrica)", timeoutDuration)

	msgs, err := pc.Channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("fallo al registrar consumidor: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("📨 [PZEM] Mensaje recibido: %s", string(d.Body)[:100])

			var pzemMsg PZEMMessage
			if err := json.Unmarshal(d.Body, &pzemMsg); err != nil {
				log.Printf("❌ [PZEM] Error al decodificar JSON: %v", err)
				d.Nack(false, false)
				continue
			}

			// Procesar mensaje en paralelo
			if err := pc.processMessage(&pzemMsg, d.Body); err != nil {
				log.Printf("❌ [PZEM] Error procesando mensaje: %v", err)
				d.Nack(false, true) // Requeue para reintento
				continue
			}

			// Éxito
			d.Ack(false)
			log.Printf("✅ [PZEM] Mensaje procesado - MAC: %s, Potencia: %.1fW, Voltaje: %.1fV",
				pzemMsg.Payload.MAC, pzemMsg.Payload.Power, pzemMsg.Payload.Voltage)
		}
	}()

	log.Println("🔄 [PZEM] Consumer iniciado. Esperando datos de medidores eléctricos...")
	<-forever
	return nil
}

func (pc *PZEMConsumer) processMessage(pzemMsg *PZEMMessage, originalBody []byte) error {
	var wg sync.WaitGroup
	var errors []error
	var errorsMutex sync.Mutex

	// Actualizar timestamp de último mensaje visto
	pc.updateLastSeen(pzemMsg.Payload.MAC)

	wg.Add(2) // InfluxDB + WebSocket

	// Goroutine 1: Escribir a InfluxDB
	go func() {
		defer wg.Done()
		if err := pc.writeToInfluxDB(pzemMsg); err != nil {
			errorsMutex.Lock()
			errors = append(errors, fmt.Errorf("InfluxDB error: %v", err))
			errorsMutex.Unlock()
		}
	}()

	// Goroutine 2: Publicar a WebSocket
	go func() {
		defer wg.Done()
		if err := pc.publishToWebSocket(pzemMsg, originalBody); err != nil {
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

func (pc *PZEMConsumer) writeToInfluxDB(pzemMsg *PZEMMessage) error {
	p := influxdb2.NewPoint("energy_metrics",
		map[string]string{"mac": pzemMsg.Payload.MAC},
		map[string]interface{}{
			"voltage":     pzemMsg.Payload.Voltage,
			"current":     pzemMsg.Payload.Current,
			"power":       pzemMsg.Payload.Power,
			"energy":      pzemMsg.Payload.Energy,
			"frequency":   pzemMsg.Payload.Frequency,
			"powerFactor": pzemMsg.Payload.PowerFactor,
		},
		time.Now(),
	)

	if err := pc.WriteAPI.WritePoint(context.Background(), p); err != nil {
		log.Printf("❌ [PZEM][InfluxDB] Error: %v", err)
		return err
	}

	log.Printf("✅ [PZEM][InfluxDB] Datos escritos para MAC: %s", pzemMsg.Payload.MAC)
	return nil
}

func (pc *PZEMConsumer) publishToWebSocket(pzemMsg *PZEMMessage, originalBody []byte) error {
	contentPayload := ContentPayload{
		MAC:     pzemMsg.Payload.MAC,
		Message: string(originalBody),
	}

	contentBytes, _ := json.Marshal(contentPayload)
	wsMsg := WebSocketMessage{Content: string(contentBytes)}

	if err := pc.WSConn.WriteJSON(wsMsg); err != nil {
		log.Printf("❌ [PZEM][WebSocket] Error: %v", err)
		return err
	}

	log.Printf("✅ [PZEM][WebSocket] Mensaje enviado para MAC: %s", pzemMsg.Payload.MAC)
	return nil
}

func (pc *PZEMConsumer) updateLastSeen(mac string) {
	pc.LastSeenMutex.Lock()
	defer pc.LastSeenMutex.Unlock()
	pc.LastSeen[mac] = time.Now()
}

func (pc *PZEMConsumer) timeoutChecker() {
	log.Println("🕐 [PZEM] Iniciando verificador de timeouts CRÍTICOS...")
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pc.checkTimeouts()
		}
	}
}

func (pc *PZEMConsumer) checkTimeouts() {
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
		log.Printf("🚨 [PZEM] TIMEOUT CRÍTICO detectado para MAC: %s", mac)
		pc.publishCriticalAlert(mac)

		// Remover del mapa para evitar spam de alertas
		pc.LastSeenMutex.Lock()
		delete(pc.LastSeen, mac)
		pc.LastSeenMutex.Unlock()
	}

	if len(timeouts) > 0 {
		log.Printf("🚨 [PZEM] %d timeouts CRÍTICOS procesados", len(timeouts))
	}
}

func (pc *PZEMConsumer) publishCriticalAlert(mac string) {
	alert := AlertMessage{
		SensorMAC:  mac,
		SensorType: "pzem",
		AlertType:  "CRITICAL", // CRÍTICO para equipos eléctricos
		Message:    fmt.Sprintf("FALLA CRÍTICA: El medidor PZEM '%s' ha dejado de reportar datos por más de %v. REVISAR INMEDIATAMENTE sistemas eléctricos.", mac, timeoutDuration),
		Timestamp:  time.Now(),
	}

	alertBytes, err := json.Marshal(alert)
	if err != nil {
		log.Printf("❌ [PZEM] Error marshaling alerta CRÍTICA: %v", err)
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
		log.Printf("❌ [PZEM] Error publicando alerta CRÍTICA: %v", err)
	} else {
		log.Printf("🚨 [PZEM] ⚠️  ALERTA CRÍTICA publicada para MAC: %s ⚠️", mac)
	}
}

func (pc *PZEMConsumer) Close() {
	log.Println("🔒 [PZEM] Cerrando conexiones...")
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
	log.Println("✅ [PZEM] Conexiones cerradas")
}

func main() {
	log.Println("🚀 Iniciando PZEM Consumer con Alertas CRÍTICAS...")

	consumer, err := NewPZEMConsumer()
	failOnError(err, "Error creando PZEM consumer")
	defer consumer.Close()

	log.Fatal(consumer.Start())
}
