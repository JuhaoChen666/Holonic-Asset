package router

import (
	"github.com/labstack/echo/v4"

	"github.com/1024XEngineer/Holonic-Asset/internal/taxonomy/dto"
	"github.com/1024XEngineer/Holonic-Asset/pkg/echox"
)

type TaxonomyRouter interface {
	FindRelatedAssets(
		c *echox.Context,
		request dto.FindRelatedAssetsRequest,
	) (*dto.AssetSearchResult, error)
}

// RegisterRoutes registers the public asset discovery routes.
func RegisterRoutes(e *echo.Group, r TaxonomyRouter) {
	assets := e.Group("/projects/:projectId/assets")
	assets.GET("/:assetId/related", echox.WrapReq(r.FindRelatedAssets))
}
