package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/markasaicharan/chronos/internal/domain"
	"github.com/markasaicharan/chronos/pkg/eventstore"
	"github.com/markasaicharan/chronos/pkg/queue"
)

// Worker continuously polls the distributed queue and executes tasks.
// Activity defines the signature for any business logic step.
type Activity func(ctx context.Context, payload []byte) ([]byte, error)

type Worker struct {
	id         string
	store      eventstore.Store
	queue      queue.Queue
	activities map[string]Activity
}

// NewWorker creates a new worker node with a unique ID.
func NewWorker(store eventstore.Store, q queue.Queue) *Worker {
	return &Worker{
		id:         fmt.Sprintf("worker-%s", uuid.New().String()[:8]),
		store:      store,
		queue:      q,
		activities: make(map[string]Activity),
	}
}

// RegisterActivity connects a string task type to a Go function.
func (w *Worker) RegisterActivity(taskType string, activity Activity) {
	w.activities[taskType] = activity
}

// Start begins the infinite polling loop for this worker node.
func (w *Worker) Start(ctx context.Context) {
	log.Info("Worker online and polling Redis for tasks...", "worker_id", w.id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Shutting down gracefully...\n", w.id)
			return
		default:
			task, err := w.queue.Pop(ctx)
			if err != nil {
				log.Error("Error popping task", "worker_id", w.id, "error", err)
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
	log.Info("Claimed task", "worker_id", w.id, "task_type", task.TaskType, "workflow_id", task.WorkflowID)

	activity, exists := w.activities[task.TaskType]
	if !exists {
		log.Error("Unknown activity type", "worker_id", w.id, "task_type", task.TaskType)
		return
	}

	resultPayload, err := activity(ctx, task.Payload)
	if err != nil {
		log.Warn("Activity failed", "worker_id", w.id, "task_type", task.TaskType, "attempt", task.RetryCount+1, "error", err)

		task.RetryCount++
		if task.RetryCount < 3 {
			log.Info("Re-queuing task for retry...", "worker_id", w.id, "task_id", task.ID)
			_ = w.queue.Push(ctx, *task)
			return
		}

		log.Error("Task exceeded max retries. Moving to DLQ.", "worker_id", w.id, "task_id", task.ID)
		_ = w.queue.PushDLQ(ctx, *task)

		failedEvent := domain.Event{
			ID:         uuid.New().String(),
			WorkflowID: task.WorkflowID,
			EventType:  domain.EventTypeTaskFailed,
			Payload:    []byte(fmt.Sprintf(`{"error": "%s", "attempts": %d}`, err.Error(), task.RetryCount)),
			Timestamp:  time.Now().UTC(),
		}
		_ = w.store.SaveEvent(ctx, failedEvent)
		return
	}

	completedEvent := domain.Event{
		ID:         uuid.New().String(),
		WorkflowID: task.WorkflowID,
		EventType:  domain.EventTypeTaskCompleted,
		Payload:    resultPayload,
		Timestamp:  time.Now().UTC(),
	}

	if err := w.store.SaveEvent(ctx, completedEvent); err != nil {
		log.Fatal("Failed to save event to Postgres", "worker_id", w.id, "error", err)
		return
	}

	log.Info("Task completed & event persisted", "worker_id", w.id, "workflow_id", task.WorkflowID)
}
