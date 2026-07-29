package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/1024XEngineer/Holonic-Asset/internal/dto"

	"github.com/labstack/echo/v4"

	"github.com/1024XEngineer/Holonic-Asset/internal/handler"
	domain "github.com/1024XEngineer/Holonic-Asset/internal/model/generation"
	taskdomain "github.com/1024XEngineer/Holonic-Asset/internal/module/task"
	"github.com/1024XEngineer/Holonic-Asset/internal/service"
	"github.com/1024XEngineer/Holonic-Asset/pkg/echox"
)

type requestServiceStub struct {
	createRequest *domain.GenerationRequest
	listQuery     *domain.RunListQuery
	listPage      *domain.RunListPage
	listErr       error
	detail        *domain.GenerationDetail
	cancelErr     error
}

func (s *requestServiceStub) Create(
	_ context.Context,
	request *domain.GenerationRequest,
) (domain.RunID, error) {
	s.createRequest = request
	return 17, nil
}

func (s *requestServiceStub) List(
	_ context.Context,
	query *domain.RunListQuery,
) (*domain.RunListPage, error) {
	s.listQuery = query
	if s.listPage == nil {
		return &domain.RunListPage{}, s.listErr
	}
	return s.listPage, s.listErr
}

func (s *requestServiceStub) Get(
	context.Context,
	domain.RunID,
) (*domain.GenerationDetail, error) {
	return s.detail, nil
}

func (s *requestServiceStub) Cancel(context.Context, domain.RunID) error {
	return s.cancelErr
}

func TestCreateMapsTransportRequest(t *testing.T) {
	stub := &requestServiceStub{}
	generationHandler := handler.NewGenerationHandler(stub)
	parameters := json.RawMessage(`{"size":{"width":64,"height":64}}`)
	request := dto.CreateGenerationRequest{
		ProjectID:         2,
		AssetID:           3,
		Kind:              domain.RequestKindGenerateCharacter,
		Prompt:            "hero",
		ReferenceMediaIDs: []string{"media-1"},
		TargetAssetPaths:  []string{"animations.walk.directions.left"},
		Parameters:        parameters,
	}

	response, err := generationHandler.Create(newGenerationHandlerContext(), request)
	if err != nil {
		t.Fatalf("create generation: %v", err)
	}
	if response.GenerationRunID != 17 {
		t.Fatalf("expected run ID 17, got %d", response.GenerationRunID)
	}
	if stub.createRequest == nil {
		t.Fatal("expected generation request")
	}
	if stub.createRequest.ProjectID != request.ProjectID ||
		stub.createRequest.AssetID != request.AssetID ||
		stub.createRequest.Kind != request.Kind ||
		stub.createRequest.Prompt != request.Prompt ||
		!reflect.DeepEqual(stub.createRequest.ReferenceMediaIDs, request.ReferenceMediaIDs) ||
		!reflect.DeepEqual(stub.createRequest.TargetAssetPaths, request.TargetAssetPaths) ||
		!reflect.DeepEqual(stub.createRequest.Parameters, request.Parameters) {
		t.Fatalf("unexpected generation request: %+v", stub.createRequest)
	}
}

func TestGetMapsGenerationDetail(t *testing.T) {
	taskStatus := taskdomain.StatusProcessing
	stub := &requestServiceStub{detail: &domain.GenerationDetail{
		Run: domain.GenerationRun{
			ID:        7,
			ProjectID: 2,
			Lifecycle: domain.RunLifecycleGenerating,
			Request: domain.GenerationRequest{
				Kind: domain.RequestKindGenerateCharacter,
			},
		},
		Steps: []domain.StepDetail{{
			Step: domain.Step{
				ID:           8,
				Type:         "generate_image",
				Executor:     domain.StepExecutorAI,
				Dependencies: []domain.StepID{6},
			},
			TaskStatus: &taskStatus,
		}},
	}}
	generationHandler := handler.NewGenerationHandler(stub)

	response, err := generationHandler.Get(
		newGenerationHandlerContext(),
		dto.GetGenerationRequest{GenerationRunID: 7},
	)
	if err != nil {
		t.Fatalf("get generation: %v", err)
	}
	if response.ID != 7 || response.ProjectID != 2 ||
		response.Kind != domain.RequestKindGenerateCharacter ||
		response.Lifecycle != domain.RunLifecycleGenerating {
		t.Fatalf("unexpected run response: %+v", response)
	}
	if len(response.Steps) != 1 || response.Steps[0].ID != 8 ||
		response.Steps[0].TaskStatus == nil ||
		*response.Steps[0].TaskStatus != taskdomain.StatusProcessing {
		t.Fatalf("unexpected steps response: %+v", response.Steps)
	}
}

func TestListMapsQueryAndRuns(t *testing.T) {
	assetID := uint(3)
	stub := &requestServiceStub{listPage: &domain.RunListPage{
		Runs: []domain.GenerationRun{{
			ID:        7,
			ProjectID: 2,
			Request: domain.GenerationRequest{
				AssetID: assetID,
				Kind:    domain.RequestKindGenerateCharacter,
			},
			Lifecycle: domain.RunLifecycleGenerating,
		}},
		NextCursor: "next",
	}}
	generationHandler := handler.NewGenerationHandler(stub)

	query := dto.ListGenerationRunsRequest{
		ProjectID: 42,
		AssetID:   &assetID,
		Status:    domain.RunListStatusActive,
		Limit:     10,
		Cursor:    "cursor",
	}
	response, err := generationHandler.List(newGenerationHandlerContext(), query)
	if err != nil {
		t.Fatalf("list generation runs: %v", err)
	}
	if stub.listQuery == nil || stub.listQuery.ProjectID != query.ProjectID ||
		stub.listQuery.AssetID == nil || *stub.listQuery.AssetID != assetID ||
		stub.listQuery.Status != query.Status || stub.listQuery.Limit != query.Limit ||
		stub.listQuery.Cursor != query.Cursor {
		t.Fatalf("unexpected list query: %+v", stub.listQuery)
	}
	if len(response.Items) != 1 || response.Items[0].ID != 7 ||
		response.Items[0].AssetID != assetID ||
		response.Items[0].Kind != domain.RequestKindGenerateCharacter ||
		response.Items[0].Lifecycle != domain.RunLifecycleGenerating ||
		response.NextCursor != "next" {
		t.Fatalf("unexpected list response: %+v", response)
	}
}

func TestListRejectsUnsupportedStatus(t *testing.T) {
	stub := &requestServiceStub{listErr: service.ErrInvalidRunListStatus}
	_, err := handler.NewGenerationHandler(stub).List(
		newGenerationHandlerContext(),
		dto.ListGenerationRunsRequest{Status: "completed"},
	)
	if !errors.Is(err, echo.ErrBadRequest) {
		t.Fatalf("expected bad request, got %v", err)
	}
}

func TestCommandResponsesReflectServiceResult(t *testing.T) {
	wantErr := errors.New("command failed")

	tests := []struct {
		name    string
		stub    *requestServiceStub
		invoke  func(*handler.GenerationHandler) (bool, error)
		want    bool
		wantErr error
	}{
		{
			name: "cancel success",
			stub: &requestServiceStub{},
			invoke: func(h *handler.GenerationHandler) (bool, error) {
				response, err := h.Cancel(newGenerationHandlerContext(), dto.CancelGenerationRequest{GenerationRunID: 7})
				return response.Cancelled, err
			},
			want: true,
		},
		{
			name: "cancel failure",
			stub: &requestServiceStub{cancelErr: wantErr},
			invoke: func(h *handler.GenerationHandler) (bool, error) {
				response, err := h.Cancel(newGenerationHandlerContext(), dto.CancelGenerationRequest{GenerationRunID: 7})
				return response.Cancelled, err
			},
			wantErr: wantErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.invoke(handler.NewGenerationHandler(test.stub))
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("expected error %v, got %v", test.wantErr, err)
			}
			if got != test.want {
				t.Fatalf("expected result %t, got %t", test.want, got)
			}
		})
	}
}

func newGenerationHandlerContext() *echox.Context {
	e := echo.New()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	return &echox.Context{Context: e.NewContext(request, httptest.NewRecorder())}
}
