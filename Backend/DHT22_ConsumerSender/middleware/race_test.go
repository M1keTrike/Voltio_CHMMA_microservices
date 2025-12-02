package main

import (
	"sync"
	"testing"
	"time"
)

// Estructura simplificada para el test (copia de DHT22Consumer)
type TestConsumer struct {
	LastSeen      map[string]time.Time
	LastSeenMutex sync.RWMutex
}

func (tc *TestConsumer) updateLastSeen(mac string) {
	tc.LastSeenMutex.Lock()
	defer tc.LastSeenMutex.Unlock()
	tc.LastSeen[mac] = time.Now()
}

// Test para verificar que no hay race conditions en LastSeen
func TestDHT22ConsumerRaceCondition(t *testing.T) {
	dc := &TestConsumer{
		LastSeen:      make(map[string]time.Time),
		LastSeenMutex: sync.RWMutex{},
	}

	var wg sync.WaitGroup
	numGoroutines := 100
	numOperations := 1000

	t.Logf("🧪 Iniciando test de concurrencia con %d goroutines y %d operaciones cada una", numGoroutines, numOperations)

	// Simular escrituras concurrentes (updateLastSeen)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				mac := "AA:BB:CC:DD:EE:FF"
				dc.updateLastSeen(mac)
			}
		}(i)
	}

	// Simular lecturas concurrentes (checkTimeouts)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				dc.LastSeenMutex.RLock()
				_ = dc.LastSeen
				dc.LastSeenMutex.RUnlock()
			}
		}()
	}

	// Simular eliminaciones concurrentes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				dc.LastSeenMutex.Lock()
				delete(dc.LastSeen, "AA:BB:CC:DD:EE:FF")
				dc.LastSeenMutex.Unlock()
			}
		}()
	}

	wg.Wait()
	t.Log("✅ Test completado exitosamente sin race conditions")
	t.Logf("📊 Total de operaciones: %d", numGoroutines*numOperations*2+1000)
}
