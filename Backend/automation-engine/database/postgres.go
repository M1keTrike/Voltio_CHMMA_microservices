package database

import (
    "database/sql"
    _ "github.com/lib/pq"
    "log"
    "os"
    "automation-engine/models"
)

var db *sql.DB

func init() {
    connStr := os.Getenv("POSTGRES_CONN")
    if connStr == "" {
        connStr = "host=" + os.Getenv("POSTGRES_HOST") +
            " port=" + os.Getenv("POSTGRES_PORT") +
            " user=" + os.Getenv("POSTGRES_USER") +
            " password=" + os.Getenv("POSTGRES_PASSWORD") +
            " dbname=" + os.Getenv("POSTGRES_DB") +
            " sslmode=disable"
    }
    var err error
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatalf("[DB] Error abriendo conexión: %v", err)
    }
}

func LoadAllActiveRules() ([]models.AutomationRuleModel, error) {
    rows, err := db.Query(`SELECT id, trigger_device_mac, action_device_mac, trigger_metric, operator, threshold, action_payload, active_start, active_end FROM automation_rules WHERE enabled = true`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var result []models.AutomationRuleModel
    for rows.Next() {
        var rule models.AutomationRuleModel
        var activeStart, activeEnd sql.NullTime
        if err := rows.Scan(&rule.ID, &rule.TriggerDeviceMAC, &rule.ActionDeviceMAC, &rule.TriggerMetric, &rule.Operator, &rule.Threshold, &rule.ActionPayload, &activeStart, &activeEnd); err != nil {
            return nil, err
        }
        if activeStart.Valid {
            rule.ActiveStart = &activeStart.Time
        }
        if activeEnd.Valid {
            rule.ActiveEnd = &activeEnd.Time
        }
        result = append(result, rule)
    }
    return result, nil
}
