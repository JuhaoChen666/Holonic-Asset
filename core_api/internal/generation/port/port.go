// Package port defines module dependencies used by generation orchestration.
package port

import (
	"context"

	"github.com/1024XEngineer/Holonic-Asset/internal/generation/domain"
)

// Scheduler hides the task library and River job identifiers from generation.
type Scheduler interface {
	SchedulePlanning(ctx context.Context, runID domain.RunID) error
	ScheduleStep(ctx context.Context, runID domain.RunID, stepID domain.StepID, executor domain.StepExecutor) error
	CancelRun(ctx context.Context, runID domain.RunID) error
}

// CandidateConfirmer hands an accepted candidate to the asset module without
// exposing Asset persistence or version internals to generation.
type CandidateConfirmer interface {
	Confirm(ctx context.Context, runID domain.RunID, candidate *domain.Candidate) error
}
