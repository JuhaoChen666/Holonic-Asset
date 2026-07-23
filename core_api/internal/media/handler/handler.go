package handler

import (
	"github.com/1024XEngineer/Holonic-Asset/internal/media/dto"
	"github.com/1024XEngineer/Holonic-Asset/internal/media/router"
	"github.com/1024XEngineer/Holonic-Asset/internal/media/service"
	"github.com/1024XEngineer/Holonic-Asset/pkg/echox"
)

type MediaHandler struct {
	service service.MediaUploadService
}

func NewMediaHandler(mediaUploadService service.MediaUploadService) *MediaHandler {
	return &MediaHandler{service: mediaUploadService}
}

func (h *MediaHandler) CreateUploadTarget(
	c *echox.Context,
	request dto.CreateMediaUploadRequest,
) (*dto.ObjectUploadTarget, error) {
	return h.service.CreateUploadTarget(c, &request)
}

var _ router.MediaRouter = (*MediaHandler)(nil)
