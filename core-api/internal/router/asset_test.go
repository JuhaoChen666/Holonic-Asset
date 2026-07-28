package router_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/1024XEngineer/Holonic-Asset/internal/dto"
	"github.com/1024XEngineer/Holonic-Asset/internal/router"
	"github.com/1024XEngineer/Holonic-Asset/pkg/echox"
)

type assetRouterStub struct {
	router.AssetRouter
	projectID string
	assetID   string
	request   dto.GetAssetsRequest
	update    dto.UpdateAssetRequest
}

func (s *assetRouterStub) GetAssets(context *echox.Context, request dto.GetAssetsRequest) (dto.Response, error) {
	s.projectID = context.Param("project_id")
	s.request = request
	return dto.NewSuccessResponse(dto.GetAssetsResponse{Assets: []dto.AssetListItemResponse{{AssetID: 7, ProjectID: 42}}}), nil
}

func (s *assetRouterStub) Detail(context *echox.Context) (dto.Response, error) {
	s.assetID = context.Param("asset_id")
	return dto.NewSuccessResponse(dto.AssetDetailResponse{AssetID: 7}), nil
}

func (s *assetRouterStub) Tags(_ *echox.Context, request dto.UpdateAssetRequest) (dto.Response, error) {
	s.update = request
	return dto.NewSuccessResponse(dto.UpdateAssetResponse{AssetID: request.AssetID}), nil
}

func TestAssetRoutesBindPathParameters(t *testing.T) {
	assetStub := &assetRouterStub{}
	e := router.Register(assetStub, nil, nil, nil)

	t.Run("get assets", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/projects/42/assets", nil)
		recorder := httptest.NewRecorder()

		e.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
		}
		if assetStub.projectID != "42" {
			t.Fatalf("expected project ID 42, got %q", assetStub.projectID)
		}
		if recorder.Body.String() != `{"code":200,"message":"success","data":{"assets":[{"assetId":7,"name":"","projectId":42,"type":"","description":"","tags":null,"version":0}]}}`+"\n" {
			t.Fatalf("unexpected response: %s", recorder.Body.String())
		}
	})

	t.Run("binds asset filters", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/projects/42/assets?query=hero&tags=hero&tags=player&types=character", nil)
		recorder := httptest.NewRecorder()

		e.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
		}
		if len(assetStub.request.Tags) != 2 || assetStub.request.Tags[0] != "hero" || assetStub.request.Tags[1] != "player" {
			t.Fatalf("unexpected tag filters: %+v", assetStub.request.Tags)
		}
		if len(assetStub.request.Types) != 1 || assetStub.request.Types[0] != "character" {
			t.Fatalf("unexpected type filters: %+v", assetStub.request.Types)
		}
		if assetStub.request.Query != "hero" {
			t.Fatalf("unexpected query filter: %q", assetStub.request.Query)
		}
	})

	t.Run("binds asset update", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/asset/tags", strings.NewReader(`{
			"assetId": 7,
			"name": "hero",
			"projectId": 42,
			"type": "character",
			"description": "main character",
			"tags": ["player"],
			"attributes": {"scale": 2},
			"version": 3
		}`))
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder := httptest.NewRecorder()

		e.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
		}
		if assetStub.update.AssetID != 7 || assetStub.update.Name == nil || *assetStub.update.Name != "hero" || assetStub.update.Type == nil || *assetStub.update.Type != "character" {
			t.Fatalf("unexpected asset update: %+v", assetStub.update)
		}
	})

	t.Run("asset detail", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/asset/7", nil)
		recorder := httptest.NewRecorder()

		e.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
		}
		if assetStub.assetID != "7" {
			t.Fatalf("expected asset ID 7, got %q", assetStub.assetID)
		}
		if recorder.Body.String() != `{"code":200,"message":"success","data":{"assetId":7,"name":"","projectId":0,"type":"","description":"","tags":null,"attributes":null,"version":0}}`+"\n" {
			t.Fatalf("unexpected response: %s", recorder.Body.String())
		}
	})
}
