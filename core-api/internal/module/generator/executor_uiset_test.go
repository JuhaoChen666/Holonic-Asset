package generator

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/1024XEngineer/Holonic-Asset/internal/module/generator/llmclient"
	assetdomain "github.com/1024XEngineer/Holonic-Asset/internal/module/workspace/asset"
)

type uiSetLLMStub struct {
	request *llmclient.CompletionRequest
	result  *llmclient.CompletionResult
	err     error
}

func (s *uiSetLLMStub) Complete(
	_ context.Context,
	request *llmclient.CompletionRequest,
) (*llmclient.CompletionResult, error) {
	s.request = request
	return s.result, s.err
}

func TestPlanUISetComponentsUsesContextAndReturnsRequestOrder(t *testing.T) {
	llm := &uiSetLLMStub{result: &llmclient.CompletionResult{JSON: json.RawMessage(`{
			"components":[
				{"request_index":0,"name":"Inventory Panel","description":"main item grid container","kind":"panel","states":["default"],"size":{"width":640,"height":480}},
				{"request_index":1,"name":"Close Button","description":"icon-only close control","kind":"button","states":["normal","hover","pressed"],"size":{"width":64,"height":64}},
				{"request_index":-1,"name":"Health Hearts","description":"player vitality indicator","kind":"indicator","states":["full","damaged","empty"],"size":{"width":32,"height":32}}
			]
		}`)}}
	executor := &executor{llm: llm}
	payload := validUISetPlanningPayload()

	plans, err := executor.planUISetComponents(context.Background(), payload)
	if err != nil {
		t.Fatalf("plan UI Set components: %v", err)
	}
	if len(plans) != 3 || plans[0].Index != 0 || plans[0].Name != "Inventory Panel" ||
		plans[0].Description != "main item grid container" || plans[0].Size.Width != 640 ||
		plans[1].Index != 1 || plans[1].Name != "Close Button" || plans[1].Size.Height != 64 ||
		plans[2].Name != "Health Hearts" || !reflect.DeepEqual(plans[2].States, []string{"full", "damaged", "empty"}) {
		t.Fatalf("unexpected UI Set plan: %+v", plans)
	}
	if llm.request == nil || len(llm.request.Images) != 0 ||
		llm.request.ResponseSchema.Name != uiSetComponentPlanSchemaName ||
		!reflect.DeepEqual(llm.request.ResponseSchema.Schema, uiSetComponentPlanJSONSchema) {
		t.Fatalf("unexpected planning request: %+v", llm.request)
	}
	for _, required := range []string{
		"Fantasy Inventory", "compact fantasy inventory", "ornate brass", "Moon Forge", "RPG", "PC", "moonlit pixel art",
		"inventory-driven adventure", `<canvas width="1024" height="768" />`,
		`<component request_index="0"><name>Inventory Panel</name><description>main item grid container</description></component>`,
		"add useful inferred Components", "one horizontal sprite strip", "bar is an engine-composited frame",
	} {
		if !strings.Contains(llm.request.Prompt, required) {
			t.Fatalf("planning prompt omitted %q: %s", required, llm.request.Prompt)
		}
	}
}

func TestDecodeUISetComponentPlanRejectsInvalidPlans(t *testing.T) {
	definitions := validUISetPlanningPayload().Components
	canvas := assetdomain.Size{Width: 1024, Height: 768}
	tests := []struct {
		name string
		raw  string
	}{
		{"missing components", `{}`},
		{"missing component", `{"components":[{"request_index":0,"name":"Inventory Panel","description":"main","kind":"panel","states":["default"],"size":{"width":100,"height":100}}]}`},
		{"wrong requested order", `{"components":[{"request_index":1,"name":"Inventory Panel","description":"main","kind":"panel","states":["default"],"size":{"width":100,"height":100}},{"request_index":0,"name":"Close Button","description":"close","kind":"button","states":["normal"],"size":{"width":50,"height":50}}]}`},
		{"inferred index", `{"components":[{"request_index":0,"name":"Inventory Panel","description":"main","kind":"panel","states":["default"],"size":{"width":100,"height":100}},{"request_index":1,"name":"Close Button","description":"close","kind":"button","states":["normal"],"size":{"width":50,"height":50}},{"request_index":2,"name":"Pause","description":"pause","kind":"button","states":["normal"],"size":{"width":20,"height":20}}]}`},
		{"bar with fill states", `{"components":[{"request_index":0,"name":"Inventory Panel","description":"main","kind":"bar","states":["full","empty"],"size":{"width":100,"height":20}},{"request_index":1,"name":"Close Button","description":"close","kind":"button","states":["normal"],"size":{"width":50,"height":50}}]}`},
		{"missing size", `{"components":[{"request_index":0,"name":"Inventory Panel","description":"main","kind":"panel","states":["default"]},{"request_index":1,"name":"Close Button","description":"close","kind":"button","states":["normal"],"size":{"width":50,"height":50}}]}`},
		{"outside canvas", `{"components":[{"request_index":0,"name":"Inventory Panel","description":"main","kind":"panel","states":["default"],"size":{"width":1025,"height":100}},{"request_index":1,"name":"Close Button","description":"close","kind":"button","states":["normal"],"size":{"width":50,"height":50}}]}`},
		{"state strip too wide", `{"components":[{"request_index":0,"name":"Inventory Panel","description":"main","kind":"panel","states":["a","b","c","d","e"],"size":{"width":1000,"height":100}},{"request_index":1,"name":"Close Button","description":"close","kind":"button","states":["normal"],"size":{"width":50,"height":50}}]}`},
		{"unknown field", `{"components":[{"request_index":0,"name":"Inventory Panel","description":"main","kind":"panel","states":["default"],"size":{"width":100,"height":100},"tiles":[]},{"request_index":1,"name":"Close Button","description":"close","kind":"button","states":["normal"],"size":{"width":50,"height":50}}]}`},
		{"trailing JSON", `{"components":[{"request_index":0,"name":"Inventory Panel","description":"main","kind":"panel","states":["default"],"size":{"width":100,"height":100}},{"request_index":1,"name":"Close Button","description":"close","kind":"button","states":["normal"],"size":{"width":50,"height":50}}]} {}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := decodeUISetComponentPlan([]byte(test.raw), definitions, canvas)
			if !errors.Is(err, ErrInvalidUISetPlan) {
				t.Fatalf("expected invalid UI Set plan, got %v", err)
			}
		})
	}
}

func TestPlanUISetComponentsPreservesProviderFailure(t *testing.T) {
	wantErr := errors.New("planning unavailable")
	executor := &executor{llm: &uiSetLLMStub{err: wantErr}}
	_, err := executor.planUISetComponents(context.Background(), validUISetPlanningPayload())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected provider error, got %v", err)
	}
}

func validUISetPlanningPayload() CreateUISetPayload {
	return CreateUISetPayload{
		AssetName: "Fantasy Inventory", ProjectID: 42, CreativeBrief: "compact fantasy inventory",
		Style: "ornate brass", Dimensions: assetdomain.Size{Width: 1024, Height: 768},
		Components: []UISetComponentDefinition{
			{Name: "Inventory Panel", Description: "main item grid container"},
			{Name: "Close Button", Description: "icon-only close control"},
		},
		ProjectContext: UISetProjectContext{
			Name: "Moon Forge", GameType: "RPG", TargetPlatform: "PC", Description: "inventory-driven adventure",
			Style: "moonlit pixel art", Reference: "projects/42/reference.png",
		},
	}
}
