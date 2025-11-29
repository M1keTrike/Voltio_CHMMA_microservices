package rules

import (
	"automation-engine/database"
	"log"
	"sync"
	"time"
)

type AutomationRule struct {
	ID                 int
	UserID             int
	Name               string
	IsActive           bool
	TriggerDeviceMAC   string
	ActionDeviceMAC    string
	TriggerMetric      string
	ComparisonOperator string
	ThresholdValue     float64
	ActionCapabilityID int
	ActionPayload      string
	ActiveStart        *time.Time
	ActiveEnd          *time.Time
	CreatedAt          *time.Time
}

var (
	Cache      = make(map[string]map[string][]AutomationRule)
	CacheMutex = &sync.Mutex{}
)

func UpdateCache() {
	rulesList, err := database.LoadAllActiveRules()
	if err != nil {
		log.Printf("[Rules] Error cargando reglas: %v", err)
		return
	}

	newCache := make(map[string]map[string][]AutomationRule)
	for _, rule := range rulesList {
		mac := rule.TriggerDeviceMAC
		metric := rule.TriggerMetric
		if newCache[mac] == nil {
			newCache[mac] = make(map[string][]AutomationRule)
		}
		converted := AutomationRule{
			ID:                 rule.ID,
			UserID:             rule.UserID,
			Name:               rule.Name,
			IsActive:           rule.IsActive,
			TriggerDeviceMAC:   rule.TriggerDeviceMAC,
			ActionDeviceMAC:    rule.ActionDeviceMAC,
			TriggerMetric:      rule.TriggerMetric,
			ComparisonOperator: rule.ComparisonOperator,
			ThresholdValue:     rule.ThresholdValue,
			ActionCapabilityID: rule.ActionCapabilityID,
			ActionPayload:      rule.ActionPayload,
			ActiveStart:        rule.ActiveStart,
			ActiveEnd:          rule.ActiveEnd,
			CreatedAt:          rule.CreatedAt,
		}
		newCache[mac][metric] = append(newCache[mac][metric], converted)
	}

	CacheMutex.Lock()
	Cache = newCache
	CacheMutex.Unlock()
	log.Println("[Rules] Caché de reglas actualizada.")
}
