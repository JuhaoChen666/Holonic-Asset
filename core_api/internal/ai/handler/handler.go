package handler

import (
	"github.com/1024XEngineer/Holonic-Asset/internal/ai/dto"
	"github.com/1024XEngineer/Holonic-Asset/internal/ai/router"
	"github.com/1024XEngineer/Holonic-Asset/internal/ai/service"
	"github.com/1024XEngineer/Holonic-Asset/pkg/echox"
)

type AIHandler struct {
	service service.AIService
}

func NewAIHandler(aiService service.AIService) *AIHandler {
	return &AIHandler{service: aiService}
}

func (h *AIHandler) GenerateCharacter(
	c *echox.Context,
	request dto.GenerateCharacterRequest,
) (*dto.GenerateCharacterResponse, error) {
	return h.service.GenerateCharacter(c, &request)
}

func (h *AIHandler) EditCharacter(
	c *echox.Context,
	request dto.EditCharacterRequest,
) (*dto.EditCharacterResponse, error) {
	return h.service.EditCharacter(c, &request)
}

func (h *AIHandler) GenerateProjectPreview(
	c *echox.Context,
	request dto.GenerateProjectPreviewRequest,
) (*dto.GenerateProjectPreviewResponse, error) {
	return h.service.GenerateProjectPreview(c, &request)
}

func (h *AIHandler) GenerateTileSetItem(
	c *echox.Context,
	request dto.GenerateTileSetItemRequest,
) (*dto.GenerateTileSetItemResponse, error) {
	return h.service.GenerateTileSetItem(c, &request)
}

func (h *AIHandler) EditTileSetItem(
	c *echox.Context,
	request dto.EditTileSetItemRequest,
) (*dto.EditTileSetItemResponse, error) {
	return h.service.EditTileSetItem(c, &request)
}

func (h *AIHandler) GenerateObject(
	c *echox.Context,
	request dto.GenerateObjectRequest,
) (*dto.GenerateObjectResponse, error) {
	return h.service.GenerateObject(c, &request)
}

func (h *AIHandler) EditObject(
	c *echox.Context,
	request dto.EditObjectRequest,
) (*dto.EditObjectResponse, error) {
	return h.service.EditObject(c, &request)
}

func (h *AIHandler) GenerateSceneryLayer(
	c *echox.Context,
	request dto.GenerateSceneryLayerRequest,
) (*dto.GenerateSceneryLayerResponse, error) {
	return h.service.GenerateSceneryLayer(c, &request)
}

func (h *AIHandler) EditSceneryLayer(
	c *echox.Context,
	request dto.EditSceneryLayerRequest,
) (*dto.EditSceneryLayerResponse, error) {
	return h.service.EditSceneryLayer(c, &request)
}

func (h *AIHandler) GenerateAnimation(
	c *echox.Context,
	request dto.GenerateAnimationRequest,
) (*dto.GenerateAnimationResponse, error) {
	return h.service.GenerateAnimation(c, &request)
}

func (h *AIHandler) EditFrame(
	c *echox.Context,
	request dto.EditFrameRequest,
) (*dto.EditFrameResponse, error) {
	return h.service.EditFrame(c, &request)
}

func (h *AIHandler) GenerateUI(
	c *echox.Context,
	request dto.GenerateUIRequest,
) (*dto.GenerateUIResponse, error) {
	return h.service.GenerateUI(c, &request)
}

func (h *AIHandler) EditUIComponent(
	c *echox.Context,
	request dto.EditUIComponentRequest,
) (*dto.EditUIComponentResponse, error) {
	return h.service.EditUIComponent(c, &request)
}

var _ router.AIRouter = (*AIHandler)(nil)
