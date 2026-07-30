package repository

import (
	"context"

	"github.com/1024XEngineer/Holonic-Asset/internal/config"
	"github.com/1024XEngineer/Holonic-Asset/internal/module/upload"
)

// QiniuStorage is the Qiniu Object Storage implementation skeleton.
type QiniuStorage struct{}

// NewQiniuStorage creates a Qiniu-backed storage service.
// Qiniu client initialization is intentionally deferred.
func NewQiniuStorage(config.QiniuConfig) (*QiniuStorage, error) {
	return &QiniuStorage{}, nil
}

// CreateUploadTarget creates a temporary Qiniu upload target.
// Upload token generation is intentionally deferred.
func (*QiniuStorage) CreateUploadTarget(context.Context, upload.UploadRequest) (*upload.UploadTarget, error) {
	return &upload.UploadTarget{}, nil
}

// GetObjectMetadata retrieves metadata for an object.
// Qiniu resource access is intentionally deferred.
func (*QiniuStorage) GetObjectMetadata(context.Context, string) (*upload.ObjectMetadata, error) {
	return &upload.ObjectMetadata{}, nil
}

// DeleteObject deletes an object.
// Qiniu resource access is intentionally deferred.
func (*QiniuStorage) DeleteObject(context.Context, string) error {
	return nil
}

var _ upload.Store = (*QiniuStorage)(nil)
