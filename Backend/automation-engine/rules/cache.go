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
	TriggerDeviceID    int
	ActionDeviceID     int
	Name               string
	IsActive           bool
	TriggerDeviceMAC   string
	ActionDeviceMAC    string
	TriggerMetric      string
	Operator           string
	Threshold          float64
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
			ID:               rule.ID,
			TriggerDeviceMAC: rule.TriggerDeviceMAC,
			ActionDeviceMAC:  rule.ActionDeviceMAC,
			TriggerMetric:    rule.TriggerMetric,
			Operator:         rule.Operator,
			Threshold:        rule.Threshold,
			ActionPayload:    rule.ActionPayload,
			ActiveStart:      rule.ActiveStart,
			ActiveEnd:        rule.ActiveEnd,
		}
		newCache[mac][metric] = append(newCache[mac][metric], converted)
	}

	CacheMutex.Lock()
	Cache = newCache
	CacheMutex.Unlock()
	log.Println("[Rules] Caché de reglas actualizada.")
}
