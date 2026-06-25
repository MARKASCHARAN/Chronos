package eventstore

import (
	"context"

	"github.com/markasaicharan/chronos/internal/domain"
)

// Store defines the interface for persisting and retrieving workflow events.
type Store interface {
	SaveEvent(ctx context.Context, event domain.Event) error
	GetEvents(ctx context.Context, workflowID string) ([]domain.Event, error)
}
