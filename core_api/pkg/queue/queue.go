package queue

import "context"

type Job interface {
	Kind() string
}

type Handler interface {
	JobKind() string
	Handle(ctx context.Context, payload []byte) error
}

type Publisher interface {
	Publish(ctx context.Context, job Job) (int64, error)
}
