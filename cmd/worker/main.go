package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/markasaicharan/chronos/internal/worker"
	"github.com/markasaicharan/chronos/pkg/eventstore"
	"github.com/markasaicharan/chronos/pkg/queue"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://chronos:chronos_password@localhost:5432/chronos_db?sslmode=disable"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	store, err := eventstore.NewPostgresStore(dsn)
	if err != nil {
		log.Fatal("Failed to connect to Postgres", "error", err)
	}

	q, err := queue.NewRedisQueue(redisAddr, "", 0)
	if err != nil {
		log.Fatal("Failed to connect to Redis", "error", err)
	}

	w := worker.NewWorker(store, q)

	w.RegisterActivity("charge_customer", func(ctx context.Context, payload []byte) ([]byte, error) {
		log.Info("Executing charge_customer", "payload", string(payload))

		// Simulate a flaky network that fails 70% of the time to demonstrate the Retry & DLQ engine
		if time.Now().UnixNano()%10 < 7 {
			return nil, fmt.Errorf("connection reset by peer (simulated stripe failure)")
		}

		return []byte(`{"status": "charged", "transaction_id": "tx_999"}`), nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Received termination signal, initiating graceful shutdown...")
		cancel()
	}()

	w.Start(ctx)
}
