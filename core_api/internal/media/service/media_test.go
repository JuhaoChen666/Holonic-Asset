package service_test

import (
	"context"
	"testing"

	"github.com/1024XEngineer/Holonic-Asset/internal/media/dto"
	"github.com/1024XEngineer/Holonic-Asset/internal/media/service"
)

func TestCreateUploadTargetReturnsPlaceholderTarget(t *testing.T) {
	mediaService := service.NewMediaUploadService()
	request := &dto.CreateMediaUploadRequest{
		AssetID:         42,
		AssetResourceID: 99,
		ContentType:     "image/png",
	}

	target, err := mediaService.CreateUploadTarget(context.Background(), request)
	if err != nil {
		t.Fatalf("create upload target: %v", err)
	}
	if target == nil {
		t.Fatal("expected a non-nil placeholder target")
	}
	if target.ObjectKey != "" || target.UploadURL != "" {
		t.Fatalf("expected an empty placeholder target, got %+v", target)
	}
}
