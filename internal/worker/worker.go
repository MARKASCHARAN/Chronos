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

	activity, exists := w.activities[task.TaskType]
	if !exists {
		log.Printf("[%s] Unknown activity type: %s\n", w.id, task.TaskType)
		return
	}

	resultPayload, err := activity(ctx, task.Payload)
	if err != nil {
		log.Printf("[%s] Activity %s failed (Attempt %d): %v\n", w.id, task.TaskType, task.RetryCount+1, err)
		
		task.RetryCount++
		if task.RetryCount < 3 {
			log.Printf("[%s] Re-queuing task %s for retry...\n", w.id, task.ID)
			// Small delay before retry could be added here in a real system
			_ = w.queue.Push(ctx, *task)
			return
		}

		log.Printf("[%s] 🚨 Task %s exceeded max retries. Moving to DLQ.\n", w.id, task.ID)
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
		log.Printf("[%s] CRITICAL: Failed to save event to Postgres: %v\n", w.id, err)
		return
	}

	log.Printf("[%s] ✅ Task completed & event persisted (Workflow: %s)\n", w.id, task.WorkflowID)
}
