# Distributed Task Queue

Chronos uses Redis as an ultra-fast, distributed task router.

## The Push/Pop Model
When the Orchestrator starts a workflow, it `LPush`es a JSON task into the `chronos:tasks` list.

Workers use `BRPop` (Blocking Right Pop) to listen to this list. `BRPop` is a critical distributed systems pattern: it blocks the Redis connection safely without consuming CPU cycles until a task is available. 

When a task arrives, Redis guarantees that exactly *one* worker claims it, preventing race conditions where multiple workers process the same payment.
