// StartServer initializes and starts the HTTP server using the Gin framework.
// It sets up the necessary dependencies, including an in-memory repository,
// a message service, and a WebSocket adapter. The server exposes a WebSocket
// endpoint at "/ws" and listens on port 8081.
package server

import (
	"net/http"

	"github.com/M1keTrike/EventDriven/internal/adapters"
	"github.com/M1keTrike/EventDriven/internal/core"
	"github.com/gin-gonic/gin"
)

func StartServer() {

	r := gin.Default()

	// Configurar CORS y headers para WebSocket
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	repo := adapters.NewInMemoryRepository()
	service := core.NewMessageService(repo)
	wsAdapter := adapters.NewWebSocketAdapter(service)

	// Endpoint de salud
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "websocket-server",
		})
	})

	r.GET("/ws", wsAdapter.HandleWebSocket)

	r.Run(":8081")
}
