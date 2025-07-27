package messaging

import (
	"automation-engine/rules"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/streadway/amqp"
)

var (
	conn    *amqp.Connection
	channel *amqp.Channel
)

// Recorre reglas en caché y dispara acciones si el horario actual coincide con ActiveStart/ActiveEnd
func CheckAndTriggerWorkdayRules() {
	now := time.Now()
	rules.CacheMutex.Lock()
	defer rules.CacheMutex.Unlock()
	for _, metricRules := range rules.Cache {
		for _, reglas := range metricRules {
			for _, regla := range reglas {
				// Encender al iniciar jornada laboral
				if regla.ActiveStart != nil && now.Hour() == regla.ActiveStart.Hour() && now.Minute() == regla.ActiveStart.Minute() {
					if regla.TriggerMetric == "workday_start" {
						triggerAction(regla)
					}
				}
				// Apagar al terminar jornada laboral
				if regla.ActiveEnd != nil && now.Hour() == regla.ActiveEnd.Hour() && now.Minute() == regla.ActiveEnd.Minute() {
					if regla.TriggerMetric == "workday_end" {
						triggerAction(regla)
					}
				}
			}
		}
	}
}

func StartConsumer() {
	var err error
	uri := os.Getenv("RABBITMQ_URI")
	if uri == "" {
		uri = "amqp://guest:guest@localhost:5672/"
	}
	conn, err = amqp.Dial(uri)
	if err != nil {
		log.Fatalf("[RabbitMQ] Error de conexión: %v", err)
	}
	channel, err = conn.Channel()
	if err != nil {
		log.Fatalf("[RabbitMQ] Error abriendo canal: %v", err)
	}

	// Declarar exchange y cola
	exchange := "amq.topic"
	queue, err := channel.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		log.Fatalf("[RabbitMQ] Error declarando cola: %v", err)
	}
	if err := channel.QueueBind(queue.Name, "*.data.events", exchange, false, nil); err != nil {
		log.Fatalf("[RabbitMQ] Error en binding: %v", err)
	}

	msgs, err := channel.Consume(queue.Name, "", true, true, false, false, nil)
	if err != nil {
		log.Fatalf("[RabbitMQ] Error iniciando consumo: %v", err)
	}

	log.Println("[RabbitMQ] Esperando mensajes de eventos...")
	for d := range msgs {
		go processMessage(d.Body)
	}
}

func processMessage(body []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("[RabbitMQ] Error decodificando mensaje: %v", err)
		return
	}

	mac, _ := msg["mac"].(string)
	payload, ok := msg["payload"].(map[string]interface{})
	if !ok {
		log.Printf("[RabbitMQ] Payload inválido")
		return
	}

	for metric, value := range payload {
		var metricValue float64
		switch v := value.(type) {
		case float64:
			metricValue = v
		case int:
			metricValue = float64(v)
		}

		rules.CacheMutex.Lock()
		rulesToEvaluate := rules.Cache[mac][metric]
		rules.CacheMutex.Unlock()

		for _, rule := range rulesToEvaluate {
			if !isRuleActive(rule) {
				continue
			}
			if evaluateRule(rule, metricValue) {
				triggerAction(rule)
			}
		}
	}
}

func isRuleActive(rule rules.AutomationRule) bool {
	now := time.Now()
	if rule.ActiveStart != nil && rule.ActiveEnd != nil {
		if now.Before(*rule.ActiveStart) || now.After(*rule.ActiveEnd) {
			return false
		}
	}
	return true
}

func evaluateRule(rule rules.AutomationRule, value float64) bool {
	switch rule.Operator {
	case "GREATER_THAN":
		return value > rule.Threshold
	case "LESS_THAN":
		return value < rule.Threshold
	case "EQUAL":
		return value == rule.Threshold
	case "NOT_EQUAL":
		return value != rule.Threshold
	default:
		return false
	}
}

func triggerAction(rule rules.AutomationRule) {
	var endpoint string
	var capability string
	switch rule.ActionCapabilityID {
	case 1:
		endpoint = fmt.Sprintf("https://voltioapi.acstree.xyz/api/v1/devices/%s/command/relay", rule.ActionDeviceMAC)
		capability = "RELAY_CONTROL"
	case 2:
		endpoint = fmt.Sprintf("https://voltioapi.acstree.xyz/api/v1/devices/%s/command/ir", rule.ActionDeviceMAC)
		capability = "INFRARED_EMITTER"
	default:
		log.Printf("[AutomationEngine] Acción ignorada: action_capability_id inválido (%v)", rule.ActionCapabilityID)
		return
	}

	// El body debe ser: {"action": "ON"} o {"action": "OFF"}
	var payload struct {
		Action string `json:"action"`
	}
	// Se asume que ActionPayload es "ON" u "OFF" en mayúsculas
	payload.Action = rule.ActionPayload
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[AutomationEngine] Error serializando payload: %v", err)
		return
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[AutomationEngine] Error creando request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[AutomationEngine] Error enviando acción (%s): %v", capability, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("[AutomationEngine] Acción disparada (%s): %s -> %s", capability, endpoint, string(body))
	} else {
		log.Printf("[AutomationEngine] Error en respuesta (%s): %s, status: %d", capability, endpoint, resp.StatusCode)
	}
}
