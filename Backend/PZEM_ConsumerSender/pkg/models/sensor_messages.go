// pkg/models/sensor_messages.go
package models

import "time"

// DHT22 Message - simplified structure as per specifications
type DHT22Message struct {
	MAC         string  `json:"mac"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

// Light Sensor Message - simplified structure
type LightSensorMessage struct {
	MAC        string  `json:"mac"`
	LightLevel float64 `json:"light_level"`
}

// PIR Message - simplified structure
type PIRSensorMessage struct {
	MAC            string `json:"mac"`
	MotionDetected bool   `json:"motion_detected"`
}

// Timeout Alert Message for notification consumer
type TimeoutAlert struct {
	MAC       string    `json:"mac"`
	ErrorType string    `json:"error_type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
