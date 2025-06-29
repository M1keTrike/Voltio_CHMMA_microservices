// StartServer initializes and starts the HTTP server using the Gin framework.
// It sets up the necessary dependencies, including an in-memory repository,
// a message service, and a WebSocket adapter. The server exposes a WebSocket
// endpoint at "/ws" and listens on port 8081.
package server

import (
	"github.com/M1keTrike/EventDriven/internal/adapters"
	"github.com/M1keTrike/EventDriven/internal/core"
	"github.com/gin-gonic/gin"
)

func StartServer() {

	r := gin.Default()

	repo := adapters.NewInMemoryRepository()
	service := core.NewMessageService(repo)
	wsAdapter := adapters.NewWebSocketAdapter(service)

	r.GET("/ws", wsAdapter.HandleWebSocket)

	r.Run(":8081")
}
