// WebSocketAdapter manages WebSocket connections for both emitters and subscribers,
// organizing them by topic and MAC address. It provides methods to handle WebSocket
// upgrades, add and remove clients, and send messages to appropriate subscribers.
//
// The adapter maintains two main maps:
//   - clients: a nested map of topic -> mac -> set of WebSocket connections for subscribers.
//   - emitters: a map of topic -> WebSocket connection for emitters.
//
// Usage:
//   - NewWebSocketAdapter(service *core.MessageService) *WebSocketAdapter
//     Creates a new WebSocketAdapter instance.
//   - HandleWebSocket(c *gin.Context)
//     Handles incoming WebSocket connections, distinguishing between emitters and subscribers.
//   - SendMessage(topic string, msg *models.Message)
//     Sends a message to all subscribers of a specific topic and MAC address.
//
// Thread safety is ensured via an internal mutex.
package adapters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/M1keTrike/EventDriven/internal/core"
	"github.com/M1keTrike/EventDriven/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:       func(r *http.Request) bool { return true },
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
}

type WebSocketAdapter struct {
	mu       sync.Mutex
	clients  map[string]map[string]map[*websocket.Conn]bool // topic -> mac -> connections
	emitters map[string]*websocket.Conn
	service  *core.MessageService
}

func NewWebSocketAdapter(service *core.MessageService) *WebSocketAdapter {
	return &WebSocketAdapter{
		clients:  make(map[string]map[string]map[*websocket.Conn]bool),
		emitters: make(map[string]*websocket.Conn),
		service:  service,
	}
}

func (ws *WebSocketAdapter) HandleWebSocket(c *gin.Context) {
	topic := c.Query("topic")
	mac := c.Query("mac")
	isEmitter := c.Query("emitter") == "true"

	if topic == "" {
		fmt.Println("Error: No se proporcionó un tema en la conexión WebSocket")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Se requiere un tema en la URL"})
		return
	}

	if !isEmitter && mac == "" {
		fmt.Println("Error: No se proporcionó una MAC para el suscriptor")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Se requiere una MAC para suscriptores"})
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("Error al actualizar conexión WebSocket: %v\n", err)
		return
	}

	// Configurar timeouts y límites
	conn.SetReadLimit(512 * 1024) // 512KB max message size

	defer func() {
		conn.Close()
		ws.removeClient(topic, conn, isEmitter)
		if isEmitter {
			fmt.Printf("Emisor desconectado del tema: %s\n", topic)
		} else {
			fmt.Printf("Suscriptor (MAC: %s) desconectado del tema: %s\n", mac, topic)
		}
	}()

	ws.addClient(topic, mac, conn, isEmitter)

	fmt.Printf("Cliente %s conectado al tema: %s\n", ws.getConnectionType(isEmitter), topic)

	for {
		var msg models.Message
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				fmt.Printf("Error inesperado al leer WebSocket [%s]: %v\n", topic, err)
			}
			break
		}

		var prettyMessage map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Content), &prettyMessage); err != nil {
			fmt.Printf("Mensaje recibido en el servidor [%s]: %s\n", topic, msg.Content)
		} else {
			prettyJSON, _ := json.MarshalIndent(prettyMessage, "", "  ")
			fmt.Printf("Mensaje recibido en el servidor [%s]:\n%s\n", topic, string(prettyJSON))
		}

		ws.SendMessage(topic, &msg)
	}
}

func (ws *WebSocketAdapter) addClient(topic string, mac string, conn *websocket.Conn, isEmitter bool) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if isEmitter {
		if existingEmitter, exists := ws.emitters[topic]; exists {
			existingEmitter.Close()
		}
		ws.emitters[topic] = conn
	} else {
		if _, exists := ws.clients[topic]; !exists {
			ws.clients[topic] = make(map[string]map[*websocket.Conn]bool)
		}
		if _, exists := ws.clients[topic][mac]; !exists {
			ws.clients[topic][mac] = make(map[*websocket.Conn]bool)
		}
		ws.clients[topic][mac][conn] = true
	}
}

func (ws *WebSocketAdapter) removeClient(topic string, conn *websocket.Conn, isEmitter bool) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if isEmitter {

		if ws.emitters[topic] == conn {
			delete(ws.emitters, topic)
			delete(ws.clients, topic)
			fmt.Printf("Tema %s eliminado porque el emisor se desconectó\n", topic)
		}
	} else {
		if _, exists := ws.clients[topic]; exists {
			// Iterate through all MACs to find and remove the connection
			for mac, connections := range ws.clients[topic] {
				if _, exists := connections[conn]; exists {
					delete(connections, conn)
					// Remove the MAC if no connections remain
					if len(connections) == 0 {
						delete(ws.clients[topic], mac)
					}
					break
				}
			}
		}
	}
}

func (ws *WebSocketAdapter) SendMessage(topic string, msg *models.Message) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	fmt.Printf("Intentando enviar mensaje en el tema: %s\n", topic)

	if macSubscribers, exists := ws.clients[topic]; exists {
		// Extraer MAC del mensaje
		var msgContent map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Content), &msgContent); err != nil {
			fmt.Printf("Error parseando mensaje: %v\n", err)
			return
		}

		msgMAC, ok := msgContent["mac"].(string)
		if !ok {
			fmt.Printf("MAC no encontrada en el mensaje\n")
			return
		}

		// Enviar a suscriptores de la MAC específica
		if subscribers, exists := macSubscribers[msgMAC]; exists {
			sentCount := 0
			for conn := range subscribers {
				if err := conn.WriteJSON(msg); err != nil {
					fmt.Printf("Error enviando mensaje a suscriptor: %v\n", err)
					conn.Close()
					delete(subscribers, conn)
				} else {
					sentCount++
				}
			}
			fmt.Printf("Mensaje enviado a %d suscriptor(es) de MAC %s en tema %s\n", sentCount, msgMAC, topic)
		} else {
			fmt.Printf("No hay suscriptores para MAC %s en tema %s\n", msgMAC, topic)
		}
	}
}

func (ws *WebSocketAdapter) getConnectionType(isEmitter bool) string {
	if isEmitter {
		return "Emisor"
	}
	return "Suscriptor"
}
