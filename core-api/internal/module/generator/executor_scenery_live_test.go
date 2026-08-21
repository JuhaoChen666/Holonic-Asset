package generator

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/1024XEngineer/Holonic-Asset/internal/config"
	"github.com/1024XEngineer/Holonic-Asset/internal/module/generator/imageclient"
	"github.com/1024XEngineer/Holonic-Asset/internal/module/generator/llmclient"
	imageprocessor "github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image"
	assetdomain "github.com/1024XEngineer/Holonic-Asset/internal/module/workspace/asset"
)

const sceneryLiveEnv = "HOLONIC_LLM_INTEGRATION"

func TestLiveSceneryPlanningAndLayoutWithRealLLM(t *testing.T) {
	if strings.TrimSpace(os.Getenv(sceneryLiveEnv)) != "1" {
		t.Skip("set HOLONIC_LLM_INTEGRATION=1 to run real scenery LLM regression test")
	}

	llmConfig, err := loadLiveLLMConfig()
	if err != nil {
		t.Fatalf("load live LLM config: %v", err)
	}

	provider := llmclient.NewQNAProvider(llmclient.QNAConfig{
		BaseURL:      llmConfig.BaseURL,
		APIKey:       llmConfig.APIKey,
		DefaultModel: llmConfig.DefaultModel,
	})
	llmService := llmclient.NewLLMService(provider)

	exec := &executor{
		llm:       llmService,
		processor: imageprocessor.NewProcessor(),
	}

	testCases := []struct {
		name        string
		assetName   string
		brief       string
		perspective string
		width       int
		height      int
	}{
		{
			name:        "Forest",
			assetName:   "森林",
			brief:       "A mystical ancient forest at twilight with tall pine silhouettes and misty atmosphere",
			perspective: "Side-On",
			width:       640,
			height:      360,
		},
		{
			name:        "Snow",
			assetName:   "冰天雪地",
			brief:       "A frozen winter landscape with snow-covered peaks, icy pine trees, and drifting snowflakes",
			perspective: "Side-On",
			width:       640,
			height:      360,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			payload := CreateSceneryPayload{
				ProjectID:     22,
				AssetName:     tc.assetName,
				CreativeBrief: tc.brief,
				Perspective:   tc.perspective,
				Dimensions:    assetdomain.Size{Width: uint(tc.width), Height: uint(tc.height)},
				ProjectContext: SceneryProjectContext{
					Name:           "Adventure Game",
					GameType:       "Platformer",
					TargetPlatform: "PC",
					Description:    "A 2D platformer adventure",
				},
			}

			// 1. Live layer planning
			t.Logf("[%s] Calling real QNA LLM planSceneryLayers for %q...", tc.name, tc.assetName)
			plan, err := exec.planSceneryLayers(ctx, payload, "")
			if err != nil {
				t.Fatalf("[%s] live planSceneryLayers failed: %v", tc.name, err)
			}
			if len(plan) == 0 {
				t.Fatalf("[%s] live planSceneryLayers returned 0 layers", tc.name)
			}
			t.Logf("[%s] live planSceneryLayers returned %d layers:", tc.name, len(plan))
			for i, layer := range plan {
				t.Logf("  Layer %d: Name=%q, CreativeBrief=%q", i+1, layer.Name, layer.CreativeBrief)
			}

			// 2. Prepare test processed layers
			processedLayers := make([]ProcessedSceneryLayer, len(plan))
			for i, p := range plan {
				r := uint8((30 * (i + 1)) & 0xFF)
				dataURI := liveTestPNGDataURI(t, tc.width, tc.height, color.RGBA{R: r, G: 150, B: 50, A: 255})
				data := strings.TrimPrefix(dataURI, "data:image/png;base64,")
				processedLayers[i] = ProcessedSceneryLayer{
					ID:          uint(i + 1),
					Name:        p.Name,
					MediaType:   "image/png",
					ImageBase64: data,
				}
			}

			// 3. Live layout analysis
			t.Logf("[%s] Calling real QNA LLM analyzeSceneryLayout...", tc.name)
			laidOut, _, _, err := exec.analyzeSceneryLayout(ctx, payload, processedLayers)
			if err != nil {
				t.Fatalf("[%s] live analyzeSceneryLayout failed: %v", tc.name, err)
			}
			if len(laidOut) != len(processedLayers) {
				t.Fatalf("[%s] laidOut count %d != processedLayers count %d", tc.name, len(laidOut), len(processedLayers))
			}
			t.Logf("[%s] live analyzeSceneryLayout returned %d laid out layers:", tc.name, len(laidOut))
			for _, l := range laidOut {
				t.Logf("  Layer %d (%s): pos=(%.1f, %.1f), scale=(%.2f, %.2f), rotation=%.1f, opacity=%.2f, zIndex=%d",
					l.ID, l.Name, l.Layout.Position.X, l.Layout.Position.Y, l.Layout.Scale.X, l.Layout.Scale.Y,
					l.Layout.Rotation, l.Layout.Opacity, l.Layout.ZIndex)
			}
		})
	}
}

func loadLiveLLMConfig() (config.LLMClientConfig, error) {
	reader := viper.New()
	for _, path := range []string{"../../config/config.yaml", "../config/config.yaml", "config.yaml", "../internal/config/config.yaml", "internal/config/config.yaml"} {
		reader.SetConfigFile(path)
		if err := reader.ReadInConfig(); err == nil {
			break
		}
	}
	section := reader.Sub("llm")
	if section == nil {
		return config.LLMClientConfig{}, nil
	}
	var value config.LLMClientConfig
	if err := section.UnmarshalExact(&value); err != nil {
		return config.LLMClientConfig{}, err
	}
	return value, nil
}

func loadLiveImageConfig() (config.ImageClientConfig, error) {
	reader := viper.New()
	for _, path := range []string{"../../config/config.yaml", "../config/config.yaml", "config.yaml", "../internal/config/config.yaml", "internal/config/config.yaml"} {
		reader.SetConfigFile(path)
		if err := reader.ReadInConfig(); err == nil {
			break
		}
	}
	section := reader.Sub("image")
	if section == nil {
		return config.ImageClientConfig{}, nil
	}
	var value config.ImageClientConfig
	if err := section.UnmarshalExact(&value); err != nil {
		return config.ImageClientConfig{}, err
	}
	if model := os.Getenv("HOLONIC_IMAGE_MODEL"); model != "" {
		value.DefaultModel = model
	}
	if fallback := os.Getenv("HOLONIC_IMAGE_FALLBACK_MODEL"); fallback != "" {
		value.FallbackModel = fallback
	} else if value.FallbackModel == "" && value.DefaultModel != "openai/gpt-image-2" {
		value.FallbackModel = "openai/gpt-image-2"
	}
	if provider := os.Getenv("HOLONIC_IMAGE_PROVIDER"); provider != "" {
		value.Provider = provider
	}
	return value, nil
}

func liveTestPNGDataURI(t *testing.T, w, h int, c color.RGBA) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetRGBA(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

func liveTestPixelArtPrototypeURI(t *testing.T) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 128, 72))
	for y := range 72 {
		for x := range 128 {
			var c color.RGBA
			if y < 20 {
				c = color.RGBA{R: 50, G: 20, B: 60, A: 255} // Dark purple sky
			} else if y < 38 {
				c = color.RGBA{R: 190, G: 80, B: 40, A: 255} // Dusk orange
			} else if y < 50 {
				c = color.RGBA{R: 245, G: 165, B: 60, A: 255} // Sunset glow
			} else if y < 62 {
				c = color.RGBA{R: 40, G: 65, B: 85, A: 255} // Lake water
			} else {
				c = color.RGBA{R: 25, G: 35, B: 20, A: 255} // Ground shore
			}
			img.SetRGBA(x, y, c)
		}
	}
	for x := range 128 {
		mHeight := 28 + int(8.0*float64(x%32)/32.0)
		for y := 72 - mHeight; y < 50; y++ {
			img.SetRGBA(x, y, color.RGBA{R: 40, G: 25, B: 45, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

type memoryResourceStore struct {
	objects map[string][]byte
}

func (m *memoryResourceStore) PutObject(_ context.Context, key, _ string, data []byte) error {
	if m.objects == nil {
		m.objects = make(map[string][]byte)
	}
	m.objects[key] = data
	return nil
}

func (m *memoryResourceStore) DeleteObject(_ context.Context, key string) error {
	delete(m.objects, key)
	return nil
}

type memoryAssetWriter struct {
	created *assetdomain.Asset
}

func (m *memoryAssetWriter) GetDetail(_ context.Context, _ uint) (assetdomain.Asset, error) {
	return assetdomain.Asset{}, nil
}
func (m *memoryAssetWriter) CreateCharacterAsset(_ context.Context, a *assetdomain.Asset) (*assetdomain.Asset, error) {
	return a, nil
}
func (m *memoryAssetWriter) CreateObjectAsset(_ context.Context, _ *assetdomain.Asset) (uint, error) {
	return 1, nil
}
func (m *memoryAssetWriter) CreateSceneryAsset(_ context.Context, a *assetdomain.Asset) (uint, error) {
	m.created = a
	return 100, nil
}
func (m *memoryAssetWriter) CreateTileSetAsset(_ context.Context, _ *assetdomain.Asset) (uint, error) {
	return 1, nil
}

type memoryReferenceStore struct {
	refs map[string]string
}

func (m *memoryReferenceStore) ResolveReference(_ context.Context, key string) (string, error) {
	if val, ok := m.refs[key]; ok {
		return val, nil
	}
	return key, nil
}
func (m *memoryReferenceStore) PersistReference(_ context.Context, key string) (string, error) {
	return key, nil
}
func (m *memoryReferenceStore) NewObjectKey(ext string) (string, error) {
	return "ref" + ext, nil
}
func (m *memoryReferenceStore) PersistReferenceAt(_ context.Context, _, _ string) error {
	return nil
}
func (m *memoryReferenceStore) DeleteObjects(_ context.Context, _ []string) error {
	return nil
}

func TestLiveFullSceneryEndToEndGeneration(t *testing.T) {
	if strings.TrimSpace(os.Getenv(sceneryLiveEnv)) != "1" {
		t.Skip("set HOLONIC_LLM_INTEGRATION=1 to run real scenery end-to-end generation")
	}

	llmCfg, err := loadLiveLLMConfig()
	if err != nil || llmCfg.APIKey == "" {
		t.Fatalf("load live LLM config: %v", err)
	}
	imgCfg, err := loadLiveImageConfig()
	if err != nil || imgCfg.APIKey == "" {
		t.Fatalf("load live Image config: %v", err)
	}

	llmProvider := llmclient.NewQNAProvider(llmclient.QNAConfig{
		BaseURL:      llmCfg.BaseURL,
		APIKey:       llmCfg.APIKey,
		DefaultModel: llmCfg.DefaultModel,
	})
	llmService := llmclient.NewLLMService(llmProvider)

	imgProvider := imageclient.NewImageProvider(imageclient.FactoryConfig{
		BaseURL:       imgCfg.BaseURL,
		APIKey:        imgCfg.APIKey,
		DefaultModel:  imgCfg.DefaultModel,
		FallbackModel: imgCfg.FallbackModel,
		Provider:      imgCfg.Provider,
	})
	imageService := imageclient.NewImageGenerationService(imgProvider)
	processor := imageprocessor.NewProcessor()

	resources := &memoryResourceStore{objects: make(map[string][]byte)}
	assets := &memoryAssetWriter{}

	// Sample pixel art prototype image data URL as project prototype
	prototypeURI := liveTestPixelArtPrototypeURI(t)
	references := &memoryReferenceStore{refs: map[string]string{"proto": prototypeURI}}

	exec := NewExecutorWithDependencies(
		imageService,
		processor,
		assets,
		ExecutorDependencies{
			LLM:        llmService,
			Resources:  resources,
			References: references,
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	payload := CreateSceneryPayload{
		ProjectID:     2,
		AssetName:     "MoonDrew Valley Farm Village",
		CreativeBrief: "A cozy MoonDrew Valley farming village at late afternoon. Show a small farmhouse beside tilled crop fields, wooden fences, a stone path, a calm river, distant hills, scattered trees, and warm golden sunlight. The scene should feel peaceful, welcoming, and suitable for a casual farming life-simulation game.",
		Perspective:   "Side-On",
		Reference:     "proto",
		Dimensions:    assetdomain.Size{Width: 640, Height: 360},
		ProjectContext: SceneryProjectContext{
			Name:           "MoonDrew Valley",
			GameType:       "casual farming life-simulation",
			TargetPlatform: "PC / Mobile",
			Description:    "A casual farming life-simulation game set in MoonDrew Valley.",
		},
	}

	t.Log("Starting live full scenery generation...")
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	resultRaw, err := exec.Generate(ctx, GenerateScenery, payloadBytes)
	if err != nil {
		t.Fatalf("execute scenery generation failed: %v", err)
	}
	t.Logf("Generation succeeded! Result: %s", string(resultRaw))

	if assets.created == nil {
		t.Fatal("expected asset to be created")
	}

	var content assetdomain.AssetContent
	if err := json.Unmarshal(assets.created.Content, &content); err != nil {
		t.Fatalf("unmarshal asset content: %v", err)
	}
	t.Logf("Created scenery asset with %d layers:", len(content.Layers))

	// Export individual layers and composite image
	composite := image.NewRGBA(image.Rect(0, 0, 640, 360))
	artifactDir := "/Users/lx/.gemini/antigravity/brain/03e53355-54e3-4379-9149-a8bb19de001d"
	for _, l := range content.Layers {
		data := resources.objects[l.Resource]
		t.Logf("Layer %d (%s): resource=%s, bytes=%d, pos=(%.1f, %.1f)", l.ID, l.Name, l.Resource, len(data), l.Position.X, l.Position.Y)
		layerPath := fmt.Sprintf("%s/layer_%d_%s.png", artifactDir, l.ID, l.Name)
		_ = os.WriteFile(layerPath, data, 0644)
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			t.Logf("decode layer %d error: %v", l.ID, err)
			continue
		}
		// Alpha-blend layer onto composite
		pt := image.Point{X: -int(l.Position.X), Y: -int(l.Position.Y)}
		draw.Draw(composite, composite.Bounds(), img, pt, draw.Over)
	}

	compositePath := artifactDir + "/generated_scenery_composite.png"
	out, err := os.Create(compositePath)
	if err != nil {
		t.Fatalf("create composite image file: %v", err)
	}
	defer out.Close()
	if err := png.Encode(out, composite); err != nil {
		t.Fatalf("encode composite PNG: %v", err)
	}
	t.Logf("Saved live regression test scenery composite image to: %s", compositePath)
}
