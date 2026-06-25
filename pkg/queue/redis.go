package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	defaultQueueName = "chronos:tasks"
	dlqName          = "chronos:tasks:dlq"
)

// RedisQueue implements the Queue interface using Redis.
type RedisQueue struct {
	client    *redis.Client
	queueName string
}

// NewRedisQueue creates a new Redis-backed task queue.
func NewRedisQueue(addr, password string, db int) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &RedisQueue{
		client:    client,
		queueName: defaultQueueName,
	}, nil
}

// Push implements the Queue interface.
func (q *RedisQueue) Push(ctx context.Context, task Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to serialize task: %w", err)
	}

	if err := q.client.LPush(ctx, q.queueName, data).Err(); err != nil {
		return fmt.Errorf("failed to push task to redis: %w", err)
	}

	return nil
}

// PushDLQ implements the Queue interface by pushing irrecoverable tasks to a Dead Letter Queue.
func (q *RedisQueue) PushDLQ(ctx context.Context, task Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to serialize task for DLQ: %w", err)
	}

	if err := q.client.LPush(ctx, dlqName, data).Err(); err != nil {
		return fmt.Errorf("failed to push task to DLQ redis: %w", err)
	}

	return nil
}

// Pop implements the Queue interface.
func (q *RedisQueue) Pop(ctx context.Context) (*Task, error) {
	result, err := q.client.BRPop(ctx, 0, q.queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to pop task from redis: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid response from redis")
	}

	var task Task
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, fmt.Errorf("failed to deserialize task: %w", err)
	}

	return &task, nil
}
