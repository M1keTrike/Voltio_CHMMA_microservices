package main

import (
	"automation-engine/config"
	"automation-engine/messaging"
	"automation-engine/rules"
	"log"
	"time"
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

	 go func() {
        ticker := time.NewTicker(1 * time.Minute)
        for range ticker.C {
            messaging.CheckAndTriggerWorkdayRules()
        }
    }()

	
	log.Println("[Automation-Engine] Iniciando consumidor de eventos...")
	messaging.StartConsumer()
}
