package repository

import (
	"context"

	"github.com/1024XEngineer/Holonic-Asset/internal/generation/domain"
)

// Repository defines persistence required by the generation application layer.
// Concrete database behavior is intentionally deferred.
type Repository interface {
	CreateRun(ctx context.Context, run *domain.GenerationRun) error
	GetRun(ctx context.Context, runID domain.RunID) (*domain.GenerationRun, error)
	SaveRun(ctx context.Context, run *domain.GenerationRun) error

	SavePlan(ctx context.Context, plan *domain.Plan) error
	GetStep(ctx context.Context, stepID domain.StepID) (*domain.Step, error)
	SaveStep(ctx context.Context, step *domain.Step) error
	ListSteps(ctx context.Context, runID domain.RunID) ([]domain.Step, error)

	SaveCandidate(ctx context.Context, candidate *domain.Candidate) error
	GetCandidate(ctx context.Context, candidateID domain.CandidateID) (*domain.Candidate, error)
}
