# Retry Engine & Dead Letter Queue (DLQ)

Failures are inevitable in distributed systems (network timeouts, API rate limits, database locks). A durable engine must handle them gracefully.

## Automatic Retries
If an Activity returns an `error` in Chronos, the Worker does not crash. Instead, it:
1. Catches the error.
2. Increments the task's `RetryCount`.
3. Instantly re-queues the task in Redis (`LPush`).

Another worker will eventually pick it up and try again.

## Dead Letter Queue (DLQ)
If a task fails 3 times consecutively, it is considered a "Poison Pill". 
Instead of infinitely retrying and crashing the system in a loop, Chronos quarantines the task by pushing it to a dedicated Redis list called `chronos:tasks:dlq` using the `PushDLQ` method. 

It then records a final `TaskFailed` event in Postgres, ensuring the engineering team has a perfect audit log of the failure and the quarantined task data.
