package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/markasaicharan/chronos/internal/workflow"
	"github.com/markasaicharan/chronos/pkg/eventstore"
	"github.com/markasaicharan/chronos/pkg/queue"
)

// Worker continuously polls the distributed queue and executes tasks.
// In a real deployment, you would run dozens or hundreds of these concurrently.
type Worker struct {
	id    string
	store eventstore.Store
	queue queue.Queue
}

// NewWorker creates a new worker node with a unique ID.
func NewWorker(store eventstore.Store, q queue.Queue) *Worker {
	return &Worker{
		// Give each worker a tiny random ID so we can track logs
		id:    fmt.Sprintf("worker-%s", uuid.New().String()[:8]),
		store: store,
		queue: q,
	}
}

// Start begins the infinite polling loop for this worker node.
func (w *Worker) Start(ctx context.Context) {
	log.Printf("[%s] Worker online and polling Redis for tasks...\n", w.id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Shutting down gracefully...\n", w.id)
			return
		default:
			// 1. Pop blocks safely until Redis hands it a task
			task, err := w.queue.Pop(ctx)
			if err != nil {
				log.Printf("[%s] Error popping task: %v\n", w.id, err)
				time.Sleep(1 * time.Second) // Small backoff on error
				continue
			}

			if task == nil {
				continue
			}

			// 2. We have a task! Process it.
			w.processTask(ctx, task)
		}
	}
}

// processTask handles the core execution logic.
func (w *Worker) processTask(ctx context.Context, task *queue.Task) {
	log.Printf("[%s] Claimed task: %s for Workflow: %s\n", w.id, task.TaskType, task.WorkflowID)

	// Simulate heavy lifting (e.g., calling an external API, charging a credit card)
	time.Sleep(2 * time.Second)

	// In a complete Durable Execution engine, we would now:
	// 1. Fetch entire event history: w.store.GetEvents(ctx, task.WorkflowID)
	// 2. Replay history to rebuild the workflow state exactly.
	// 3. Execute the specific task logic.

	// 4. Save the result as an Event Sourced record.
	completedEvent := workflow.Event{
		ID:         uuid.New().String(),
		WorkflowID: task.WorkflowID,
		EventType:  workflow.EventTypeTaskCompleted,
		Payload:    []byte(fmt.Sprintf(`{"worker_id": "%s", "task_id": "%s"}`, w.id, task.ID)),
		Timestamp:  time.Now().UTC(),
	}

	if err := w.store.SaveEvent(ctx, completedEvent); err != nil {
		log.Printf("[%s] CRITICAL: Failed to save event to Postgres: %v\n", w.id, err)
		return
	}

	log.Printf("[%s] ✅ Task completed & event persisted (Workflow: %s)\n", w.id, task.WorkflowID)
}
