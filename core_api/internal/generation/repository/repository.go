package repository

import (
	"context"

	"github.com/1024XEngineer/Holonic-Asset/internal/generation/domain"
)

// Reader defines generation queries that do not require a write transaction.
type Reader interface {
	GetRun(ctx context.Context, runID domain.RunID) (*domain.GenerationRun, error)
	GetStep(ctx context.Context, stepID domain.StepID) (*domain.Step, error)
	ListSteps(ctx context.Context, runID domain.RunID) ([]domain.Step, error)
	GetCandidate(ctx context.Context, candidateID domain.CandidateID) (*domain.Candidate, error)
	ListCandidates(ctx context.Context, runID domain.RunID) ([]domain.Candidate, error)
}

// Writer requires expected statuses for mutable lifecycle entities so delayed
// retries cannot overwrite a terminal or otherwise incompatible state.
type Writer interface {
	CreateRun(ctx context.Context, run *domain.GenerationRun) error
	UpdateRun(ctx context.Context, run *domain.GenerationRun, expectedStatus domain.RunStatus) (bool, error)

	// CreatePlan persists the Plan and its initial Steps as one aggregate.
	CreatePlan(ctx context.Context, plan *domain.Plan) error
	UpdateStep(ctx context.Context, step *domain.Step, expectedStatus domain.StepStatus) (bool, error)

	CreateCandidate(ctx context.Context, candidate *domain.Candidate) error
	UpdateCandidate(
		ctx context.Context,
		candidate *domain.Candidate,
		expectedStatus domain.CandidateStatus,
	) (bool, error)
}

// Repository is the transaction-bound persistence contract.
type Repository interface {
	Reader
	Writer
}
