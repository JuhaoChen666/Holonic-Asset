package router_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/1024XEngineer/Holonic-Asset/internal"
	"github.com/1024XEngineer/Holonic-Asset/internal/generation/handler"
	"github.com/1024XEngineer/Holonic-Asset/internal/generation/service"
)

func TestGenerationRoutesReturnPlaceholderResponses(t *testing.T) {
	generationService := service.NewGenerationService()
	generationHandler := handler.NewGenerationHandler(generationService)
	e := internal.Register(nil, nil, generationHandler, nil, nil)

	tests := []struct {
		method string
		path   string
		body   string
		want   string
	}{
		{
			method: http.MethodPost,
			path:   "/api/v1/projects/42/generation-runs",
			body:   `{"kind":"generate_character","prompt":"hero"}`,
			want:   `{"generationRunId":0}` + "\n",
		},
		{
			method: http.MethodGet,
			path:   "/api/v1/generation-runs/7",
			want:   `{"id":0,"projectId":0,"kind":"","status":"","steps":null,"candidates":null}` + "\n",
		},
		{
			method: http.MethodPost,
			path:   "/api/v1/generation-runs/7/cancel",
			body:   `{}`,
			want:   `{"cancelled":false}` + "\n",
		},
		{
			method: http.MethodPost,
			path:   "/api/v1/generation-runs/7/candidates/9/confirm",
			body:   `{}`,
			want:   `{"confirmed":false}` + "\n",
		},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			recorder := httptest.NewRecorder()

			e.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
			}
			if recorder.Body.String() != test.want {
				t.Fatalf("unexpected placeholder response: %s", recorder.Body.String())
			}
		})
	}
}

func TestAIRoutesAreNotExposed(t *testing.T) {
	generationService := service.NewGenerationService()
	generationHandler := handler.NewGenerationHandler(generationService)
	e := internal.Register(nil, nil, generationHandler, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/tile-set/item/edit", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, recorder.Code, recorder.Body.String())
	}
}
