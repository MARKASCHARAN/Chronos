package eventstore

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/markasaicharan/chronos/internal/domain"
)

// PostgresStore implements the Store interface using PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new PostgreSQL-backed event store.
func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PostgresStore{db: db}

	if err := store.createTables(); err != nil {
		return nil, err
	}

	return store, nil
}

// createTables sets up the event sourcing table schema.
func (s *PostgresStore) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS events (
		id VARCHAR(255) PRIMARY KEY,
		workflow_id VARCHAR(255) NOT NULL,
		event_type VARCHAR(100) NOT NULL,
		payload BYTEA,
		timestamp TIMESTAMP WITH TIME ZONE NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_workflow_id ON events(workflow_id);
	`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create events table: %w", err)
	}
	return nil
}

// SaveEvent implements the Store interface.
func (s *PostgresStore) SaveEvent(ctx context.Context, event domain.Event) error {
	query := `
		INSERT INTO events (id, workflow_id, event_type, payload, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.ExecContext(ctx, query,
		event.ID,
		event.WorkflowID,
		event.EventType,
		event.Payload,
		event.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to save event %s: %w", event.ID, err)
	}
	return nil
}

// GetEvents implements the Store interface.
func (s *PostgresStore) GetEvents(ctx context.Context, workflowID string) ([]domain.Event, error) {
	query := `
		SELECT id, workflow_id, event_type, payload, timestamp 
		FROM events 
		WHERE workflow_id = $1 
		ORDER BY timestamp ASC
	`
	rows, err := s.db.QueryContext(ctx, query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events for workflow %s: %w", workflowID, err)
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var e domain.Event
		if err := rows.Scan(&e.ID, &e.WorkflowID, &e.EventType, &e.Payload, &e.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return events, nil
}
