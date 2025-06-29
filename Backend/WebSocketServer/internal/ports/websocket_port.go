// WebSocketPort defines the interface for sending messages over a WebSocket connection.
// Implementations of this interface are responsible for handling the transmission of
// messages represented by the models.Message type.
package ports

import "github.com/M1keTrike/EventDriven/internal/models"

type WebSocketPort interface {
	SendMessage(msg *models.Message) error
}
