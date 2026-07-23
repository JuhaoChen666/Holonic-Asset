package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/1024XEngineer/Holonic-Asset/internal/media/dto"
	"github.com/1024XEngineer/Holonic-Asset/internal/media/handler"
	"github.com/1024XEngineer/Holonic-Asset/pkg/echox"
)

type mediaUploadServiceStub struct {
	target  *dto.ObjectUploadTarget
	err     error
	ctx     context.Context
	request *dto.CreateMediaUploadRequest
	calls   int
}

func (s *mediaUploadServiceStub) CreateUploadTarget(
	ctx context.Context,
	request *dto.CreateMediaUploadRequest,
) (*dto.ObjectUploadTarget, error) {
	s.calls++
	s.ctx = ctx
	s.request = request
	return s.target, s.err
}

func TestCreateUploadTargetForwardsRequestToService(t *testing.T) {
	wantTarget := &dto.ObjectUploadTarget{
		ObjectKey: "assets/42/resources/99/source.png",
		UploadURL: "https://storage.example/upload",
	}
	stub := &mediaUploadServiceStub{target: wantTarget}
	mediaHandler := handler.NewMediaHandler(stub)
	request := dto.CreateMediaUploadRequest{
		AssetID:         42,
		AssetResourceID: 99,
		ContentType:     "image/png",
	}
	handlerContext := newHandlerContext()

	target, err := mediaHandler.CreateUploadTarget(handlerContext, request)
	if err != nil {
		t.Fatalf("create upload target: %v", err)
	}
	if stub.calls != 1 {
		t.Fatalf("expected one service call, got %d", stub.calls)
	}
	if stub.ctx != handlerContext {
		t.Fatal("expected handler context to be forwarded to the service")
	}
	if stub.request == nil || *stub.request != request {
		t.Fatalf("expected request %+v, got %+v", request, stub.request)
	}
	if target != wantTarget {
		t.Fatalf("expected target pointer %p, got %p", wantTarget, target)
	}
}

func TestCreateUploadTargetPropagatesServiceError(t *testing.T) {
	wantErr := errors.New("create upload target failed")
	stub := &mediaUploadServiceStub{err: wantErr}
	mediaHandler := handler.NewMediaHandler(stub)

	target, err := mediaHandler.CreateUploadTarget(newHandlerContext(), dto.CreateMediaUploadRequest{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
	if target != nil {
		t.Fatalf("expected a nil target, got %+v", target)
	}
}

func newHandlerContext() *echox.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	recorder := httptest.NewRecorder()
	return &echox.Context{Context: e.NewContext(req, recorder)}
}
