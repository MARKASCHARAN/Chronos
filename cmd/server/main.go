package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/markasaicharan/chronos/internal/workflow"
	"github.com/markasaicharan/chronos/pkg/eventstore"
	"github.com/markasaicharan/chronos/pkg/queue"
)

func main() {
	log.Println("Starting Chronos Workflow Server on :8080...")

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

	orchestrator := workflow.NewOrchestrator(store, q)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Chronos is running"))
	})

	http.HandleFunc("/workflow/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		payload := []byte(`{"action": "charge_customer", "amount": 99.99}`)

		workflowID, err := orchestrator.StartWorkflow(context.Background(), "DemoWorkflow", payload)
		if err != nil {
			log.Printf("Failed to start workflow: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		log.Printf("Started new workflow via API: %s\n", workflowID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":      "started",
			"workflow_id": workflowID,
		})
	})

	http.HandleFunc("/workflow/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		workflowID := r.URL.Query().Get("id")
		if workflowID == "" {
			http.Error(w, "Missing workflow id", http.StatusBadRequest)
			return
		}

		events, err := store.GetEvents(context.Background(), workflowID)
		if err != nil {
			log.Printf("Failed to get events: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
