package service

import (
	"context"
	"encoding/json"
	"errors"

	domain "github.com/1024XEngineer/Holonic-Asset/internal/model/generation"
	taskdomain "github.com/1024XEngineer/Holonic-Asset/internal/module/task"
	"github.com/1024XEngineer/Holonic-Asset/internal/repository"
)

const (
	defaultRunListLimit = 20
	maxRunListLimit     = 100
)

var (
	ErrInvalidRunListStatus   = errors.New("generation: invalid run list status")
	ErrTaskServiceRequired    = errors.New("generation: task service is required")
	ErrAIServiceRequired      = errors.New("generation: AI service is required")
	ErrGenerationTaskRequired = errors.New("generation: task is required")
)

// GenerationAIService owns image generation and the resulting asset creation.
// Payload is the same generation-owned JSON stored on the task record.
type GenerationAIService interface {
	Generate(
		ctx context.Context,
		taskType domain.TaskType,
		payload json.RawMessage,
	) (json.RawMessage, error)
}

// RequestService exposes task-backed generation use cases used by transports.
type RequestService interface {
	Create(ctx context.Context, request *domain.GenerationRequest) (domain.RunID, error)
	List(ctx context.Context, query *domain.RunListQuery) (*domain.RunListPage, error)
	Get(ctx context.Context, runID domain.RunID) (*domain.GenerationRun, error)
	Cancel(ctx context.Context, runID domain.RunID) error
}

type GenerationService interface {
	RequestService
	Process(ctx context.Context, message *taskdomain.Task) error
}

// generationService is the application boundary over the generic task module.
// AI Service performs generation and owns any asset creation required by it.
type generationService struct {
	reader repository.GenerationTaskReader
	tasks  TaskService
	ai     GenerationAIService
}

func NewGenerationService(
	reader repository.GenerationTaskReader,
	tasks TaskService,
	ai GenerationAIService,
) GenerationService {
	return &generationService{
		reader: reader,
		tasks:  tasks,
		ai:     ai,
	}
}

func (s *generationService) Create(
	ctx context.Context,
	request *domain.GenerationRequest,
) (domain.RunID, error) {
	if s.tasks == nil {
		return 0, ErrTaskServiceRequired
	}

	payload, err := json.Marshal(domain.NewTaskPayload(request))
	if err != nil {
		return 0, err
	}

	taskID, err := s.tasks.Create(ctx, &taskdomain.Task{
		Type:    string(request.Kind),
		Status:  taskdomain.StatusPending,
		Payload: payload,
	})
	return domain.RunID(taskID), err
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
		ProjectID: query.ProjectID,
		AssetID:   query.AssetID,
		Statuses:  domain.ActiveTaskStatuses(),
		Limit:     limit,
		Cursor:    query.Cursor,
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

func (s *generationService) Get(ctx context.Context, runID domain.RunID) (*domain.GenerationRun, error) {
	if s.tasks == nil {
		return nil, ErrTaskServiceRequired
	}

	message, err := s.tasks.GetDetail(ctx, uint(runID))
	if err != nil {
		return nil, err
	}

	var payload domain.TaskPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return nil, err
	}

	return &domain.GenerationRun{
		ID:        domain.RunID(message.ID),
		ProjectID: payload.ProjectID,
		AssetID:   payload.AssetID,
		Kind:      domain.TaskType(message.Type),
		Status:    message.Status,
		Result:    message.Result,
		Error:     message.Error,
	}, nil
}

func (s *generationService) Cancel(ctx context.Context, runID domain.RunID) error {
	if s.tasks == nil {
		return ErrTaskServiceRequired
	}
	return s.tasks.UpdateStatus(ctx, uint(runID), taskdomain.StatusCancelled)
}

func (s *generationService) Process(ctx context.Context, message *taskdomain.Task) error {
	if message == nil {
		return ErrGenerationTaskRequired
	}
	if s.tasks == nil {
		return ErrTaskServiceRequired
	}
	if s.ai == nil {
		return ErrAIServiceRequired
	}
	if err := s.tasks.UpdateStatus(ctx, message.ID, taskdomain.StatusProcessing); err != nil {
		return err
	}

	result, err := s.ai.Generate(ctx, domain.TaskType(message.Type), message.Payload)
	if err != nil {
		if updateErr := s.tasks.UpdateStatus(ctx, message.ID, taskdomain.StatusFailed); updateErr != nil {
			return errors.Join(err, updateErr)
		}
		return err
	}

	return s.tasks.UpdateResult(ctx, message.ID, result)
}

var _ GenerationService = (*generationService)(nil)
