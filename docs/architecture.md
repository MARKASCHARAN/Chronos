# Chronos System Design & Architecture

Chronos is a highly decoupled, distributed execution platform. It guarantees that background jobs (workflows) finish successfully, even if the servers executing them crash midway.

## 🛠 Tech Stack
- **Language:** Go (Golang) - chosen for its massive concurrency model and clean performance.
- **Event Store:** PostgreSQL - used as an append-only immutable ledger (Event Sourcing).
- **Task Queue:** Redis - handles distributed task routing using `LPush` and `BRPop`.
- **API:** Standard HTTP REST API.

---

## 🏛 The Core Components

1. **API Gateway & Orchestrator:** The HTTP server. It receives requests, persists the initial `WorkflowStarted` event to Postgres, and pushes the execution task to Redis.
2. **Event Store (The Time Machine):** Instead of a standard database updating `status="done"`, every single state change is saved as a brand new event. This creates a perfect, rewindable audit log.
3. **Task Queue (The Router):** Redis holds JSON tasks. It guarantees that exactly *one* worker claims a specific task, preventing race conditions.
4. **Worker Cluster (The Brawn):** Horizontally scalable nodes. They use `BRPop` to listen to Redis without wasting CPU. When they get a task, they execute your custom Go functions (Activities).

---

## 🌐 HTTP API Reference

The Orchestrator exposes a simple REST API to interact with the engine.

### 1. Start a Workflow
Triggers a new distributed workflow.
- **Endpoint:** `POST /workflow/start`
- **Response:**
  ```json
  {
    "status": "started",
    "workflow_id": "c82f...a1b2"
  }
  ```

### 2. View Event History (Time Machine)
Because the system is Event Sourced, you can retrieve the exact, chronological audit log of a workflow.
- **Endpoint:** `GET /workflow/history?id=<WORKFLOW_ID>`
- **Response:**
  ```json
  [
    {
      "ID": "event-1",
      "EventType": "WorkflowStarted",
      "Timestamp": "2026-06-25T10:00:00Z"
    },
    {
      "ID": "event-2",
      "EventType": "TaskCompleted",
      "Timestamp": "2026-06-25T10:00:02Z"
    }
  ]
  ```

---

## 🛡 Fault Tolerance & Reliability

1. **Stateless Workers:** If a Worker crashes mid-execution, the Orchestrator notices the task timed out and simply assigns it to another Worker. 
2. **Activity Registry:** You write business logic as standard Go functions. The engine handles the heavy lifting of routing and retry logic automatically.
3. **Dead Letter Queue (DLQ):** If an external API is down, the Worker retries the task automatically. After 3 consecutive failures, it quarantines the task in a Redis DLQ (`chronos:tasks:dlq`) so it doesn't block the system, and saves a `TaskFailed` event for the engineering team to review.
