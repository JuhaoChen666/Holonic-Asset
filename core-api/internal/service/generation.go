package service

import (
	"context"
	"errors"

	domain "github.com/1024XEngineer/Holonic-Asset/internal/model/generation"
	"github.com/1024XEngineer/Holonic-Asset/internal/repository"
)

const (
	defaultRunListLimit = 20
	maxRunListLimit     = 100
)

var ErrInvalidRunListStatus = errors.New("generation: invalid run list status")

// RequestService exposes the generation lifecycle use cases used by transports.
type RequestService interface {
	Create(ctx context.Context, request *domain.GenerationRequest) (domain.RunID, error)
	List(ctx context.Context, query *domain.RunListQuery) (*domain.RunListPage, error)
	Get(ctx context.Context, runID domain.RunID) (*domain.GenerationDetail, error)
	Cancel(ctx context.Context, runID domain.RunID) error
}

// PlanningService is called by planning job handlers.
type PlanningService interface {
	Plan(ctx context.Context, runID domain.RunID) (*domain.Plan, error)
}

// StepResultService records generation-owned outputs. Task handlers maintain
// execution status, attempts, retry, and cancellation.
type StepResultService interface {
	RecordResult(ctx context.Context, stepID domain.StepID, result *domain.StepResult) error
}

type GenerationService interface {
	RequestService
	PlanningService
	StepResultService
}

// generationService is the application-service skeleton. Persistence, planning,
// task coordination, and business lifecycle transitions are intentionally deferred.
type generationService struct {
	reader     repository.Reader
	unitOfWork UnitOfWork
	tasks      TaskService
}

func NewGenerationService(
	reader repository.Reader,
	unitOfWork UnitOfWork,
	tasks TaskService,
) GenerationService {
	return &generationService{
		reader:     reader,
		unitOfWork: unitOfWork,
		tasks:      tasks,
	}
}

func (*generationService) Create(context.Context, *domain.GenerationRequest) (domain.RunID, error) {
	return 0, nil
}

func (s *generationService) List(
	ctx context.Context,
	query *domain.RunListQuery,
) (*domain.RunListPage, error) {
	status := query.Status
	if status == "" {
		status = domain.RunListStatusActive
	}
	if status != domain.RunListStatusActive {
		return nil, ErrInvalidRunListStatus
	}

	limit := query.Limit
	if limit <= 0 {
		limit = defaultRunListLimit
	} else if limit > maxRunListLimit {
		limit = maxRunListLimit
	}

	filter := &domain.RunListFilter{
		ProjectID:  query.ProjectID,
		AssetID:    query.AssetID,
		Lifecycles: domain.ActiveRunLifecycles(),
		Limit:      limit,
		Cursor:     query.Cursor,
	}
	if query.AssetID == nil {
		filter.IncludeTaskTypes = domain.ProjectLevelTaskTypes()
	} else {
		filter.ExcludeTaskTypes = domain.ProjectLevelTaskTypes()
	}

	// Generation persistence is intentionally deferred while the module remains
	// an architecture skeleton.
	if s.reader == nil {
		return &domain.RunListPage{Runs: []domain.GenerationRun{}}, nil
	}
	return s.reader.ListRuns(ctx, filter)
}

func (*generationService) Get(context.Context, domain.RunID) (*domain.GenerationDetail, error) {
	return &domain.GenerationDetail{}, nil
}

func (*generationService) Cancel(context.Context, domain.RunID) error {
	return nil
}

func (*generationService) Plan(context.Context, domain.RunID) (*domain.Plan, error) {
	return &domain.Plan{}, nil
}

func (*generationService) RecordResult(context.Context, domain.StepID, *domain.StepResult) error {
	return nil
}

var _ GenerationService = (*generationService)(nil)
