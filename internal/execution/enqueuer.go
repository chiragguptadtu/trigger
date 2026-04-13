package execution

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

// NewRiverClient creates a River client with the execution worker registered.
// Call client.Start(ctx) after creation to begin processing jobs.
func NewRiverClient(pool *pgxpool.Pool, worker *Worker) (*river.Client[pgx.Tx], error) {
	workers := river.NewWorkers()
	river.AddWorker(workers, worker)

	return river.NewClient(riverpgxv5.New(pool), &river.Config{
		Workers: workers,
		Queues:  map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: 10}},
	})
}

// RiverEnqueuer wraps a River client and satisfies handler.JobEnqueuer.
type RiverEnqueuer struct {
	client *river.Client[pgx.Tx]
}

func NewRiverEnqueuer(client *river.Client[pgx.Tx]) *RiverEnqueuer {
	return &RiverEnqueuer{client: client}
}

func (e *RiverEnqueuer) Enqueue(ctx context.Context, executionID string) error {
	_, err := e.client.Insert(ctx, ExecutionArgs{ExecutionID: executionID}, nil)
	if err != nil {
		return fmt.Errorf("enqueue execution %s: %w", executionID, err)
	}
	return nil
}
