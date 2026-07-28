package service

import (
	"context"

	"github.com/1024XEngineer/Holonic-Asset/pkg/storage"
)

// ProjectPreviewUploadService creates direct upload targets for Project previews.
type ProjectPreviewUploadService interface {
	CreateProjectPreviewUploadTarget(
		ctx context.Context,
		request *CreateProjectPreviewUploadRequest,
	) (*ProjectPreviewUploadTarget, error)
}

// mediaService provides the Project preview upload application skeleton.
type mediaService struct {
	storage storage.Storage
}

// NewMediaService creates the Media application service used by the HTTP handler.
func NewMediaService(objectStorage storage.Storage) ProjectPreviewUploadService {
	return &mediaService{storage: objectStorage}
}

func (*mediaService) CreateProjectPreviewUploadTarget(
	context.Context,
	*CreateProjectPreviewUploadRequest,
) (*ProjectPreviewUploadTarget, error) {
	return &ProjectPreviewUploadTarget{}, nil
}

var _ ProjectPreviewUploadService = (*mediaService)(nil)
