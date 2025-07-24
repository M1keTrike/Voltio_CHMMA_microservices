// pkg/config/config.go
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// RabbitMQ Configuration
	RabbitMQURI string
	QueueName   string

	// WebSocket Configuration
	WebSocketURI string

	// InfluxDB Configuration
	InfluxURL    string
	InfluxToken  string
	InfluxOrg    string
	InfluxBucket string

	// Consumer Configuration
	ConsumerType   string
	TopicName      string
	TimeoutSeconds int

	// Timeout Configuration
	TimeoutDuration time.Duration

	// Alerts Configuration
	AlertsQueueName string
}

func LoadConfig() *Config {
	return &Config{
		// RabbitMQ
		RabbitMQURI: getEnvOrDefault("RABBITMQ_URI", "amqp://admin:trike@52.73.74.139:5672/"),
		QueueName:   getEnvOrDefault("QUEUE_NAME", "PZEM_queue"),

		// WebSocket
		WebSocketURI: getEnvOrDefault("WEBSOCKET_URI", "wss://websocketvoltio.acstree.xyz/ws?topic=pzem&emitter=true"),

		// InfluxDB
		InfluxURL:    getEnvOrDefault("INFLUX_URL", "http://52.201.107.193:8086"),
		InfluxToken:  getEnvOrDefault("INFLUX_TOKEN", "lJLzxtHLHvPNgdvU9dcInGYb/qLbLxUPgrePzLd47EKCLUWBzJ+RmJkpH0f1HkmQ"),
		InfluxOrg:    getEnvOrDefault("INFLUX_ORG", "mi-org"),
		InfluxBucket: getEnvOrDefault("INFLUX_BUCKET", "sensores"),

		// Consumer
		ConsumerType:   getEnvOrDefault("CONSUMER_TYPE", "pzem"),
		TopicName:      getEnvOrDefault("TOPIC_NAME", "pzem"),
		TimeoutSeconds: getTimeoutSeconds(),

		// Timeout
		TimeoutDuration: getTimeoutDuration(),

		// Alerts
		AlertsQueueName: getEnvOrDefault("ALERTS_QUEUE_NAME", "alerts_queue"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getTimeoutSeconds() int {
	timeoutStr := getEnvOrDefault("TIMEOUT_SECONDS", "300")
	seconds, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return 300
	}
	return seconds
}

func getTimeoutDuration() time.Duration {
	return time.Duration(getTimeoutSeconds()) * time.Second
}
