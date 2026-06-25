# Architecture Overview

Chronos is built as a highly decoupled, distributed execution platform designed to mimic the core architecture of systems like Temporal or AWS Step Functions.

## The Core Components
1. **API Gateway & Orchestrator:** The entry point. It translates incoming HTTP requests into durable workflows.
2. **Event Store (PostgreSQL):** The absolute source of truth. Every single state change is recorded here as an immutable event.
3. **Task Queue (Redis):** The distributed router. It holds JSON tasks that are waiting to be executed by the worker cluster.
4. **Worker Cluster:** Horizontally scalable nodes that constantly poll the queue, safely claim tasks, and execute registered Go functions (Activities).

## Why Decoupled?
By separating the Orchestrator, Queue, and Workers, Chronos guarantees incredible fault tolerance. If a Worker crashes mid-execution, the task simply times out and is picked up by another Worker. The Orchestrator doesn't care *who* executes the task, only that it gets done durbably.
