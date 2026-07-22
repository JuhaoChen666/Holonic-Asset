package router

import (
	"github.com/labstack/echo/v4"

	"github.com/1024XEngineer/Holonic-Asset/internal/ai/dto"
	"github.com/1024XEngineer/Holonic-Asset/pkg/echox"
)

type AIRouter interface {
	GenerateCharacter(*echox.Context, dto.GenerateCharacterRequest) (*dto.GenerateCharacterResponse, error)
	EditCharacter(*echox.Context, dto.EditCharacterRequest) (*dto.EditCharacterResponse, error)
	GenerateProjectPreview(*echox.Context, dto.GenerateProjectPreviewRequest) (*dto.GenerateProjectPreviewResponse, error)
	GenerateTileSetItem(*echox.Context, dto.GenerateTileSetItemRequest) (*dto.GenerateTileSetItemResponse, error)
	EditTileSetItem(*echox.Context, dto.EditTileSetItemRequest) (*dto.EditTileSetItemResponse, error)
	GenerateObject(*echox.Context, dto.GenerateObjectRequest) (*dto.GenerateObjectResponse, error)
	EditObject(*echox.Context, dto.EditObjectRequest) (*dto.EditObjectResponse, error)
	GenerateSceneryLayer(*echox.Context, dto.GenerateSceneryLayerRequest) (*dto.GenerateSceneryLayerResponse, error)
	EditSceneryLayer(*echox.Context, dto.EditSceneryLayerRequest) (*dto.EditSceneryLayerResponse, error)
	GenerateAnimation(*echox.Context, dto.GenerateAnimationRequest) (*dto.GenerateAnimationResponse, error)
	EditFrame(*echox.Context, dto.EditFrameRequest) (*dto.EditFrameResponse, error)
	GenerateUI(*echox.Context, dto.GenerateUIRequest) (*dto.GenerateUIResponse, error)
	EditUIComponent(*echox.Context, dto.EditUIComponentRequest) (*dto.EditUIComponentResponse, error)
}

// RegisterRoutes registers all AI HTTP routes.
func RegisterRoutes(e *echo.Group, r AIRouter) {
	ai := e.Group("/ai")
	ai.POST("/character/generate", echox.WrapReq(r.GenerateCharacter))
	ai.POST("/character/edit", echox.WrapReq(r.EditCharacter))
	ai.POST("/project-preview/generate", echox.WrapReq(r.GenerateProjectPreview))
	ai.POST("/tile-set/item/generate", echox.WrapReq(r.GenerateTileSetItem))
	ai.POST("/tile-set/item/edit", echox.WrapReq(r.EditTileSetItem))
	ai.POST("/object/generate", echox.WrapReq(r.GenerateObject))
	ai.POST("/object/edit", echox.WrapReq(r.EditObject))
	ai.POST("/scenery/layer/generate", echox.WrapReq(r.GenerateSceneryLayer))
	ai.POST("/scenery/layer/edit", echox.WrapReq(r.EditSceneryLayer))
	ai.POST("/animation/generate", echox.WrapReq(r.GenerateAnimation))
	ai.POST("/animation/frame/edit", echox.WrapReq(r.EditFrame))
	ai.POST("/ui/generate", echox.WrapReq(r.GenerateUI))
	ai.POST("/ui/component/edit", echox.WrapReq(r.EditUIComponent))
}
