package models

import "time"

// AutomationRuleModel representa la estructura de una regla en la base de datos.
type AutomationRuleModel struct {
	ID                 int
	UserID             int
	TriggerDeviceMAC   string
	ActionDeviceMAC    string
	Name               string
	IsActive           bool
	TriggerMetric      string
	ComparisonOperator string
	ThresholdValue     float64
	ActionCapabilityID int
	ActionPayload      string
	ActiveStart        *time.Time
	ActiveEnd          *time.Time
	CreatedAt          *time.Time
}

type AutomationRuleBody struct {
	UserID             int     `json:"user_id"`
	TriggerDeviceMAC   string  `json:"trigger_device_mac"`
	ActionDeviceMAC    string  `json:"action_device_mac"`
	Name               string  `json:"name"`
	IsActive           bool    `json:"is_active"`
	TriggerMetric      string  `json:"trigger_metric"`
	ComparisonOperator string  `json:"comparison_operator"`
	ThresholdValue     float64 `json:"threshold_value"`
	ActionCapabilityID int     `json:"action_capability_id"`
	ActionPayload      string  `json:"action_payload"`
	ActiveTimeStart    string  `json:"active_time_start,omitempty"` // formato "HH:MM"
	ActiveTimeEnd      string  `json:"active_time_end,omitempty"`   // formato "HH:MM"
}
