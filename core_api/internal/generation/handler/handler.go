package handler

import (
	"github.com/1024XEngineer/Holonic-Asset/internal/generation/domain"
	"github.com/1024XEngineer/Holonic-Asset/internal/generation/dto"
	"github.com/1024XEngineer/Holonic-Asset/internal/generation/router"
	"github.com/1024XEngineer/Holonic-Asset/internal/generation/service"
	"github.com/1024XEngineer/Holonic-Asset/pkg/echox"
)

type GenerationHandler struct {
	service service.RequestService
}

func NewGenerationHandler(generationService service.RequestService) *GenerationHandler {
	return &GenerationHandler{service: generationService}
}

func (h *GenerationHandler) Create(
	ctx *echox.Context,
	request dto.CreateGenerationRequest,
) (dto.CreateGenerationResponse, error) {
	runID, err := h.service.Create(ctx, &domain.GenerationRequest{
		ProjectID:              request.ProjectID,
		AssetID:                request.AssetID,
		Kind:                   request.Kind,
		Prompt:                 request.Prompt,
		ReferenceMediaIDs:      request.ReferenceMediaIDs,
		TargetAssetResourceIDs: request.TargetAssetResourceIDs,
		Parameters:             request.Parameters,
	})
	return dto.CreateGenerationResponse{GenerationRunID: runID}, err
}

func (h *GenerationHandler) Get(
	ctx *echox.Context,
	request dto.GetGenerationRequest,
) (dto.GetGenerationResponse, error) {
	_, err := h.service.Get(ctx, request.GenerationRunID)
	return dto.GetGenerationResponse{}, err
}

func (h *GenerationHandler) Cancel(
	ctx *echox.Context,
	request dto.CancelGenerationRequest,
) (dto.CancelGenerationResponse, error) {
	err := h.service.Cancel(ctx, request.GenerationRunID)
	return dto.CancelGenerationResponse{}, err
}

func (h *GenerationHandler) ConfirmCandidate(
	ctx *echox.Context,
	request dto.ConfirmCandidateRequest,
) (dto.ConfirmCandidateResponse, error) {
	err := h.service.ConfirmCandidate(ctx, request.GenerationRunID, request.CandidateID)
	return dto.ConfirmCandidateResponse{}, err
}

var _ router.GenerationRouter = (*GenerationHandler)(nil)
