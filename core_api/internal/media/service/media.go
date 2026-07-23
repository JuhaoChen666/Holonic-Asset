package service

import (
	"context"

	"github.com/1024XEngineer/Holonic-Asset/internal/media/dto"
)

// MediaUploadService defines direct upload-target creation.
type MediaUploadService interface {
	CreateUploadTarget(
		ctx context.Context,
		request *dto.CreateMediaUploadRequest,
	) (*dto.ObjectUploadTarget, error)
}

type mediaUploadService struct{}

func NewMediaUploadService() MediaUploadService {
	return &mediaUploadService{}
}

func (*mediaUploadService) CreateUploadTarget(
	context.Context,
	*dto.CreateMediaUploadRequest,
) (*dto.ObjectUploadTarget, error) {
	return &dto.ObjectUploadTarget{}, nil
}

var _ MediaUploadService = (*mediaUploadService)(nil)
