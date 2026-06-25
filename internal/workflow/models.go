package workflow

import (
	"time"
)

// EventType defines the type of event in the event history.
type EventType string

const (
	EventTypeWorkflowStarted   EventType = "WorkflowStarted"
	EventTypeTaskScheduled     EventType = "TaskScheduled"
	EventTypeTaskStarted       EventType = "TaskStarted"
	EventTypeTaskCompleted     EventType = "TaskCompleted"
	EventTypeTaskFailed        EventType = "TaskFailed"
	EventTypeWorkflowCompleted EventType = "WorkflowCompleted"
	EventTypeWorkflowFailed    EventType = "WorkflowFailed"
)

// Event represents a single state change in a workflow's lifecycle.
// This is the core of our Event Sourcing engine.
type Event struct {
	ID         string
	WorkflowID string
	EventType  EventType
	Payload    []byte // JSON encoded data specific to the event
	Timestamp  time.Time
}

// WorkflowState represents the current overall status of the workflow.
type WorkflowState string

const (
	StateRunning   WorkflowState = "RUNNING"
	StateCompleted WorkflowState = "COMPLETED"
	StateFailed    WorkflowState = "FAILED"
)

// Workflow represents a running instance of a business process.
type Workflow struct {
	ID        string
	Name      string
	State     WorkflowState
	CreatedAt time.Time
	UpdatedAt time.Time
}
