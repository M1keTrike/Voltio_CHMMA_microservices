// RabbitToSocketMiddleware.go - Notification Consumer para Voltio API
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// --- CONFIGURACIÓN ---
const (
	alertsQueueName = "alerts-queue"
	requestTimeout  = 30 * time.Second
	maxRetries      = 3
	retryDelay      = 5 * time.Second
)

var (
	amqpURI       = getEnv("RABBITMQ_URI", "amqp://admin:trike@52.73.74.139:5672/")
	apiWebhookURL = getEnv("WEBHOOK_URL", "https://voltioapi.acstree.xyz/api/internal/notifications/service")
)

// --- TIPOS DE ALERTAS SOPORTADOS ---
const (
	TIMEOUT_ALERT     = "TIMEOUT"     // ⏰ Dispositivo Sin Respuesta
	OFFLINE_ALERT     = "OFFLINE"     // 🔴 Dispositivo Desconectado
	ERROR_ALERT       = "ERROR"       // ⚠️ Error en Dispositivo
	WARNING_ALERT     = "WARNING"     // ⚠️ Advertencia del Sistema
	CRITICAL_ALERT    = "CRITICAL"    // 🚨 ALERTA CRÍTICA
	MAINTENANCE_ALERT = "MAINTENANCE" // 🔧 Mantenimiento Programado
	TEST_EMAIL        = "TEST_EMAIL"  // 🧪 Email de Prueba
)

// --- ESTRUCTURAS ---
type AlertMessage struct {
	SensorMAC      string                 `json:"sensor_mac"`
	SensorType     string                 `json:"sensor_type"`
	AlertType      string                 `json:"alert_type"`
	Message        string                 `json:"message"`
	Timestamp      time.Time              `json:"timestamp"`
	DeviceID       string                 `json:"device_id,omitempty"`
	ErrorCode      string                 `json:"error_code,omitempty"`
	Severity       string                 `json:"severity,omitempty"`
	Location       string                 `json:"location,omitempty"`
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

type APIWebhookPayload struct {
	ErrorType string `json:"error_type"`
	MAC       string `json:"mac"`
	Message   string `json:"message"`
}

type NotificationConsumer struct {
	RabbitConn *amqp091.Connection
	Channel    *amqp091.Channel
	HTTPClient *http.Client
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func NewNotificationConsumer() (*NotificationConsumer, error) {
	nc := &NotificationConsumer{
		HTTPClient: &http.Client{
			Timeout: requestTimeout,
		},
	}

	if err := nc.setupConnections(); err != nil {
		return nil, err
	}

	return nc, nil
}

func (nc *NotificationConsumer) setupConnections() error {
	var err error

	log.Println("Conectando a RabbitMQ...")
	nc.RabbitConn, err = amqp091.Dial(amqpURI)
	if err != nil {
		return fmt.Errorf("fallo al conectar con RabbitMQ: %v", err)
	}

	nc.Channel, err = nc.RabbitConn.Channel()
	if err != nil {
		return fmt.Errorf("fallo al abrir canal: %v", err)
	}

	// Declarar cola de alertas
	_, err = nc.Channel.QueueDeclare(alertsQueueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("fallo al declarar cola de alertas: %v", err)
	}

	log.Printf("✅ Conexión a RabbitMQ establecida - Cola: %s", alertsQueueName)
	return nil
}

func (nc *NotificationConsumer) Start() error {
	log.Println("🚨 Iniciando Notification Consumer para Voltio API...")
	log.Printf("📡 API Webhook: %s", apiWebhookURL)
	log.Printf("👂 Escuchando cola: %s", alertsQueueName)

	msgs, err := nc.Channel.Consume(alertsQueueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("fallo al registrar consumidor: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("📨 [ALERT] Mensaje recibido: %s", string(d.Body))

			// Parse del mensaje de alerta
			var alert AlertMessage
			if err := json.Unmarshal(d.Body, &alert); err != nil {
				log.Printf("❌ [ALERT] Error al parsear mensaje: %v", err)
				d.Nack(false, false) // Rechazar mensaje malformado
				continue
			}

			// Validar tipo de alerta
			if !nc.isValidAlertType(alert.AlertType) {
				log.Printf("⚠️ [ALERT] Tipo de alerta no válido: %s. Usando TIMEOUT por defecto.", alert.AlertType)
				alert.AlertType = TIMEOUT_ALERT
			}

			// Enviar a webhook con reintentos
			if err := nc.sendToWebhookWithRetry(&alert); err != nil {
				log.Printf("❌ [ALERT] Error enviando a webhook después de %d reintentos: %v", maxRetries, err)
				d.Nack(false, true) // Requeue para reintento posterior
				continue
			}

			// Éxito - confirmar mensaje
			d.Ack(false)
			log.Printf("✅ [ALERT] Alerta procesada y enviada - MAC: %s, Tipo: %s", alert.SensorMAC, alert.AlertType)
		}
	}()

	log.Println("🔄 [NOTIFICATION] Consumer iniciado. Esperando alertas...")
	<-forever
	return nil
}

func (nc *NotificationConsumer) isValidAlertType(alertType string) bool {
	validTypes := []string{TIMEOUT_ALERT, OFFLINE_ALERT, ERROR_ALERT, WARNING_ALERT, CRITICAL_ALERT, MAINTENANCE_ALERT, TEST_EMAIL}
	for _, valid := range validTypes {
		if alertType == valid {
			return true
		}
	}
	return false
}

func (nc *NotificationConsumer) sendToWebhookWithRetry(alert *AlertMessage) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("🔄 [WEBHOOK] Intento %d/%d - Enviando alerta %s para MAC: %s", attempt, maxRetries, alert.AlertType, alert.SensorMAC)

		if err := nc.sendToWebhook(alert); err != nil {
			lastErr = err
			log.Printf("❌ [WEBHOOK] Intento %d falló: %v", attempt, err)

			if attempt < maxRetries {
				log.Printf("⏳ [WEBHOOK] Esperando %v antes del siguiente intento...", retryDelay)
				time.Sleep(retryDelay)
			}
			continue
		}

		// Éxito
		log.Printf("✅ [WEBHOOK] Alerta enviada exitosamente en intento %d", attempt)
		return nil
	}

	return fmt.Errorf("falló después de %d intentos, último error: %v", maxRetries, lastErr)
}

func (nc *NotificationConsumer) sendToWebhook(alert *AlertMessage) error {
	// Crear payload según formato de tu API
	webhookPayload := APIWebhookPayload{
		ErrorType: alert.AlertType,
		MAC:       alert.SensorMAC,
		Message:   nc.formatAlertMessage(alert),
	}

	payloadBytes, err := json.Marshal(webhookPayload)
	if err != nil {
		return fmt.Errorf("error al marshaling payload: %v", err)
	}

	log.Printf("📤 [WEBHOOK] Enviando payload: %s", string(payloadBytes))

	// Crear petición HTTP
	req, err := http.NewRequest("POST", apiWebhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creando petición HTTP: %v", err)
	}

	// Headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Voltio-Notification-Consumer/2.0")
	req.Header.Set("X-Source", "RabbitMQ-Consumer")

	// Enviar petición
	resp, err := nc.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error enviando petición: %v", err)
	}
	defer resp.Body.Close()

	// Verificar respuesta
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook devolvió status no exitoso: %d %s", resp.StatusCode, resp.Status)
	}

	log.Printf("✅ [WEBHOOK] Respuesta exitosa - Status: %d", resp.StatusCode)
	return nil
}

func (nc *NotificationConsumer) formatAlertMessage(alert *AlertMessage) string {
	// Formatear mensaje según tipo de alerta
	switch alert.AlertType {
	case TIMEOUT_ALERT:
		// Mensaje específico para PZEM vs sensores regulares
		if alert.SensorType == "pzem" {
			return fmt.Sprintf("⏰ TIMEOUT - El medidor eléctrico PZEM '%s' ha dejado de reportar datos. REVISA INMEDIATAMENTE la instalación eléctrica y conexiones.", alert.SensorMAC)
		}
		return fmt.Sprintf("⏰ TIMEOUT - El dispositivo '%s' ha dejado de reportar datos. Por favor, revisa su conexión y estado físico.", alert.SensorMAC)
	case OFFLINE_ALERT:
		return fmt.Sprintf("🔴 OFFLINE - El dispositivo '%s' está completamente desconectado. Verifica alimentación y conexiones físicas.", alert.SensorMAC)
	case ERROR_ALERT:
		return fmt.Sprintf("⚠️ ERROR - Se ha detectado un error en el dispositivo '%s'. Revisa logs y contacta soporte técnico si persiste.", alert.SensorMAC)
	case WARNING_ALERT:
		return fmt.Sprintf("⚠️ WARNING - Advertencia del sistema para dispositivo '%s'. Revisa configuración y planifica mantenimiento.", alert.SensorMAC)
	case CRITICAL_ALERT:
		// Mensaje CRÍTICO específico para PZEM
		if alert.SensorType == "pzem" {
			return fmt.Sprintf("🚨 CRÍTICO - FALLA ELÉCTRICA DETECTADA en medidor PZEM '%s'. ATENCIÓN INMEDIATA REQUERIDA. Revisar sistemas eléctricos y seguridad.", alert.SensorMAC)
		}
		return fmt.Sprintf("🚨 CRÍTICO - ATENCIÓN INMEDIATA REQUERIDA para dispositivo '%s'. Situación crítica detectada.", alert.SensorMAC)
	case MAINTENANCE_ALERT:
		return fmt.Sprintf("🔧 MAINTENANCE - Mantenimiento programado para dispositivo '%s'. Solo informativo.", alert.SensorMAC)
	case TEST_EMAIL:
		return fmt.Sprintf("🧪 TEST_EMAIL - Email de prueba del sistema de notificaciones para dispositivo '%s'.", alert.SensorMAC)
	default:
		return fmt.Sprintf("📢 ALERTA - %s: %s", alert.AlertType, alert.Message)
	}
}

func (nc *NotificationConsumer) Close() {
	log.Println("🔒 Cerrando conexiones...")
	if nc.Channel != nil {
		nc.Channel.Close()
	}
	if nc.RabbitConn != nil {
		nc.RabbitConn.Close()
	}
	log.Println("✅ Conexiones cerradas correctamente")
}

func main() {
	log.Println("🚀 Iniciando Voltio Notification Consumer...")

	consumer, err := NewNotificationConsumer()
	failOnError(err, "Error creando notification consumer")
	defer consumer.Close()

	log.Fatal(consumer.Start())
}
