package database

import (
	"automation-engine/models"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	// Obtener configuración de variables de entorno
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "chmma")
	password := getEnv("POSTGRES_PASSWORD", "HSQCx3Ajt4p^aJGC")
	dbname := getEnv("POSTGRES_DB", "voltiodb")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	log.Printf("[DB] Conectando a PostgreSQL: host=%s port=%s dbname=%s", host, port, dbname)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("[DB] Error abriendo conexión: %v", err)
	}
}

// getEnv obtiene variable de entorno con valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func LoadAllActiveRules() ([]models.AutomationRuleModel, error) {
	rows, err := db.Query(`SELECT id, user_id, trigger_device_mac, action_device_mac, name, is_active, trigger_metric, comparison_operator, threshold_value, action_capability_id, action_payload, active_time_start, active_time_end, created_at FROM automation_rules WHERE is_active = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.AutomationRuleModel
	for rows.Next() {
		var rule models.AutomationRuleModel
		var activeStart, activeEnd, createdAt sql.NullTime
		if err := rows.Scan(
			&rule.ID,
			&rule.UserID,
			&rule.TriggerDeviceMAC,
			&rule.ActionDeviceMAC,
			&rule.Name,
			&rule.IsActive,
			&rule.TriggerMetric,
			&rule.ComparisonOperator,
			&rule.ThresholdValue,
			&rule.ActionCapabilityID,
			&rule.ActionPayload,
			&activeStart,
			&activeEnd,
			&createdAt,
		); err != nil {
			return nil, err
		}
		if activeStart.Valid {
			rule.ActiveStart = &activeStart.Time
		}
		if activeEnd.Valid {
			rule.ActiveEnd = &activeEnd.Time
		}
		if createdAt.Valid {
			rule.CreatedAt = &createdAt.Time
		}
		result = append(result, rule)
	}
	return result, nil
}
