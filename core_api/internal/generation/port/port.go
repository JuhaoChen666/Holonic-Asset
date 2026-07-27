// Package port defines module dependencies used by generation orchestration.
package port

import (
	"context"

	"github.com/1024XEngineer/Holonic-Asset/internal/generation/domain"
	"github.com/1024XEngineer/Holonic-Asset/internal/generation/repository"
)

// Scheduler hides the task library and River job identifiers from generation.
type Scheduler interface {
	SchedulePlanning(ctx context.Context, runID domain.RunID) error
	ScheduleStep(ctx context.Context, runID domain.RunID, stepID domain.StepID, executor domain.StepExecutor) error
	CancelRun(ctx context.Context, runID domain.RunID) error
}

// Transaction exposes repository and scheduling operations bound to the same
// database transaction.
type Transaction interface {
	Repository() repository.Repository
	Scheduler() Scheduler
	CandidateConfirmer() CandidateConfirmer
}

type UnitOfWork interface {
	WithinTransaction(ctx context.Context, work func(context.Context, Transaction) error) error
}

// CandidateConfirmer hands an accepted candidate to the asset module without
// exposing Asset persistence or version internals to generation.
type CandidateConfirmer interface {
	Confirm(ctx context.Context, command *domain.ConfirmCandidateCommand) error
}
