package router

import (
	"github.com/labstack/echo/v4"

	"github.com/1024XEngineer/Holonic-Asset/internal/media/dto"
	"github.com/1024XEngineer/Holonic-Asset/pkg/echox"
)

type MediaRouter interface {
	CreateUploadTarget(
		c *echox.Context,
		request dto.CreateMediaUploadRequest,
	) (*dto.ObjectUploadTarget, error)
}

// RegisterRoutes registers all Media HTTP routes.
func RegisterRoutes(e *echo.Group, r MediaRouter) {
	media := e.Group("/media")
	media.POST("/upload-target", echox.WrapReq(r.CreateUploadTarget))
}
