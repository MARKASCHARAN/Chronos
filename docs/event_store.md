# Event Store & Event Sourcing

Traditional databases use "CRUD" (Create, Read, Update, Delete) to overwrite a record (e.g., `status = 'completed'`). This permanently destroys the historical context of how the system arrived at that state.

Chronos uses **Event Sourcing** with PostgreSQL.

## How it works
Every action generates a strongly-typed `Event`:
1. `EventTypeWorkflowStarted`
2. `EventTypeTaskCompleted`
3. `EventTypeTaskFailed`

These events are appended to the `events` table. Because the log is append-only and never updated, it serves as a perfect, immutable audit trail.

## The Time Machine API
Because we store a timeline rather than a snapshot, you can hit `GET /workflow/history?id=<WORKFLOW_ID>` to instantly retrieve the full chronological history of any workflow.
