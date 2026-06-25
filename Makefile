.PHONY: run-server run-worker

run-server:
	go run cmd/server/main.go

run-worker:
	go run cmd/worker/main.go
