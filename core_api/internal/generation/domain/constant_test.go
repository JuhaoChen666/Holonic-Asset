package domain_test

import (
	"reflect"
	"testing"

	"github.com/1024XEngineer/Holonic-Asset/internal/generation/domain"
)

func TestRunStatusContract(t *testing.T) {
	got := []domain.RunStatus{
		domain.RunStatusPending,
		domain.RunStatusPlanning,
		domain.RunStatusPlanned,
		domain.RunStatusRunning,
		domain.RunStatusPostProcessing,
		domain.RunStatusWaitingConfirmation,
		domain.RunStatusCompleted,
		domain.RunStatusFailed,
		domain.RunStatusCancelled,
	}
	want := []domain.RunStatus{
		"pending",
		"planning",
		"planned",
		"running",
		"post_processing",
		"waiting_confirmation",
		"completed",
		"failed",
		"cancelled",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected run status contract: %v", got)
	}
}

func TestStepStatusContract(t *testing.T) {
	got := []domain.StepStatus{
		domain.StepStatusPending,
		domain.StepStatusReady,
		domain.StepStatusRunning,
		domain.StepStatusSucceeded,
		domain.StepStatusFailed,
		domain.StepStatusRetryWait,
		domain.StepStatusCancelled,
		domain.StepStatusSkipped,
	}
	want := []domain.StepStatus{
		"pending",
		"ready",
		"running",
		"succeeded",
		"failed",
		"retry_wait",
		"cancelled",
		"skipped",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected step status contract: %v", got)
	}
}
