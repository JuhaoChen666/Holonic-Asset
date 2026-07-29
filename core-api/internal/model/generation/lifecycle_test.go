package generation_test

import (
	"reflect"
	"testing"

	domain "github.com/1024XEngineer/Holonic-Asset/internal/model/generation"
)

func TestRunLifecycleContract(t *testing.T) {
	got := []domain.RunLifecycle{
		domain.RunLifecycleAccepted,
		domain.RunLifecyclePlanning,
		domain.RunLifecyclePlanned,
		domain.RunLifecycleGenerating,
		domain.RunLifecyclePostProcessing,
		domain.RunLifecycleCompleted,
		domain.RunLifecycleFailed,
		domain.RunLifecycleCancelled,
	}
	want := []domain.RunLifecycle{
		"accepted",
		"planning",
		"planned",
		"generating",
		"post_processing",
		"completed",
		"failed",
		"cancelled",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected run lifecycle contract: %v", got)
	}
}
