package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/markasaicharan/chronos/internal/domain"
	"github.com/markasaicharan/chronos/pkg/eventstore"
	"github.com/markasaicharan/chronos/pkg/queue"
)

// Worker continuously polls the distributed queue and executes tasks.
type Worker struct {
	id    string
	store eventstore.Store
	queue queue.Queue
}

// NewWorker creates a new worker node with a unique ID.
func NewWorker(store eventstore.Store, q queue.Queue) *Worker {
	return &Worker{
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
			task, err := w.queue.Pop(ctx)
			if err != nil {
				log.Printf("[%s] Error popping task: %v\n", w.id, err)
				time.Sleep(1 * time.Second)
				continue
			}

			if task == nil {
				continue
			}

			w.processTask(ctx, task)
		}
	}
}

// processTask handles the core execution logic.
func (w *Worker) processTask(ctx context.Context, task *queue.Task) {
	log.Printf("[%s] Claimed task: %s for Workflow: %s\n", w.id, task.TaskType, task.WorkflowID)

	time.Sleep(2 * time.Second)

	completedEvent := domain.Event{
		ID:         uuid.New().String(),
		WorkflowID: task.WorkflowID,
		EventType:  domain.EventTypeTaskCompleted,
		Payload:    []byte(fmt.Sprintf(`{"worker_id": "%s", "task_id": "%s"}`, w.id, task.ID)),
		Timestamp:  time.Now().UTC(),
	}

	if err := w.store.SaveEvent(ctx, completedEvent); err != nil {
		log.Printf("[%s] CRITICAL: Failed to save event to Postgres: %v\n", w.id, err)
		return
	}

	log.Printf("[%s] ✅ Task completed & event persisted (Workflow: %s)\n", w.id, task.WorkflowID)
}
