package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Starting Chronos Workflow Server on :8080...")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Chronos is running"))
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
