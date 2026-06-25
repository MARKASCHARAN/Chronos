package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}

	q, err := queue.NewRedisQueue(redisAddr, "", 0)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	w := worker.NewWorker(store, q)

	w.RegisterActivity("charge_customer", func(ctx context.Context, payload []byte) ([]byte, error) {
		log.Printf("Executing charge_customer with payload: %s\n", string(payload))
		
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
		log.Println("Received termination signal, initiating graceful shutdown...")
		cancel()
	}()

	w.Start(ctx)
}
