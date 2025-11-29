
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

// Mapa para guardar el último timestamp de movimiento por MAC
var lastMotion = make(map[string]time.Time)

// Goroutine para checar reglas de ausencia de movimiento
func StartMotionTimeoutChecker() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			now := time.Now()
			rules.CacheMutex.Lock()
			for mac, metricRules := range rules.Cache {
				// Buscar reglas de tipo motion_timeout
				reglas := metricRules["motion_timeout"]
				for _, regla := range reglas {
					if !isRuleActive(regla) || !regla.IsActive {
						continue
					}
					// threshold_value es el tiempo de espera en segundos
					timeout := int(regla.ThresholdValue)
					last, ok := lastMotion[mac]
					if !ok || now.Sub(last).Seconds() >= float64(timeout) {
						log.Printf("[AutomationEngine] Trigger por ausencia de movimiento para MAC: %s, timeout: %d seg", mac, timeout)
						triggerAction(regla)
						// Actualizar timestamp para evitar triggers repetidos
						lastMotion[mac] = now
					}
				}
			}
			rules.CacheMutex.Unlock()
		}
	}()
}

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

   // Iniciar checker de ausencia de movimiento
   StartMotionTimeoutChecker()
}

func processMessage(body []byte) {

	// Log de todos los mensajes recibidos
	log.Printf("[RabbitMQ] Mensaje recibido: %s", string(body))
	// Intentar decodificar como PZEMMessage primero
	var pzemMsg struct {
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
	isPzem := false
	if err := json.Unmarshal(body, &pzemMsg); err == nil {

		// Solo procesar como PZEM si el mensaje tiene payload.mac y al menos voltage, current o power
		if pzemMsg.Payload.MAC != "" && (pzemMsg.Payload.Voltage != 0 || pzemMsg.Payload.Current != 0 || pzemMsg.Payload.Power != 0) {

			isPzem = true
			mac := pzemMsg.Payload.MAC
			metrics := map[string]float64{
				"voltage":   pzemMsg.Payload.Voltage,
				"current":   pzemMsg.Payload.Current,
				"power":     pzemMsg.Payload.Power,
				"energy":    pzemMsg.Payload.Energy,
				"frequency": pzemMsg.Payload.Frequency,
				"pf":        pzemMsg.Payload.PowerFactor,
			}
			log.Printf("Mensaje PZEM processado: %s", mac)
			log.Printf("Procesando mensaje PZEM: %v", metrics)

			for metric, metricValue := range metrics {
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
	}
	if isPzem {
		return
	}

	// Si no es PZEMMessage, procesar como mensaje estándar
	var msg struct {
		SensorMAC  string                 `json:"sensor_mac"`
		SensorType string                 `json:"sensor_type"`
		Timestamp  interface{}            `json:"timestamp"`
		Data       map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("[RabbitMQ] Error decodificando mensaje: %v", err)
		return
	}

	mac := msg.SensorMAC
	payload := msg.Data
	if mac == "" || payload == nil {
		log.Printf("[RabbitMQ] Mensaje inválido: falta sensor_mac o data")
		log.Printf("[RabbitMQ] Mensaje recibido: %s", string(body))
		return
	}

   for metric, value := range payload {
	   var metricValue float64
	   switch v := value.(type) {
	   case float64:
		   metricValue = v
	   case int:
		   metricValue = float64(v)
	   case bool:
		   if v {
			   metricValue = 1.0
		   } else {
			   metricValue = 0.0
		   }
	   }

	   // Guardar timestamp de último movimiento si metric es "motion" y valor es 1
	   if metric == "motion" && metricValue == 1.0 {
		   lastMotion[mac] = time.Now()
	   }

	   rules.CacheMutex.Lock()
	   rulesToEvaluate := rules.Cache[mac][metric]
	   log.Printf("[AutomationEngine] Evaluando reglas para MAC: %s, métrica: %s, valor: %v", mac, metric, metricValue)
	   log.Printf("[AutomationEngine] Reglas encontradas: %v", rulesToEvaluate)
	   rules.CacheMutex.Unlock()
	   for _, rule := range rulesToEvaluate {
		   if !isRuleActive(rule) {
			   continue
		   }
		   // Para motion, threshold_value=1 es verdadero, 0 es falso
		   if metric == "motion" && (rule.ThresholdValue == 1 || rule.ThresholdValue == 0) {
			   if metricValue == rule.ThresholdValue {
				   log.Printf("[AutomationEngine] Trigger activado para regla: %+v", rule)
				   triggerAction(rule)
			   }
			   continue
		   }
		   if evaluateRule(rule, metricValue) {
			   log.Printf("[AutomationEngine] Trigger activado para regla: %+v", rule)
			   triggerAction(rule)
		   }
	   }
   }
}

func isRuleActive(rule rules.AutomationRule) bool {
	now := time.Now()
	if rule.ActiveStart != nil && rule.ActiveEnd != nil {
		// Comparar solo hora y minutos
		start := rule.ActiveStart.Hour()*60 + rule.ActiveStart.Minute()
		end := rule.ActiveEnd.Hour()*60 + rule.ActiveEnd.Minute()
		current := now.Hour()*60 + now.Minute()
		if current < start || current > end {
			return false
		}
	}
	return true
}

func evaluateRule(rule rules.AutomationRule, value float64) bool {
	switch rule.ComparisonOperator {
	case "GREATER_THAN":
		return value > rule.ThresholdValue
	case "LESS_THAN":
		return value < rule.ThresholdValue
	case "EQUAL":
		return value == rule.ThresholdValue
	case "NOT_EQUAL":
		return value != rule.ThresholdValue
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
