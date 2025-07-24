package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"./middleware"
)

func main() {
	log.Println("🚀 Iniciando PIR Consumer...")

	// Crear el consumer
	consumer, err := middleware.NewPIRConsumer()
	if err != nil {
		log.Fatalf("❌ Error creando PIR consumer: %v", err)
	}

	// Configurar cierre limpio
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("🛑 Señal de cierre recibida, cerrando consumer...")
		consumer.Close()
		os.Exit(0)
	}()

	// Iniciar el consumer
	if err := consumer.Start(); err != nil {
		log.Fatalf("❌ Error iniciando consumer: %v", err)
	}
}
