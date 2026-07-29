package task

import (
	"context"
	"testing"
)

type handler struct{}

func (handler) Handle(context.Context, *Task) error { return nil }

func TestRegistryIsBusinessAgnostic(t *testing.T) {
	registry := newRegistry()
	registry.register("example.v1", handler{})

	if err := registry.dispatch(context.Background(), &Task{Type: "example.v1"}); err != nil {
		t.Fatalf("dispatch generic task: %v", err)
	}
}

func TestRegistryDispatchReturnsErrorForUnknownType(t *testing.T) {
	registry := newRegistry()

	if err := registry.dispatch(context.Background(), &Task{Type: "unknown.v1"}); err == nil {
		t.Fatal("expected unknown task type error")
	}
}
