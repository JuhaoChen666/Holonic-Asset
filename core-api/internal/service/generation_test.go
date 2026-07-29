package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	domain "github.com/1024XEngineer/Holonic-Asset/internal/model/generation"
	taskdomain "github.com/1024XEngineer/Holonic-Asset/internal/module/task"
	"github.com/1024XEngineer/Holonic-Asset/internal/repository"
	"github.com/1024XEngineer/Holonic-Asset/internal/service"
)

type generationReaderStub struct {
	filter *domain.RunListFilter
	page   *domain.RunListPage
}

func (s *generationReaderStub) ListRuns(
	_ context.Context,
	filter *domain.RunListFilter,
) (*domain.RunListPage, error) {
	s.filter = filter
	return s.page, nil
}

var _ repository.GenerationTaskReader = (*generationReaderStub)(nil)

type taskServiceStub struct {
	createdTask   *taskdomain.Task
	createID      uint
	detail        *taskdomain.Task
	statusUpdates []taskStatusUpdate
	resultTaskID  uint
	result        json.RawMessage
}

type taskStatusUpdate struct {
	taskID uint
	status taskdomain.Status
}

func (s *taskServiceStub) Create(_ context.Context, message *taskdomain.Task) (uint, error) {
	s.createdTask = message
	return s.createID, nil
}

func (s *taskServiceStub) GetDetail(context.Context, uint) (*taskdomain.Task, error) {
	return s.detail, nil
}

func (s *taskServiceStub) UpdateStatus(_ context.Context, taskID uint, status taskdomain.Status) error {
	s.statusUpdates = append(s.statusUpdates, taskStatusUpdate{taskID: taskID, status: status})
	return nil
}

func (s *taskServiceStub) UpdateResult(_ context.Context, taskID uint, result json.RawMessage) error {
	s.resultTaskID = taskID
	s.result = result
	return nil
}

func TestCreateBuildsOneTaskFromGenerationRequest(t *testing.T) {
	assetID := uint(9)
	tasks := &taskServiceStub{createID: 17}
	generationService := service.NewGenerationService(nil, tasks, nil)
	request := &domain.GenerationRequest{
		ProjectID:         42,
		AssetID:           &assetID,
		Kind:              domain.GenerateCharacterAnimation,
		Prompt:            "walk",
		ReferenceMediaIDs: []string{"media-1"},
	}

	runID, err := generationService.Create(context.Background(), request)
	if err != nil {
		t.Fatalf("create generation: %v", err)
	}
	if runID != 17 || tasks.createdTask == nil {
		t.Fatalf("unexpected task creation: run=%d task=%+v", runID, tasks.createdTask)
	}
	if tasks.createdTask.Type != string(request.Kind) ||
		tasks.createdTask.Status != taskdomain.StatusPending {
		t.Fatalf("unexpected task envelope: %+v", tasks.createdTask)
	}

	var payload domain.TaskPayload
	if err := json.Unmarshal(tasks.createdTask.Payload, &payload); err != nil {
		t.Fatalf("decode task payload: %v", err)
	}
	if payload.ProjectID != request.ProjectID || payload.AssetID == nil ||
		*payload.AssetID != assetID || payload.Prompt != request.Prompt ||
		!reflect.DeepEqual(payload.ReferenceMediaIDs, request.ReferenceMediaIDs) {
		t.Fatalf("unexpected task payload: %+v", payload)
	}
}

func TestCreateProjectLevelTaskPayloadOmitsAssetID(t *testing.T) {
	tasks := &taskServiceStub{createID: 17}
	generationService := service.NewGenerationService(nil, tasks, nil)

	_, err := generationService.Create(context.Background(), &domain.GenerationRequest{
		ProjectID: 42,
		Kind:      domain.GenerateCharacterProtoType,
		Prompt:    "hero",
	})
	if err != nil {
		t.Fatalf("create generation: %v", err)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(tasks.createdTask.Payload, &payload); err != nil {
		t.Fatalf("decode task payload: %v", err)
	}
	if _, ok := payload["asset_id"]; ok {
		t.Fatalf("project-level payload must omit asset_id: %s", tasks.createdTask.Payload)
	}
}

func TestGetProjectsTaskAsGenerationRun(t *testing.T) {
	assetID := uint(9)
	payload, err := json.Marshal(domain.TaskPayload{ProjectID: 42, AssetID: &assetID})
	if err != nil {
		t.Fatalf("encode task payload: %v", err)
	}
	tasks := &taskServiceStub{detail: &taskdomain.Task{
		ID:      17,
		Type:    string(domain.GenerateCharacterAnimation),
		Status:  taskdomain.StatusProcessing,
		Payload: payload,
		Result:  json.RawMessage(`{"asset_id":9}`),
	}}
	generationService := service.NewGenerationService(nil, tasks, nil)

	run, err := generationService.Get(context.Background(), 17)
	if err != nil {
		t.Fatalf("get generation: %v", err)
	}
	if run.ID != 17 || run.ProjectID != 42 || run.AssetID == nil || *run.AssetID != assetID ||
		run.Kind != domain.GenerateCharacterAnimation || run.Status != taskdomain.StatusProcessing {
		t.Fatalf("unexpected generation run: %+v", run)
	}
}

func TestListBuildsProjectScopeTaskFilter(t *testing.T) {
	reader := &generationReaderStub{page: &domain.RunListPage{}}
	generationService := service.NewGenerationService(reader, &taskServiceStub{}, nil)

	_, err := generationService.List(context.Background(), &domain.RunListQuery{
		ProjectID: 42,
		Status:    domain.RunListStatusActive,
		Limit:     10,
		Cursor:    "cursor",
	})
	if err != nil {
		t.Fatalf("list generation runs: %v", err)
	}
	if reader.filter == nil || reader.filter.ProjectID != 42 || reader.filter.AssetID != nil ||
		reader.filter.Limit != 10 || reader.filter.Cursor != "cursor" {
		t.Fatalf("unexpected filter: %+v", reader.filter)
	}
	if !reflect.DeepEqual(reader.filter.Statuses, domain.ActiveTaskStatuses()) {
		t.Fatalf("unexpected statuses: %v", reader.filter.Statuses)
	}
	if !reflect.DeepEqual(reader.filter.IncludeTaskTypes, domain.ProjectLevelTaskTypes()) {
		t.Fatalf("unexpected project task types: %v", reader.filter.IncludeTaskTypes)
	}
}

func TestListBuildsAssetScopeTaskFilter(t *testing.T) {
	assetID := uint(9)
	reader := &generationReaderStub{page: &domain.RunListPage{}}
	generationService := service.NewGenerationService(reader, &taskServiceStub{}, nil)

	_, err := generationService.List(context.Background(), &domain.RunListQuery{
		ProjectID: 42,
		AssetID:   &assetID,
	})
	if err != nil {
		t.Fatalf("list generation runs: %v", err)
	}
	if reader.filter == nil || reader.filter.AssetID == nil || *reader.filter.AssetID != assetID {
		t.Fatalf("unexpected asset filter: %+v", reader.filter)
	}
	if len(reader.filter.IncludeTaskTypes) != 0 ||
		!reflect.DeepEqual(reader.filter.ExcludeTaskTypes, domain.ProjectLevelTaskTypes()) {
		t.Fatalf("unexpected task type filter: %+v", reader.filter)
	}
	if reader.filter.Limit != 20 {
		t.Fatalf("expected default limit 20, got %d", reader.filter.Limit)
	}
}

func TestListRejectsUnsupportedStatus(t *testing.T) {
	generationService := service.NewGenerationService(&generationReaderStub{}, &taskServiceStub{}, nil)
	_, err := generationService.List(context.Background(), &domain.RunListQuery{Status: "completed"})
	if !errors.Is(err, service.ErrInvalidRunListStatus) {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}

func TestCancelUpdatesTaskStatus(t *testing.T) {
	tasks := &taskServiceStub{}
	generationService := service.NewGenerationService(nil, tasks, nil)

	if err := generationService.Cancel(context.Background(), 17); err != nil {
		t.Fatalf("cancel generation: %v", err)
	}
	if len(tasks.statusUpdates) != 1 || tasks.statusUpdates[0].taskID != 17 ||
		tasks.statusUpdates[0].status != taskdomain.StatusCancelled {
		t.Fatalf("unexpected status updates: %+v", tasks.statusUpdates)
	}
}

type generationAIServiceStub struct {
	taskType domain.TaskType
	payload  json.RawMessage
	result   json.RawMessage
	err      error
}

func (s *generationAIServiceStub) Generate(
	_ context.Context,
	taskType domain.TaskType,
	payload json.RawMessage,
) (json.RawMessage, error) {
	s.taskType = taskType
	s.payload = payload
	return s.result, s.err
}

func TestProcessPassesPayloadToAIAndCompletesTask(t *testing.T) {
	payload := json.RawMessage(`{"project_id":42,"prompt":"hero"}`)
	result := json.RawMessage(`{"asset_id":7}`)
	tasks := &taskServiceStub{}
	ai := &generationAIServiceStub{result: result}
	generationService := service.NewGenerationService(nil, tasks, ai)

	err := generationService.Process(context.Background(), &taskdomain.Task{
		ID:      17,
		Type:    string(domain.GenerateCharacterProtoType),
		Payload: payload,
	})
	if err != nil {
		t.Fatalf("handle generation task: %v", err)
	}
	if ai.taskType != domain.GenerateCharacterProtoType || !reflect.DeepEqual(ai.payload, payload) {
		t.Fatalf("unexpected AI request: type=%s payload=%s", ai.taskType, ai.payload)
	}
	if len(tasks.statusUpdates) != 1 || tasks.statusUpdates[0].status != taskdomain.StatusProcessing {
		t.Fatalf("unexpected status updates: %+v", tasks.statusUpdates)
	}
	if tasks.resultTaskID != 17 || !reflect.DeepEqual(tasks.result, result) {
		t.Fatalf("unexpected task result: id=%d result=%s", tasks.resultTaskID, tasks.result)
	}
}

func TestProcessMarksTaskFailedWhenAIFails(t *testing.T) {
	wantErr := errors.New("AI failed")
	tasks := &taskServiceStub{}
	generationService := service.NewGenerationService(nil, tasks, &generationAIServiceStub{err: wantErr})

	err := generationService.Process(context.Background(), &taskdomain.Task{ID: 17})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected AI error, got %v", err)
	}
	if len(tasks.statusUpdates) != 2 ||
		tasks.statusUpdates[0].status != taskdomain.StatusProcessing ||
		tasks.statusUpdates[1].status != taskdomain.StatusFailed {
		t.Fatalf("unexpected status updates: %+v", tasks.statusUpdates)
	}
}
