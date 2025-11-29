package main

import (
	"github.com/joho/godotenv"
	"automation-engine/messaging"
	"automation-engine/rules"
	"log"
	"time"

)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("[Config] No se pudo cargar el archivo .env, usando variables del sistema")
	} else {
		log.Println("[Config] Variables de entorno cargadas desde .env")
	}

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
