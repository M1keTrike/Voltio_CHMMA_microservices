package models

import "time"

// AutomationRuleModel representa la estructura de una regla en la base de datos.
type AutomationRuleModel struct {
	ID                int
	TriggerDeviceMAC  string
	ActionDeviceMAC   string
	TriggerMetric     string
	Operator          string
	Threshold         float64
	ActionPayload     string
	ActiveStart       *time.Time
	ActiveEnd         *time.Time
}
