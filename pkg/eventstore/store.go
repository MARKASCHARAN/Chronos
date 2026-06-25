package eventstore

import (
	"context"

	"github.com/markasaicharan/chronos/internal/workflow"
)

// Store defines the interface for persisting and retrieving workflow events.
type Store interface {
	SaveEvent(ctx context.Context, event workflow.Event) error
	GetEvents(ctx context.Context, workflowID string) ([]workflow.Event, error)
}
