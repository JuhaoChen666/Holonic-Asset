package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/1024XEngineer/Holonic-Asset/internal/dto"
	domain "github.com/1024XEngineer/Holonic-Asset/internal/model/generation"
	taskdomain "github.com/1024XEngineer/Holonic-Asset/internal/module/task"
	"github.com/1024XEngineer/Holonic-Asset/internal/service"
	"github.com/1024XEngineer/Holonic-Asset/pkg/echox"
)

type GenerationHandler struct {
	service service.GenerationService
}

type createCharacterPrototypeTaskHandler struct{ service service.GenerationService }
type createCharacterAnimationTaskHandler struct{ service service.GenerationService }
type createObjectPrototypeTaskHandler struct{ service service.GenerationService }
type createObjectAnimationTaskHandler struct{ service service.GenerationService }
type createTileSetTaskHandler struct{ service service.GenerationService }
type emptyGenerationTaskHandler struct{ service service.GenerationService }

func NewGenerationHandler(generationService service.GenerationService) *GenerationHandler {
	return &GenerationHandler{service: generationService}
}

func (h *createCharacterPrototypeTaskHandler) Handle(
	ctx context.Context,
	message *taskdomain.Task,
) error {
	var payload domain.CreateCharacterPrototypePayload
	if err := decodeGenerationTaskPayload(message, &payload); err != nil {
		return err
	}

	return h.service.Process(ctx, message)
}

func (h *createCharacterAnimationTaskHandler) Handle(
	ctx context.Context,
	message *taskdomain.Task,
) error {
	var payload domain.CreateCharacterAnimationPayload
	if err := decodeGenerationTaskPayload(message, &payload); err != nil {
		return err
	}

	return h.service.Process(ctx, message)
}

func (h *createObjectPrototypeTaskHandler) Handle(
	ctx context.Context,
	message *taskdomain.Task,
) error {
	var payload domain.CreateObjectPrototypePayload
	if err := decodeGenerationTaskPayload(message, &payload); err != nil {
		return err
	}

	return h.service.Process(ctx, message)
}

func (h *createObjectAnimationTaskHandler) Handle(
	ctx context.Context,
	message *taskdomain.Task,
) error {
	var payload domain.CreateObjectAnimationPayload
	if err := decodeGenerationTaskPayload(message, &payload); err != nil {
		return err
	}

	return h.service.Process(ctx, message)
}

func (h *createTileSetTaskHandler) Handle(
	ctx context.Context,
	message *taskdomain.Task,
) error {
	var payload domain.CreateTileSetPayload
	if err := decodeGenerationTaskPayload(message, &payload); err != nil {
		return err
	}

	return h.service.Process(ctx, message)
}

func (h *emptyGenerationTaskHandler) Handle(
	ctx context.Context,
	message *taskdomain.Task,
) error {
	var payload struct{}
	if err := decodeGenerationTaskPayload(message, &payload); err != nil {
		return err
	}

	return h.service.Process(ctx, message)
}

func decodeGenerationTaskPayload(message *taskdomain.Task, payload any) error {
	if message == nil {
		return service.ErrGenerationTaskRequired
	}
	if err := json.Unmarshal(message.Payload, payload); err != nil {
		return fmt.Errorf("generation: decode %s task %d payload: %w", message.Type, message.ID, err)
	}
	return nil
}

func RegisterGenerationTaskHandlers(registry *taskdomain.Registry, h *GenerationHandler) {
	registry.Register(string(domain.GenerateCharacterProtoType),
		&createCharacterPrototypeTaskHandler{service: h.service})
	registry.Register(string(domain.GenerateCharacterAnimation),
		&createCharacterAnimationTaskHandler{service: h.service})
	registry.Register(string(domain.GenerateObjectProtoType),
		&createObjectPrototypeTaskHandler{service: h.service})
	registry.Register(string(domain.GenerateObjectAnimation),
		&createObjectAnimationTaskHandler{service: h.service})
	registry.Register(string(domain.GenerateTileSet),
		&createTileSetTaskHandler{service: h.service})

	emptyHandler := &emptyGenerationTaskHandler{service: h.service}
	for _, taskType := range []domain.TaskType{
		domain.RegenerateCharacterProtoType,
		domain.RegenerateCharacterAnimation,
		domain.RegenerateCharacterFrames,
		domain.RegenerateObjectProtoType,
		domain.RegenerateObjectAnimation,
		domain.RegenerateObjectFrames,
		domain.RegenerateItem,
		domain.RegenerateTiles,
	} {
		registry.Register(string(taskType), emptyHandler)
	}
}

func (h *GenerationHandler) Create(
	ctx *echox.Context,
	request dto.CreateGenerationRequest,
) (dto.CreateGenerationResponse, error) {
	runID, err := h.service.Create(ctx, &domain.GenerationRequest{
		ProjectID:         request.ProjectID,
		AssetID:           request.AssetID,
		Kind:              request.Kind,
		Prompt:            request.Prompt,
		ReferenceMediaIDs: request.ReferenceMediaIDs,
		TargetAssetPaths:  request.TargetAssetPaths,
		Parameters:        request.Parameters,
	})
	return dto.CreateGenerationResponse{GenerationRunID: runID}, err
}

func (h *GenerationHandler) List(
	ctx *echox.Context,
	request dto.ListGenerationRunsRequest,
) (dto.ListGenerationRunsResponse, error) {
	page, err := h.service.List(ctx, &domain.RunListQuery{
		ProjectID: request.ProjectID,
		AssetID:   request.AssetID,
		Status:    request.Status,
		Limit:     request.Limit,
		Cursor:    request.Cursor,
	})
	if errors.Is(err, service.ErrInvalidRunListStatus) {
		return dto.ListGenerationRunsResponse{}, echo.ErrBadRequest
	}
	if err != nil {
		return dto.ListGenerationRunsResponse{}, err
	}
	if page == nil {
		return dto.ListGenerationRunsResponse{Items: []dto.GenerationRunListItemResponse{}}, nil
	}

	items := make([]dto.GenerationRunListItemResponse, len(page.Runs))
	for i := range page.Runs {
		run := page.Runs[i]
		items[i] = dto.GenerationRunListItemResponse{
			ID:        run.ID,
			ProjectID: run.ProjectID,
			AssetID:   run.AssetID,
			Kind:      run.Kind,
			Status:    run.Status,
		}
	}

	return dto.ListGenerationRunsResponse{
		Items:      items,
		NextCursor: page.NextCursor,
	}, nil
}

func (h *GenerationHandler) Get(
	ctx *echox.Context,
	request dto.GetGenerationRequest,
) (dto.GetGenerationResponse, error) {
	run, err := h.service.Get(ctx, request.GenerationRunID)
	if err != nil {
		return dto.GetGenerationResponse{}, err
	}

	return dto.GetGenerationResponse{
		ID:        run.ID,
		ProjectID: run.ProjectID,
		AssetID:   run.AssetID,
		Kind:      run.Kind,
		Status:    run.Status,
		Result:    run.Result,
		Error:     run.Error,
	}, nil
}

func (h *GenerationHandler) Cancel(
	ctx *echox.Context,
	request dto.CancelGenerationRequest,
) (dto.CancelGenerationResponse, error) {
	err := h.service.Cancel(ctx, request.GenerationRunID)
	return dto.CancelGenerationResponse{Cancelled: err == nil}, err
}

var (
	_ taskdomain.Handler = (*createCharacterPrototypeTaskHandler)(nil)
	_ taskdomain.Handler = (*createCharacterAnimationTaskHandler)(nil)
	_ taskdomain.Handler = (*createObjectPrototypeTaskHandler)(nil)
	_ taskdomain.Handler = (*createObjectAnimationTaskHandler)(nil)
	_ taskdomain.Handler = (*createTileSetTaskHandler)(nil)
	_ taskdomain.Handler = (*emptyGenerationTaskHandler)(nil)
)
