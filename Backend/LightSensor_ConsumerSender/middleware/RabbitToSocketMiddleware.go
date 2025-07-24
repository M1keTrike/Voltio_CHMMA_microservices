// RabbitToSocketMiddleware.go - Light Sensor Consumer con Sistema de Alertas
package main

import (
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
	queueName       = "LightSensor_queue"
	alertsQueueName = "alerts-queue"
	wsURI           = "wss://websocketvoltio.acstree.xyz/ws?topic=light_sensor&emitter=true"

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
type LightSensorMessage struct {
	DeviceID string `json:"deviceId"`
	Payload  struct {
		MAC        string  `json:"mac"`
		LightLevel float64 `json:"lightLevel"`
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

type LightSensorConsumer struct {
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

func NewLightSensorConsumer() (*LightSensorConsumer, error) {
	lsc := &LightSensorConsumer{
		LastSeen: make(map[string]time.Time),
	}

	if err := lsc.setupConnections(); err != nil {
		return nil, err
	}

	// Start timeout checker goroutine
	go lsc.timeoutChecker()

	return lsc, nil
}

func (lsc *LightSensorConsumer) setupConnections() error {
	var err error

	// RabbitMQ Connection
	log.Println("Conectando a RabbitMQ...")
	lsc.RabbitConn, err = amqp091.Dial(amqpURI)
	if err != nil {
		return fmt.Errorf("fallo al conectar con RabbitMQ: %v", err)
	}

	lsc.Channel, err = lsc.RabbitConn.Channel()
	if err != nil {
		return fmt.Errorf("fallo al abrir canal principal: %v", err)
	}

	// Separate channel for alerts
	lsc.AlertsChannel, err = lsc.RabbitConn.Channel()
	if err != nil {
		return fmt.Errorf("fallo al abrir canal de alertas: %v", err)
	}

	// Declarar colas
	_, err = lsc.Channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("fallo al declarar cola Light Sensor: %v", err)
	}

	_, err = lsc.AlertsChannel.QueueDeclare(alertsQueueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("fallo al declarar cola de alertas: %v", err)
	}

	// WebSocket Connection
	log.Println("Conectando a WebSocket...")
	lsc.WSConn, _, err = websocket.DefaultDialer.Dial(wsURI, nil)
	if err != nil {
		return fmt.Errorf("fallo al conectar con WebSocket: %v", err)
	}

	// InfluxDB Connection
	log.Println("Configurando cliente de InfluxDB...")
	lsc.InfluxClient = influxdb2.NewClient(influxURL, influxToken)
	lsc.WriteAPI = lsc.InfluxClient.WriteAPI(influxOrg, influxBucket)

	log.Println("✅ Todas las conexiones Light Sensor listas")
	return nil
}

func (lsc *LightSensorConsumer) Start() error {
	log.Println("💡 Iniciando Light Sensor Consumer con Alertas...")
	log.Printf("👂 Escuchando cola: %s", queueName)
	log.Printf("🚨 Cola de alertas: %s", alertsQueueName)

	msgs, err := lsc.Channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("fallo al registrar consumidor: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("📨 [LIGHT] Mensaje recibido: %s", string(d.Body))

			var lightMsg LightSensorMessage
			if err := json.Unmarshal(d.Body, &lightMsg); err != nil {
				log.Printf("❌ [LIGHT] Error al decodificar JSON: %v", err)
				d.Nack(false, false)
				continue
			}

			// Procesar mensaje en paralelo
			if err := lsc.processMessage(&lightMsg, d.Body); err != nil {
				log.Printf("❌ [LIGHT] Error procesando mensaje: %v", err)
				d.Nack(false, true) // Requeue para reintento
				continue
			}

			// Éxito
			d.Ack(false)
			log.Printf("✅ [LIGHT] Mensaje procesado - MAC: %s, Nivel: %.1f lux",
				lightMsg.Payload.MAC, lightMsg.Payload.LightLevel)
		}
	}()

	log.Println("🔄 [LIGHT] Consumer iniciado. Esperando datos de sensores...")
	<-forever
	return nil
}

func (lsc *LightSensorConsumer) processMessage(lightMsg *LightSensorMessage, originalBody []byte) error {
	var wg sync.WaitGroup
	var errors []error
	var errorsMutex sync.Mutex

	// Actualizar timestamp de último mensaje visto
	lsc.updateLastSeen(lightMsg.Payload.MAC)

	wg.Add(2) // InfluxDB + WebSocket

	// Goroutine 1: Escribir a InfluxDB
	go func() {
		defer wg.Done()
		if err := lsc.writeToInfluxDB(lightMsg); err != nil {
			errorsMutex.Lock()
			errors = append(errors, fmt.Errorf("InfluxDB error: %v", err))
			errorsMutex.Unlock()
		}
	}()

	// Goroutine 2: Publicar a WebSocket
	go func() {
		defer wg.Done()
		if err := lsc.publishToWebSocket(lightMsg, originalBody); err != nil {
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

func (lsc *LightSensorConsumer) writeToInfluxDB(lightMsg *LightSensorMessage) error {
	p := influxdb2.NewPoint("light_sensor_metrics",
		map[string]string{"mac": lightMsg.Payload.MAC},
		map[string]interface{}{
			"light_level": lightMsg.Payload.LightLevel,
		},
		time.Now(),
	)

	lsc.WriteAPI.WritePoint(p)
	lsc.WriteAPI.Flush()

	log.Printf("✅ [LIGHT][InfluxDB] Datos escritos para MAC: %s", lightMsg.Payload.MAC)
	return nil
}

func (lsc *LightSensorConsumer) publishToWebSocket(lightMsg *LightSensorMessage, originalBody []byte) error {
	contentPayload := ContentPayload{
		MAC:     lightMsg.Payload.MAC,
		Message: string(originalBody),
	}

	contentBytes, _ := json.Marshal(contentPayload)
	wsMsg := WebSocketMessage{Content: string(contentBytes)}

	if err := lsc.WSConn.WriteJSON(wsMsg); err != nil {
		log.Printf("❌ [LIGHT][WebSocket] Error: %v", err)
		return err
	}

	log.Printf("✅ [LIGHT][WebSocket] Mensaje enviado para MAC: %s", lightMsg.Payload.MAC)
	return nil
}

func (lsc *LightSensorConsumer) updateLastSeen(mac string) {
	lsc.LastSeenMutex.Lock()
	defer lsc.LastSeenMutex.Unlock()
	lsc.LastSeen[mac] = time.Now()
}

func (lsc *LightSensorConsumer) timeoutChecker() {
	log.Println("🕐 [LIGHT] Iniciando verificador de timeouts...")
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lsc.checkTimeouts()
		}
	}
}

func (lsc *LightSensorConsumer) checkTimeouts() {
	lsc.LastSeenMutex.RLock()
	now := time.Now()
	timeouts := make([]string, 0)

	for mac, lastSeen := range lsc.LastSeen {
		if now.Sub(lastSeen) > timeoutDuration {
			timeouts = append(timeouts, mac)
		}
	}
	lsc.LastSeenMutex.RUnlock()

	// Procesar timeouts encontrados
	for _, mac := range timeouts {
		log.Printf("⏰ [LIGHT] TIMEOUT detectado para MAC: %s", mac)
		lsc.publishTimeoutAlert(mac)

		// Remover del mapa para evitar spam de alertas
		lsc.LastSeenMutex.Lock()
		delete(lsc.LastSeen, mac)
		lsc.LastSeenMutex.Unlock()
	}

	if len(timeouts) > 0 {
		log.Printf("🚨 [LIGHT] %d timeouts procesados", len(timeouts))
	}
}

func (lsc *LightSensorConsumer) publishTimeoutAlert(mac string) {
	alert := AlertMessage{
		SensorMAC:  mac,
		SensorType: "light_sensor",
		AlertType:  "TIMEOUT",
		Message:    fmt.Sprintf("El sensor de luz '%s' ha dejado de reportar datos por más de %v", mac, timeoutDuration),
		Timestamp:  time.Now(),
	}

	alertBytes, err := json.Marshal(alert)
	if err != nil {
		log.Printf("❌ [LIGHT] Error marshaling alerta: %v", err)
		return
	}

	err = lsc.AlertsChannel.Publish(
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
		log.Printf("❌ [LIGHT] Error publicando alerta: %v", err)
	} else {
		log.Printf("🚨 [LIGHT] Alerta TIMEOUT publicada para MAC: %s", mac)
	}
}

func (lsc *LightSensorConsumer) Close() {
	log.Println("🔒 [LIGHT] Cerrando conexiones...")
	if lsc.WSConn != nil {
		lsc.WSConn.Close()
	}
	if lsc.Channel != nil {
		lsc.Channel.Close()
	}
	if lsc.AlertsChannel != nil {
		lsc.AlertsChannel.Close()
	}
	if lsc.RabbitConn != nil {
		lsc.RabbitConn.Close()
	}
	if lsc.InfluxClient != nil {
		lsc.InfluxClient.Close()
	}
	log.Println("✅ [LIGHT] Conexiones cerradas")
}

func main() {
	log.Println("🚀 Iniciando Light Sensor Consumer...")

	consumer, err := NewLightSensorConsumer()
	failOnError(err, "Error creando Light Sensor consumer")
	defer consumer.Close()

	log.Fatal(consumer.Start())
}
