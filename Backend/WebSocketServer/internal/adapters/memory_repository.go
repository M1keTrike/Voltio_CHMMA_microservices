// InMemoryRepository provides a thread-safe in-memory implementation for storing messages.
// It uses a mutex to ensure safe concurrent access to the underlying slice of messages.
package adapters

import (
	"sync"

	"github.com/M1keTrike/EventDriven/internal/models"
)

type InMemoryRepository struct {
	messages []models.Message
	mu       sync.Mutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{}
}

func (r *InMemoryRepository) SaveMessage(msg *models.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = append(r.messages, *msg)
	return nil
}


