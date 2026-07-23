// Package river implements the queue.Port interfaces using River as the
// job queue backend. This is the ONLY package (along with ioc) that imports
// River directly — business modules never see these types.
package river

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"github.com/1024XEngineer/Holonic-Asset/pkg/queue"
)

// riverJobArgs is a single concrete type that satisfies river.JobArgs for
// ALL business job types. It carries the original job's kind and its
// JSON-serialized payload so that a single River worker can route to any
// number of queue.Handler implementations.
type riverJobArgs struct {
	// KindName is the job kind, used by River to route jobs to workers.
	KindName string `json:"kind"`
	// Payload is the original business job serialized as JSON.
	Payload json.RawMessage `json:"payload"`
}

func (r riverJobArgs) Kind() string { return r.KindName }

// riverWorker is a single River Worker that dispatches to the appropriate
// queue.Handler based on the job kind. It allows any number of business
// handlers to be registered through one River worker type.
type riverWorker struct {
	river.WorkerDefaults[riverJobArgs]
	handlers map[string]queue.Handler
}

func (w *riverWorker) Work(ctx context.Context, job *river.Job[riverJobArgs]) error {
	handler, ok := w.handlers[job.Args.KindName]
	if !ok {
		return fmt.Errorf("river adapter: no handler registered for job kind %q", job.Args.KindName)
	}
	return handler.Handle(ctx, []byte(job.Args.Payload))
}

// Publisher implements queue.Publisher using a River client.
type Publisher struct {
	client *river.Client[pgx.Tx]
}

// NewPublisher creates a Publisher backed by a River client.
func NewPublisher(client *river.Client[pgx.Tx]) *Publisher {
	return &Publisher{client: client}
}

// Publish enqueues a business job by wrapping it in a riverJobArgs and
// inserting it into River.
func (p *Publisher) Publish(ctx context.Context, job queue.Job) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("river adapter: failed to marshal job %q: %w", job.Kind(), err)
	}

	args := riverJobArgs{
		KindName: job.Kind(),
		Payload:  json.RawMessage(payload),
	}

	_, err = p.client.Insert(ctx, args, nil)
	if err != nil {
		return fmt.Errorf("river adapter: failed to insert job %q: %w", job.Kind(), err)
	}
	return nil
}

// BuildClient creates a fully configured River client from a set of business
// handlers. Each handler is registered under its JobKind so that incoming
// jobs are dispatched correctly.
//
// This is the only place where river.AddWorker is called — all handler
// registration goes through a single riverWorker instance that routes by kind.
func BuildClient(
	ctx context.Context,
	dbPool *pgxpool.Pool,
	config *river.Config,
	handlers ...queue.Handler,
) (*river.Client[pgx.Tx], error) {
	workers := river.NewWorkers()

	if len(handlers) > 0 {
		w := &riverWorker{
			handlers: make(map[string]queue.Handler, len(handlers)),
		}
		for _, h := range handlers {
			kind := h.JobKind()
			if _, exists := w.handlers[kind]; exists {
				return nil, fmt.Errorf("river adapter: duplicate handler for job kind %q", kind)
			}
			w.handlers[kind] = h
		}
		river.AddWorker(workers, w)
	}

	config.Workers = workers

	client, err := river.NewClient(riverpgxv5.New(dbPool), config)
	if err != nil {
		return nil, fmt.Errorf("river adapter: failed to create client: %w", err)
	}

	return client, nil
}
