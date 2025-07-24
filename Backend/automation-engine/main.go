package main

import (
	"time"
	"log"
	"automation-engine/config"
	"automation-engine/rules"
	"automation-engine/messaging"
)

func main() {
	// 1. Cargar configuración desde variables de entorno.
	config.Load()

	// 2. Cargar la caché de reglas por primera vez.
	rules.UpdateCache()

	// 3. Iniciar la goroutine de sincronización periódica de la caché.
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			rules.UpdateCache()
		}
	}()

	// 4. Iniciar el consumidor de RabbitMQ (bloqueante).
	log.Println("[Automation-Engine] Iniciando consumidor de eventos...")
	messaging.StartConsumer()
}
