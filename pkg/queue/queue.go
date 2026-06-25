package queue

import (
	"context"
)

// Task represents a unit of work that needs to be executed.
type Task struct {
	ID         string
	WorkflowID string
	TaskType   string
	Payload    []byte
}

// Queue defines the interface for our distributed task queue.
type Queue interface {
	Push(ctx context.Context, task Task) error
	Pop(ctx context.Context) (*Task, error)
}
