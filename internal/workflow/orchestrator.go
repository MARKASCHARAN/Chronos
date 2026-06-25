package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/markasaicharan/chronos/pkg/eventstore"
	"github.com/markasaicharan/chronos/pkg/queue"
)

// Orchestrator coordinates the execution of workflows.
// It acts as the brain, bridging the Event Store and the Task Queue.
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
// It records the exact starting event in PostgreSQL, then pushes the initial
// task to Redis so the worker cluster can pick it up.
func (o *Orchestrator) StartWorkflow(ctx context.Context, name string, payload []byte) (string, error) {
	workflowID := uuid.New().String()

	startedEvent := Event{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		EventType:  EventTypeWorkflowStarted,
		Payload:    payload,
		Timestamp:  time.Now().UTC(),
	}

	if err := o.store.SaveEvent(ctx, startedEvent); err != nil {
		return "", fmt.Errorf("failed to persist WorkflowStarted event: %w", err)
	}

	task := queue.Task{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		TaskType:   "ProcessWorkflow", // A generic task instructing workers to evaluate the workflow state
		Payload:    payload,
	}

	if err := o.queue.Push(ctx, task); err != nil {
		return "", fmt.Errorf("failed to queue initial task: %w", err)
	}

	return workflowID, nil
}
