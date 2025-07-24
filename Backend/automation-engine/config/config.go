package config

import (
    "log"
    "github.com/joho/godotenv"
)

func Load() {
    if err := godotenv.Load(); err != nil {
        log.Println("[Config] .env no encontrado, usando variables de entorno del sistema")
    }
    log.Println("[Config] Variables de entorno cargadas")
}
