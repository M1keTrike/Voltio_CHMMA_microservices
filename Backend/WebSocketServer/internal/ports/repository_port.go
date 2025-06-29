// MessageRepositoryPort defines the contract for a repository that handles
// the persistence of Message entities. Implementations of this interface
// are responsible for saving messages to a storage medium.
//
// SaveMessage persists the provided Message entity and returns an error
// if the operation fails.
package ports

import "github.com/M1keTrike/EventDriven/internal/models"

type MessageRepositoryPort interface {
	SaveMessage(msg *models.Message) error
}
