package generator

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/1024XEngineer/Holonic-Asset/internal/module/generator/imageclient"
	"github.com/1024XEngineer/Holonic-Asset/internal/module/generator/llmclient"
	imageprocessor "github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image"
	assetdomain "github.com/1024XEngineer/Holonic-Asset/internal/module/workspace/asset"
	projectdomain "github.com/1024XEngineer/Holonic-Asset/internal/module/workspace/project"
)

func TestUISetGenerateAndEditWorkflowPreservesStateAndHistory(t *testing.T) {
	llm := &uiSetWorkflowLLM{responses: []json.RawMessage{
		json.RawMessage(`{"components":[
			{"request_index":0,"name":"Health Hearts","description":"player health","kind":"indicator","states":["full","damaged","empty"],"size":{"width":24,"height":24}},
			{"request_index":-1,"name":"Health Bar","description":"boss health frame","kind":"bar","states":["empty"],"size":{"width":160,"height":16}}
		]}`),
		json.RawMessage(`{"components":[
			{"index":0,"position":{"x":8,"y":8}},
			{"index":1,"position":{"x":8,"y":8}}
		]}`),
	}}
	images := &uiSetWorkflowImages{}
	resources := newUISetWorkflowResources()
	assets := &uiSetWorkflowAssets{}
	projects := &uiSetWorkflowProjects{project: &projectdomain.Project{
		ID: 42, Name: "Moon Forge", GameType: "RPG", TargetPlatform: projectdomain.PlatformTypePC,
		Description: "a tactical moonlit dungeon adventure", Style: "limited-palette moonlit pixel art",
		Reference: "projects/42/reference.png",
	}}
	executor := &executor{
		images: images, llm: llm, processor: imageprocessor.NewProcessor(), assets: assets,
		projects: projects, references: resources, resources: resources,
	}
	payload := validUISetWorkflowPayload()
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	result, err := executor.Generate(context.Background(), GenerateUISet, encodedPayload)
	if err != nil {
		t.Fatalf("generate UI Set: %v", err)
	}
	assertUISetResult(t, result, 100, 1)
	if assets.asset.Type != assetdomain.AssetTypeUISet || assets.asset.Version != 1 || len(assets.content.Components) != 2 {
		t.Fatalf("unexpected created UI Set: asset=%+v content=%+v", assets.asset, assets.content)
	}
	if len(images.requests) != 2 || len(llm.requests) != 2 || len(llm.requests[1].Images) != 2 {
		t.Fatalf("unexpected workflow calls: images=%d llm=%d layout_images=%d", len(images.requests), len(llm.requests), len(llm.requests[1].Images))
	}
	for _, required := range []string{"limited-palette moonlit pixel art", "RPG", "tactical moonlit dungeon adventure", "full", "damaged", "empty"} {
		if !strings.Contains(images.requests[0].Prompt, required) {
			t.Fatalf("Component prompt omitted %q: %s", required, images.requests[0].Prompt)
		}
	}
	heartsBefore := assets.content.Components[0]
	barBefore := assets.content.Components[1]
	if heartsBefore.Position != barBefore.Position {
		t.Fatalf("intentional overlap was not preserved: hearts=%+v bar=%+v", heartsBefore.Position, barBefore.Position)
	}
	heartState := decodeUISetStateForTest(t, heartsBefore.State)
	barState := decodeUISetStateForTest(t, barBefore.State)
	if len(heartState.Frames) != 3 || heartState.Frames[1].Name != "damaged" || heartState.TextureSize.Width != 72 {
		t.Fatalf("unexpected heart states: %+v", heartState)
	}
	if barState.RuntimeFill == nil || !barState.RuntimeFill.Enabled || len(barState.Frames) != 1 || barState.Frames[0].Name != "empty" {
		t.Fatalf("bar did not persist the empty runtime-fill contract: %+v", barState)
	}
	heartURLBefore := decodeUISetURLForTest(t, heartsBefore.Texture)
	barURLBefore := decodeUISetURLForTest(t, barBefore.Texture)
	if !resources.has(heartURLBefore) || !resources.has(barURLBefore) {
		t.Fatalf("created Component resources are missing: %v", resources.keys())
	}

	editPayload, err := json.Marshal(EditUISetComponentsPayload{
		AssetID: 100, ProjectID: 42, CreativeBrief: "make the heart highlights brighter",
		TargetAssetPaths: []string{"components.0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err = executor.Generate(context.Background(), EditUISetComponents, editPayload)
	if err != nil {
		t.Fatalf("edit UI Set Component: %v", err)
	}
	assertUISetResult(t, result, 100, 2)
	if assets.asset.Version != 2 || len(assets.records) != 1 {
		t.Fatalf("edit did not create one revision: asset=%+v records=%+v", assets.asset, assets.records)
	}
	heartsAfter := assets.content.Components[0]
	barAfter := assets.content.Components[1]
	heartURLAfter := decodeUISetURLForTest(t, heartsAfter.Texture)
	if heartURLAfter == heartURLBefore || decodeUISetURLForTest(t, barAfter.Texture) != barURLBefore {
		t.Fatalf("edit changed the wrong resources: before=%q after=%q bar=%q", heartURLBefore, heartURLAfter, decodeUISetURLForTest(t, barAfter.Texture))
	}
	if heartsAfter.ID != heartsBefore.ID || heartsAfter.Size != heartsBefore.Size || heartsAfter.Position != heartsBefore.Position ||
		string(heartsAfter.State) != string(heartsBefore.State) || barAfter.ID != barBefore.ID || string(barAfter.Texture) != string(barBefore.Texture) {
		t.Fatalf("edit changed preserved Component data: before=%+v after=%+v", heartsBefore, heartsAfter)
	}
	if !resources.has(heartURLBefore) || !resources.has(heartURLAfter) {
		t.Fatalf("historical or current Component resource was deleted: %v", resources.keys())
	}
}

func TestUISetEditFailureIsAtomicAndCleansNewResources(t *testing.T) {
	executor, assets, resources, images := generatedUISetWorkflow(t)
	beforeKeys := resources.keys()
	beforeContent := append(json.RawMessage(nil), assets.asset.Content...)
	assets.recordErr = errors.New("revision unavailable")
	payload, err := json.Marshal(EditUISetComponentsPayload{
		AssetID: 100, ProjectID: 42, CreativeBrief: "brighter",
		TargetAssetPaths: []string{"components.0", "components.1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = executor.Generate(context.Background(), EditUISetComponents, payload)
	if err == nil || !strings.Contains(err.Error(), "revision unavailable") {
		t.Fatalf("expected revision failure, got %v", err)
	}
	if string(assets.asset.Content) != string(beforeContent) || assets.asset.Version != 1 || len(assets.records) != 0 {
		t.Fatalf("failed edit mutated persisted UI Set: asset=%+v records=%+v", assets.asset, assets.records)
	}
	if fmt.Sprint(resources.keys()) != fmt.Sprint(beforeKeys) {
		t.Fatalf("failed edit leaked or deleted resources: before=%v after=%v", beforeKeys, resources.keys())
	}

	assets.recordErr = nil
	images.failContains = "Health Bar"
	_, err = executor.Generate(context.Background(), EditUISetComponents, payload)
	if err == nil || !strings.Contains(err.Error(), "Health Bar") {
		t.Fatalf("expected one Component generation failure, got %v", err)
	}
	if len(assets.records) != 0 || fmt.Sprint(resources.keys()) != fmt.Sprint(beforeKeys) {
		t.Fatalf("multi-Component generation failure was not atomic: records=%v resources=%v", assets.records, resources.keys())
	}
}

func TestDecodeUISetLayoutAllowsOverlapAndRejectsInvalidBounds(t *testing.T) {
	components := []processedUISetComponent{
		{Plan: UISetComponentPlan{Index: 0, Size: assetdomain.Size{Width: 24, Height: 24}}},
		{Plan: UISetComponentPlan{Index: 1, Size: assetdomain.Size{Width: 160, Height: 16}}},
	}
	positions, err := decodeUISetLayout([]byte(`{"components":[{"index":0,"position":{"x":8,"y":8}},{"index":1,"position":{"x":8,"y":8}}]}`), components, assetdomain.Size{Width: 320, Height: 180})
	if err != nil || positions[0] != positions[1] {
		t.Fatalf("intentional overlap should be accepted: positions=%v err=%v", positions, err)
	}
	for _, raw := range []string{
		`{}`,
		`{"components":[{"index":0,"position":{"x":8,"y":8}}]}`,
		`{"components":[{"index":0,"position":{"x":8,"y":8}},{"index":0,"position":{"x":8,"y":8}}]}`,
		`{"components":[{"index":0,"position":{"x":-1,"y":8}},{"index":1,"position":{"x":8,"y":8}}]}`,
		`{"components":[{"index":0,"position":{"x":8,"y":8}},{"index":1,"position":{"x":200,"y":8}}]}`,
	} {
		if _, err := decodeUISetLayout([]byte(raw), components, assetdomain.Size{Width: 320, Height: 180}); !errors.Is(err, ErrInvalidUISetLayout) {
			t.Fatalf("expected invalid layout for %s, got %v", raw, err)
		}
	}
}

func TestNormalizeUISetSingleStateBarCropsTransparentPadding(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 640, 640))
	frame := color.RGBA{R: 180, G: 40, B: 70, A: 255}
	for x := 120; x < 520; x++ {
		source.SetRGBA(x, 300, frame)
		source.SetRGBA(x, 339, frame)
	}
	for y := 300; y < 340; y++ {
		source.SetRGBA(120, y, frame)
		source.SetRGBA(519, y, frame)
	}
	encoded, err := imageprocessor.EncodePNGBase64(source)
	if err != nil {
		t.Fatal(err)
	}
	plan := UISetComponentPlan{
		Index: 0, Name: "Boss Health Bar", Kind: "bar", States: []string{"empty"},
		Size: assetdomain.Size{Width: 160, Height: 16},
	}
	normalized, mediaType, err := (&executor{processor: imageprocessor.NewProcessor()}).normalizeUISetStateStrip(
		context.Background(), encoded, plan, imageprocessor.RasterModePixel,
	)
	if err != nil {
		t.Fatalf("normalize single-state Bar: %v", err)
	}
	if mediaType != "image/png" {
		t.Fatalf("unexpected media type %q", mediaType)
	}
	data, err := decodeUISetPNG(normalized)
	if err != nil {
		t.Fatal(err)
	}
	value, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	bounds := visibleBoundsForTest(value)
	if bounds != image.Rect(0, 0, 160, 16) {
		t.Fatalf("single-state Bar did not fill its target frame: visible=%v image=%v", bounds, value.Bounds())
	}
	if _, _, _, alpha := value.At(80, 8).RGBA(); alpha != 0 {
		t.Fatalf("Bar interior must remain transparent, got alpha=%d", alpha)
	}
}

func TestUISetEditReferenceOrderPreservesSparseComponentIndexes(t *testing.T) {
	references := uiSetEditReferenceOrder(0, map[uint]string{
		0: "component-0.png",
		2: "component-2.png",
	}, []string{"shared.png"})
	want := []string{"component-0.png", "shared.png", "component-2.png"}
	if !slices.Equal(references, want) {
		t.Fatalf("unexpected sparse reference order: got=%v want=%v", references, want)
	}
}

func visibleBoundsForTest(value image.Image) image.Rectangle {
	bounds := value.Bounds()
	visible := image.Rectangle{}
	found := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := value.At(x, y).RGBA()
			if alpha>>8 <= uint32(imageprocessor.TransparentAlphaMax) {
				continue
			}
			pixel := image.Rect(x, y, x+1, y+1)
			if !found {
				visible = pixel
				found = true
			} else {
				visible = visible.Union(pixel)
			}
		}
	}
	return visible
}

func generatedUISetWorkflow(t *testing.T) (*executor, *uiSetWorkflowAssets, *uiSetWorkflowResources, *uiSetWorkflowImages) {
	t.Helper()
	llm := &uiSetWorkflowLLM{responses: []json.RawMessage{
		json.RawMessage(`{"components":[{"request_index":0,"name":"Health Hearts","description":"player health","kind":"indicator","states":["full","damaged","empty"],"size":{"width":24,"height":24}},{"request_index":-1,"name":"Health Bar","description":"boss health frame","kind":"bar","states":["empty"],"size":{"width":160,"height":16}}]}`),
		json.RawMessage(`{"components":[{"index":0,"position":{"x":8,"y":8}},{"index":1,"position":{"x":8,"y":40}}]}`),
	}}
	images := &uiSetWorkflowImages{}
	resources := newUISetWorkflowResources()
	assets := &uiSetWorkflowAssets{}
	projects := &uiSetWorkflowProjects{project: &projectdomain.Project{
		ID: 42, Name: "Moon Forge", GameType: "RPG", Description: "dungeon adventure", Style: "pixel art",
	}}
	executor := &executor{images: images, llm: llm, processor: imageprocessor.NewProcessor(), assets: assets, projects: projects, references: resources, resources: resources}
	payload, err := json.Marshal(validUISetWorkflowPayload())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := executor.Generate(context.Background(), GenerateUISet, payload); err != nil {
		t.Fatalf("prepare generated UI Set: %v", err)
	}
	return executor, assets, resources, images
}

func validUISetWorkflowPayload() CreateUISetPayload {
	return CreateUISetPayload{
		AssetName: "Moon HUD", ProjectID: 42, CreativeBrief: "compact combat HUD", Style: "silver ornament",
		Dimensions: assetdomain.Size{Width: 320, Height: 180},
		Components: []UISetComponentDefinition{{Name: "Health Hearts", Description: "player health"}},
		ProjectContext: UISetProjectContext{
			Name: "Moon Forge", GameType: "RPG", TargetPlatform: "PC", Description: "a tactical moonlit dungeon adventure",
			Style: "limited-palette moonlit pixel art", Reference: "projects/42/reference.png",
		},
	}
}

type uiSetWorkflowLLM struct {
	mu        sync.Mutex
	responses []json.RawMessage
	requests  []*llmclient.CompletionRequest
}

func (s *uiSetWorkflowLLM) Complete(_ context.Context, request *llmclient.CompletionRequest) (*llmclient.CompletionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copyRequest := *request
	copyRequest.Images = append([]llmclient.ImageInput(nil), request.Images...)
	s.requests = append(s.requests, &copyRequest)
	index := len(s.requests) - 1
	if index >= len(s.responses) {
		return nil, fmt.Errorf("unexpected LLM call %d", index+1)
	}
	return &llmclient.CompletionResult{JSON: append(json.RawMessage(nil), s.responses[index]...)}, nil
}

type uiSetWorkflowImages struct {
	mu           sync.Mutex
	requests     []*imageclient.GenerateRequest
	failContains string
}

func (s *uiSetWorkflowImages) Generate(_ context.Context, request *imageclient.GenerateRequest) (*imageclient.GenerateResult, error) {
	s.mu.Lock()
	copyRequest := *request
	copyRequest.ReferenceImages = append([]string(nil), request.ReferenceImages...)
	s.requests = append(s.requests, &copyRequest)
	fail := s.failContains != "" && strings.Contains(request.Prompt, s.failContains)
	s.mu.Unlock()
	if fail {
		return nil, fmt.Errorf("generation failed for %s", s.failContains)
	}
	var width, height int
	if _, err := fmt.Sscanf(request.Size, "%dx%d", &width, &height); err != nil || width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid requested size %q", request.Size)
	}
	value := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			value.SetRGBA(x, y, color.RGBA{G: 255, A: 255})
		}
	}
	for y := 2; y < height-2; y++ {
		for x := 2; x < width-2; x++ {
			if strings.Contains(request.Prompt, `kind "bar"`) && x > 4 && x < width-5 && y > 4 && y < height-5 {
				continue
			}
			value.SetRGBA(x, y, color.RGBA{R: 180, G: 40, B: 70, A: 255})
		}
	}
	encoded, err := imageprocessor.EncodePNGBase64(value)
	if err != nil {
		return nil, err
	}
	return &imageclient.GenerateResult{Images: []imageclient.GeneratedImage{{Base64: encoded, MediaType: "image/png"}}}, nil
}

type uiSetWorkflowResources struct {
	mu      sync.Mutex
	objects map[string][]byte
}

func newUISetWorkflowResources() *uiSetWorkflowResources {
	return &uiSetWorkflowResources{objects: make(map[string][]byte)}
}

func (s *uiSetWorkflowResources) ResolveReference(_ context.Context, reference string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if data, ok := s.objects[reference]; ok {
		return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data), nil
	}
	return "https://cdn.example/" + strings.TrimPrefix(reference, "/"), nil
}

func (*uiSetWorkflowResources) PersistReference(_ context.Context, reference string) (string, error) {
	return reference, nil
}

func (*uiSetWorkflowResources) NewObjectKey(string) (string, error) { return "unused", nil }

func (*uiSetWorkflowResources) PersistReferenceAt(context.Context, string, string) error { return nil }

func (s *uiSetWorkflowResources) DeleteObjects(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := s.DeleteObject(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (s *uiSetWorkflowResources) PutObject(_ context.Context, key, mediaType string, data []byte) error {
	if mediaType != "image/png" || len(data) == 0 {
		return fmt.Errorf("invalid PNG upload")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.objects[key]; exists {
		return fmt.Errorf("object %q already exists", key)
	}
	s.objects[key] = append([]byte(nil), data...)
	return nil
}

func (s *uiSetWorkflowResources) DeleteObject(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, key)
	return nil
}

func (s *uiSetWorkflowResources) has(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.objects[key]
	return ok
}

func (s *uiSetWorkflowResources) keys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := make([]string, 0, len(s.objects))
	for key := range s.objects {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

type uiSetWorkflowAssets struct {
	asset     assetdomain.Asset
	content   assetdomain.AssetContent
	records   []assetdomain.AssetRecord
	recordErr error
}

func (*uiSetWorkflowAssets) CreateCharacterAsset(context.Context, *assetdomain.Asset) (*assetdomain.Asset, error) {
	return nil, fmt.Errorf("unexpected character creation")
}

func (*uiSetWorkflowAssets) CreateObjectAsset(context.Context, *assetdomain.Asset) (uint, error) {
	return 0, fmt.Errorf("unexpected object creation")
}

func (*uiSetWorkflowAssets) CreateSceneryAsset(context.Context, *assetdomain.Asset) (uint, error) {
	return 0, fmt.Errorf("unexpected scenery creation")
}

func (*uiSetWorkflowAssets) CreateTileSetAsset(context.Context, *assetdomain.Asset) (uint, error) {
	return 0, fmt.Errorf("unexpected Tileset creation")
}

func (*uiSetWorkflowAssets) CreateAnimation(context.Context, uint, assetdomain.Animation) (uint, error) {
	return 0, fmt.Errorf("unexpected animation creation")
}

func (*uiSetWorkflowAssets) UpdateAnimationFrames(context.Context, uint, uint, []assetdomain.Frame) error {
	return fmt.Errorf("unexpected animation update")
}

func (s *uiSetWorkflowAssets) GetDetail(_ context.Context, id uint) (assetdomain.Asset, error) {
	if s.asset.ID != id {
		return assetdomain.Asset{}, nil
	}
	return s.asset, nil
}

func (s *uiSetWorkflowAssets) CreateUISetAsset(_ context.Context, value *assetdomain.Asset) (uint, error) {
	s.asset = *value
	s.asset.ID = 100
	s.asset.Version = 1
	content, err := s.asset.DecodeContent()
	if err != nil {
		return 0, err
	}
	s.content = content
	return s.asset.ID, nil
}

func (s *uiSetWorkflowAssets) CreateRecord(_ context.Context, record *assetdomain.AssetRecord, expectedVersion uint) (*assetdomain.AssetRecord, error) {
	if s.recordErr != nil {
		return nil, s.recordErr
	}
	if expectedVersion != s.asset.Version {
		return nil, fmt.Errorf("version conflict")
	}
	content, err := (assetdomain.Asset{Type: assetdomain.AssetTypeUISet, Content: record.Content}).DecodeContent()
	if err != nil {
		return nil, err
	}
	s.content = content
	s.asset.Content = append(json.RawMessage(nil), record.Content...)
	s.asset.Version++
	created := assetdomain.AssetRecord{AssetID: s.asset.ID, Version: s.asset.Version, Content: append(json.RawMessage(nil), record.Content...)}
	s.records = append(s.records, created)
	return &created, nil
}

type uiSetWorkflowProjects struct{ project *projectdomain.Project }

func (s *uiSetWorkflowProjects) GetDetail(context.Context, uint) (*projectdomain.Project, error) {
	return s.project, nil
}

func assertUISetResult(t *testing.T, raw json.RawMessage, assetID, version uint) {
	t.Helper()
	var result ExecutionResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("decode UI Set result: %v", err)
	}
	if result.AssetID != assetID || result.Version != version {
		t.Fatalf("unexpected UI Set result: %+v", result)
	}
}

func decodeUISetStateForTest(t *testing.T, raw json.RawMessage) assetdomain.UIComponentState {
	t.Helper()
	var state assetdomain.UIComponentState
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatal(err)
	}
	return state
}

func decodeUISetURLForTest(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	url, err := decodeUISetTextureURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	return url
}

var _ AssetWriter = (*uiSetWorkflowAssets)(nil)
var _ ReferenceStore = (*uiSetWorkflowResources)(nil)
var _ ResourceStore = (*uiSetWorkflowResources)(nil)
