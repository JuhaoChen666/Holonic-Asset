package service

import (
	"context"

	"github.com/1024XEngineer/Holonic-Asset/internal/generation/domain"
	"github.com/1024XEngineer/Holonic-Asset/internal/generation/port"
	"github.com/1024XEngineer/Holonic-Asset/internal/generation/repository"
)

// RequestService exposes the generation lifecycle use cases used by transports.
type RequestService interface {
	Create(ctx context.Context, request *domain.GenerationRequest) (domain.RunID, error)
	Get(ctx context.Context, runID domain.RunID) (*domain.GenerationDetail, error)
	Cancel(ctx context.Context, runID domain.RunID) error
	ConfirmCandidate(ctx context.Context, runID domain.RunID, candidateID domain.CandidateID) error
}

// PlanningService is called by planning job handlers.
type PlanningService interface {
	Plan(ctx context.Context, runID domain.RunID) (*domain.Plan, error)
}

// StepService is called by job handlers to maintain authoritative Step state.
type StepService interface {
	Start(ctx context.Context, stepID domain.StepID) (*domain.Step, error)
	Complete(ctx context.Context, stepID domain.StepID, result *domain.StepResult) error
	Fail(ctx context.Context, stepID domain.StepID, failure *domain.Failure) error
}

type GenerationService interface {
	RequestService
	PlanningService
	StepService
}

// generationService is the application-service skeleton. Persistence, planning,
// task coordination, retry, and state-transition behavior are intentionally deferred.
type generationService struct {
	reader     repository.Reader
	unitOfWork port.UnitOfWork
}

func NewGenerationService(
	reader repository.Reader,
	unitOfWork port.UnitOfWork,
) GenerationService {
	return &generationService{
		reader:     reader,
		unitOfWork: unitOfWork,
	}
}

func (*generationService) Create(context.Context, *domain.GenerationRequest) (domain.RunID, error) {
	return 0, nil
}

func (*generationService) Get(context.Context, domain.RunID) (*domain.GenerationDetail, error) {
	return &domain.GenerationDetail{}, nil
}

func (*generationService) Cancel(context.Context, domain.RunID) error {
	return nil
}

func (*generationService) ConfirmCandidate(context.Context, domain.RunID, domain.CandidateID) error {
	return nil
}

func (*generationService) Plan(context.Context, domain.RunID) (*domain.Plan, error) {
	return &domain.Plan{}, nil
}

func (*generationService) Start(context.Context, domain.StepID) (*domain.Step, error) {
	return &domain.Step{}, nil
}

func (*generationService) Complete(context.Context, domain.StepID, *domain.StepResult) error {
	return nil
}

func (*generationService) Fail(context.Context, domain.StepID, *domain.Failure) error {
	return nil
}

var _ GenerationService = (*generationService)(nil)
