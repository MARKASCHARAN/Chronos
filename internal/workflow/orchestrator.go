package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/markasaicharan/chronos/internal/domain"
	"github.com/markasaicharan/chronos/pkg/eventstore"
	"github.com/markasaicharan/chronos/pkg/queue"
)

// Orchestrator coordinates the execution of workflows.
type Orchestrator struct {
	store eventstore.Store
	queue queue.Queue
}

// NewOrchestrator creates a new workflow orchestrator.
func NewOrchestrator(store eventstore.Store, q queue.Queue) *Orchestrator {
	return &Orchestrator{
		store: store,
		queue: q,
	}
}

// StartWorkflow initiates a new workflow.
func (o *Orchestrator) StartWorkflow(ctx context.Context, name string, payload []byte) (string, error) {
	workflowID := uuid.New().String()

	startedEvent := domain.Event{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		EventType:  domain.EventTypeWorkflowStarted,
		Payload:    payload,
		Timestamp:  time.Now().UTC(),
	}

	if err := o.store.SaveEvent(ctx, startedEvent); err != nil {
		return "", fmt.Errorf("failed to persist WorkflowStarted event: %w", err)
	}

	task := queue.Task{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		TaskType:   "charge_customer",
		Payload:    payload,
	}

	if err := o.queue.Push(ctx, task); err != nil {
		return "", fmt.Errorf("failed to queue initial task: %w", err)
	}

	return workflowID, nil
}
