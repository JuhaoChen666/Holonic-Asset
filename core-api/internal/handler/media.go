package handler

import (
	"github.com/1024XEngineer/Holonic-Asset/internal/dto"
	"github.com/1024XEngineer/Holonic-Asset/internal/service"
	"github.com/1024XEngineer/Holonic-Asset/pkg/echox"
)

type MediaHandler struct {
	service service.ProjectPreviewUploadService
}

func NewMediaHandler(projectPreviewUploadService service.ProjectPreviewUploadService) *MediaHandler {
	return &MediaHandler{service: projectPreviewUploadService}
}

func (h *MediaHandler) CreateProjectPreviewUploadTarget(
	c *echox.Context,
	request dto.CreateProjectPreviewUploadRequest,
) (*dto.ProjectPreviewUploadTarget, error) {
	target, err := h.service.CreateProjectPreviewUploadTarget(c, &service.CreateProjectPreviewUploadRequest{
		ContentType:   request.ContentType,
		ContentLength: request.ContentLength,
	})
	if err != nil {
		return nil, err
	}
	return &dto.ProjectPreviewUploadTarget{
		ObjectKey:   target.ObjectKey,
		ObjectURL:   target.ObjectURL,
		UploadURL:   target.UploadURL,
		UploadToken: target.UploadToken,
	}, nil
}
