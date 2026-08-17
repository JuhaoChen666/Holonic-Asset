package generator

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/1024XEngineer/Holonic-Asset/internal/module/generator/imageclient"
	"github.com/1024XEngineer/Holonic-Asset/internal/module/generator/llmclient"
	"github.com/1024XEngineer/Holonic-Asset/internal/module/generator/prompts"
	imageprocessor "github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image"
	assetdomain "github.com/1024XEngineer/Holonic-Asset/internal/module/workspace/asset"
)

const (
	maxUISetComponentConcurrency = 4
	maxUISetGenerationAttempts   = 2
	maxUISetLLMAttempts          = 2
	uiSetBatchIDBytes            = 16
	uiSetCleanupTTL              = 15 * time.Second
	uiSetLayoutSchemaName        = "uiset_component_layout"
)

var uiSetLayoutJSONSchema = json.RawMessage(`{"type":"object","additionalProperties":false,"required":["components"],"properties":{"components":{"type":"array","minItems":1,"maxItems":64,"items":{"type":"object","additionalProperties":false,"required":["index","position"],"properties":{"index":{"type":"integer","minimum":0,"maximum":63},"position":{"type":"object","additionalProperties":false,"required":["x","y"],"properties":{"x":{"type":"number","minimum":0},"y":{"type":"number","minimum":0}}}}}}}}`)

type processedUISetComponent struct {
	Plan        UISetComponentPlan
	ImageBase64 string
	MediaType   string
	Position    assetdomain.Position
}

type uiSetLayoutResponse struct {
	Components *[]uiSetLayoutCandidate `json:"components"`
}

type uiSetLayoutCandidate struct {
	Index    *int                    `json:"index"`
	Position *uiSetPositionCandidate `json:"position"`
}

type uiSetPositionCandidate struct {
	X *float64 `json:"x"`
	Y *float64 `json:"y"`
}

func (e *executor) generateUISet(ctx context.Context, payload CreateUISetPayload) (json.RawMessage, error) {
	plans, err := e.planUISetComponents(ctx, payload)
	if err != nil {
		return nil, err
	}
	references, err := e.resolveUISetReferences(ctx, payload.Reference, payload.ProjectContext.Reference)
	if err != nil {
		return nil, err
	}
	components, err := e.generateUISetComponents(ctx, payload, plans, references)
	if err != nil {
		return nil, err
	}
	laidOut, err := e.analyzeUISetLayout(ctx, payload, components)
	if err != nil {
		return nil, err
	}
	return e.persistUISet(ctx, payload, laidOut)
}

func (e *executor) planUISetComponents(ctx context.Context, payload CreateUISetPayload) ([]UISetComponentPlan, error) {
	if e.llm == nil {
		return nil, ErrLLMServiceRequired
	}
	if err := validateCreateUISetPayload(&payload); err != nil {
		return nil, err
	}
	components := make([]prompts.UISetComponentInput, len(payload.Components))
	for index, component := range payload.Components {
		components[index] = prompts.UISetComponentInput{Index: index, Name: component.Name, Description: component.Description}
	}
	prompt := prompts.UISetPlan(prompts.UISetPlanInput{
		AssetName: payload.AssetName, CreativeBrief: payload.CreativeBrief, Style: payload.Style,
		ProjectStyle: payload.ProjectContext.Style, ProjectName: payload.ProjectContext.Name,
		GameType: payload.ProjectContext.GameType, TargetPlatform: payload.ProjectContext.TargetPlatform,
		ProjectDescription: payload.ProjectContext.Description, Width: payload.Dimensions.Width,
		Height: payload.Dimensions.Height, Components: components,
	})
	completion, err := completeUISetLLM(ctx, e.llm, &llmclient.CompletionRequest{
		Prompt:         prompt,
		ResponseSchema: llmclient.JSONSchema{Name: uiSetComponentPlanSchemaName, Schema: append(json.RawMessage(nil), uiSetComponentPlanJSONSchema...)},
	})
	if err != nil {
		return nil, fmt.Errorf("generator: plan UI Set components: %w", err)
	}
	if completion == nil {
		return nil, fmt.Errorf("%w: LLM returned no completion", ErrInvalidUISetPlan)
	}
	return decodeUISetComponentPlan(completion.JSON, payload.Components, payload.Dimensions)
}

func (e *executor) resolveUISetReferences(ctx context.Context, references ...string) ([]string, error) {
	resolved := make([]string, 0, len(references))
	seen := make(map[string]struct{}, len(references))
	for _, reference := range references {
		reference = strings.TrimSpace(reference)
		if reference == "" {
			continue
		}
		value, err := e.references.ResolveReference(ctx, reference)
		if err != nil {
			return nil, fmt.Errorf("generator: resolve UI Set reference %q: %w", reference, err)
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, fmt.Errorf("generator: resolve UI Set reference %q: empty result", reference)
		}
		if _, duplicate := seen[value]; duplicate {
			continue
		}
		seen[value] = struct{}{}
		resolved = append(resolved, value)
	}
	return resolved, nil
}

func (e *executor) generateUISetComponents(
	ctx context.Context,
	payload CreateUISetPayload,
	plans []UISetComponentPlan,
	references []string,
) ([]processedUISetComponent, error) {
	components := make([]processedUISetComponent, len(plans))
	err := runBoundedUISetJobs(ctx, len(plans), maxUISetComponentConcurrency, func(jobCtx context.Context, index int) error {
		component, generateErr := e.generateUISetComponent(jobCtx, payload, plans[index], references)
		if generateErr != nil {
			return generateErr
		}
		components[index] = component
		return nil
	})
	if err != nil {
		return nil, err
	}
	return components, nil
}

func (e *executor) generateUISetComponent(
	ctx context.Context,
	payload CreateUISetPayload,
	plan UISetComponentPlan,
	references []string,
) (processedUISetComponent, error) {
	if len(plan.States) == 0 {
		return processedUISetComponent{}, fmt.Errorf("%w: Component %d has no states", ErrInvalidUISetPlan, plan.Index)
	}
	textureWidth := plan.Size.Width * uint(len(plan.States))
	prompt := prompts.UISetComponent(prompts.UISetComponentGenerationInput{
		AssetName: payload.AssetName, CreativeBrief: payload.CreativeBrief, Style: payload.Style,
		ProjectStyle: payload.ProjectContext.Style, ProjectName: payload.ProjectContext.Name,
		GameType: payload.ProjectContext.GameType, TargetPlatform: payload.ProjectContext.TargetPlatform,
		ProjectDescription: payload.ProjectContext.Description, ComponentName: plan.Name,
		ComponentBrief: plan.Description, Kind: plan.Kind, States: plan.States,
		FrameWidth: plan.Size.Width, FrameHeight: plan.Size.Height, HasReference: len(references) > 0,
	})
	var lastQualityErr error
	for attempt := 1; attempt <= maxUISetGenerationAttempts; attempt++ {
		attemptPrompt := prompt
		if attempt > 1 {
			attemptPrompt += fmt.Sprintf(`

CORRECTION REQUIRED: The previous image failed the state-strip contract: %s
Regenerate the complete strip from scratch. Every named state cell must contain clearly visible non-green Component pixels, with no blank, omitted, merged, or duplicated frame.`, lastQualityErr)
		}
		generated, err := e.images.Generate(ctx, &imageclient.GenerateRequest{
			Prompt: attemptPrompt, ReferenceImages: append([]string(nil), references...),
			Size: fmt.Sprintf("%dx%d", textureWidth, plan.Size.Height),
		})
		if err != nil {
			return processedUISetComponent{}, fmt.Errorf("generator: generate UI Component %d %q: %w", plan.Index, plan.Name, err)
		}
		if generated == nil || len(generated.Images) != 1 || strings.TrimSpace(generated.Images[0].Base64) == "" {
			lastQualityErr = fmt.Errorf("expected exactly one image")
			continue
		}
		removed, err := e.processor.RemoveBackground(ctx, &imageprocessor.RemoveBackgroundRequest{
			ImageBase64: generated.Images[0].Base64, MatteColor: imageprocessor.DefaultMatteColor,
		})
		if err != nil {
			return processedUISetComponent{}, fmt.Errorf("generator: remove UI Component %d background: %w", plan.Index, err)
		}
		if removed == nil || strings.TrimSpace(removed.ImageBase64) == "" {
			return processedUISetComponent{}, fmt.Errorf("generator: remove UI Component %d background: empty result", plan.Index)
		}
		normalizedBase64, normalizedMediaType, err := e.normalizeUISetStateStrip(
			ctx,
			removed.ImageBase64,
			plan,
			uiSetRasterMode(payload.ProjectContext.Style, payload.Style),
		)
		if err != nil {
			lastQualityErr = err
			continue
		}
		verified, err := e.processor.Verify(ctx, &imageprocessor.VerifyRequest{
			ImageBase64: normalizedBase64, Profile: imageprocessor.ProfileGeneric,
			ExpectedMatteColor: imageprocessor.DefaultMatteColor,
		})
		if err != nil {
			return processedUISetComponent{}, fmt.Errorf("generator: verify UI Component %d: %w", plan.Index, err)
		}
		if verified == nil || !verified.Passed {
			lastQualityErr = fmt.Errorf("transparent PNG verification failed")
			continue
		}
		if err := validateUISetSpriteStrip(normalizedBase64, plan); err != nil {
			lastQualityErr = err
			continue
		}
		return processedUISetComponent{Plan: plan, ImageBase64: normalizedBase64, MediaType: normalizedMediaType}, nil
	}
	return processedUISetComponent{}, fmt.Errorf("generator: generate UI Component %d %q after %d attempts: %w",
		plan.Index, plan.Name, maxUISetGenerationAttempts, lastQualityErr)
}

func (e *executor) normalizeUISetStateStrip(
	ctx context.Context,
	imageBase64 string,
	plan UISetComponentPlan,
	mode imageprocessor.RasterMode,
) (string, string, error) {
	if len(plan.States) == 1 {
		resized, err := e.processor.Resize(ctx, &imageprocessor.ResizeRequest{
			ImageBase64: imageBase64,
			Options: imageprocessor.ResizeOptions{
				Width: int(plan.Size.Width), Height: int(plan.Size.Height), Margin: 0, CropContent: true, Mode: mode,
			},
		})
		if err != nil {
			return "", "", fmt.Errorf("resize single-state UI Component %d: %w", plan.Index, err)
		}
		if resized == nil || strings.TrimSpace(resized.ImageBase64) == "" || resized.MIMEType != "image/png" {
			return "", "", fmt.Errorf("resize single-state UI Component %d: invalid PNG result", plan.Index)
		}
		return resized.ImageBase64, resized.MIMEType, nil
	}
	normalized, err := e.processor.SplitImage(ctx, &imageprocessor.SplitImageRequest{
		ImageBase64: imageBase64, Mode: imageprocessor.ImageSplitModeAnimation,
		Columns: len(plan.States), Rows: 1, FrameCount: len(plan.States),
		FrameWidth: int(plan.Size.Width), FrameHeight: int(plan.Size.Height),
		DetectGridBounds: true, Anchor: imageprocessor.AnimationAnchorCenter,
	})
	if err != nil {
		return "", "", fmt.Errorf("normalize UI Component %d state strip: %w", plan.Index, err)
	}
	if normalized == nil || len(normalized.Regions) != len(plan.States) ||
		strings.TrimSpace(normalized.ImageBase64) == "" || normalized.MIMEType != "image/png" {
		return "", "", fmt.Errorf("normalize UI Component %d state strip: expected %d nonempty frames", plan.Index, len(plan.States))
	}
	return normalized.ImageBase64, normalized.MIMEType, nil
}

func uiSetRasterMode(styles ...string) imageprocessor.RasterMode {
	for _, style := range styles {
		if strings.Contains(strings.ToLower(style), "pixel") {
			return imageprocessor.RasterModePixel
		}
	}
	return imageprocessor.RasterModeSmooth
}

func validateUISetSpriteStrip(encoded string, plan UISetComponentPlan) error {
	data, err := decodeUISetPNG(encoded)
	if err != nil {
		return fmt.Errorf("generator: validate UI Component %d PNG: %w", plan.Index, err)
	}
	imageValue, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("generator: validate UI Component %d PNG: %w", plan.Index, err)
	}
	wantWidth := int(plan.Size.Width) * len(plan.States)
	if imageValue.Bounds().Dx() != wantWidth || imageValue.Bounds().Dy() != int(plan.Size.Height) {
		return fmt.Errorf("generator: validate UI Component %d PNG: got %dx%d, want %dx%d", plan.Index,
			imageValue.Bounds().Dx(), imageValue.Bounds().Dy(), wantWidth, plan.Size.Height)
	}
	for stateIndex := range plan.States {
		visible := false
		minX := imageValue.Bounds().Min.X + stateIndex*int(plan.Size.Width)
		maxX := minX + int(plan.Size.Width)
		for y := imageValue.Bounds().Min.Y; y < imageValue.Bounds().Max.Y && !visible; y++ {
			for x := minX; x < maxX; x++ {
				_, _, _, alpha := imageValue.At(x, y).RGBA()
				if alpha>>8 > uint32(imageprocessor.TransparentAlphaMax) {
					visible = true
					break
				}
			}
		}
		if !visible {
			return fmt.Errorf("generator: validate UI Component %d PNG: state %q is empty", plan.Index, plan.States[stateIndex])
		}
	}
	return nil
}

func decodeUISetPNG(encoded string) ([]byte, error) {
	encoded = strings.TrimSpace(encoded)
	if strings.HasPrefix(strings.ToLower(encoded), "data:") {
		_, value, found := strings.Cut(encoded, ",")
		if !found {
			return nil, fmt.Errorf("invalid data URL")
		}
		encoded = value
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(data) == 0 {
		if err == nil {
			err = fmt.Errorf("empty PNG")
		}
		return nil, err
	}
	return data, nil
}

func (e *executor) analyzeUISetLayout(
	ctx context.Context,
	payload CreateUISetPayload,
	components []processedUISetComponent,
) ([]processedUISetComponent, error) {
	if len(components) == 0 {
		return nil, fmt.Errorf("%w: at least one processed Component is required", ErrInvalidUISetLayout)
	}
	promptComponents := make([]prompts.UISetLayoutComponentInput, len(components))
	images := make([]llmclient.ImageInput, len(components))
	for index, component := range components {
		promptComponents[index] = prompts.UISetLayoutComponentInput{
			Index: component.Plan.Index, Name: component.Plan.Name, Kind: component.Plan.Kind,
			Width: component.Plan.Size.Width, Height: component.Plan.Size.Height,
		}
		images[index] = llmclient.ImageInput{URL: "data:" + component.MediaType + ";base64," + component.ImageBase64}
	}
	prompt := prompts.UISetLayout(prompts.UISetLayoutInput{
		AssetName: payload.AssetName, CreativeBrief: payload.CreativeBrief, Style: payload.Style,
		ProjectStyle: payload.ProjectContext.Style, GameType: payload.ProjectContext.GameType,
		ProjectDescription: payload.ProjectContext.Description, Width: payload.Dimensions.Width,
		Height: payload.Dimensions.Height, Components: promptComponents,
	})
	completion, err := completeUISetLLM(ctx, e.llm, &llmclient.CompletionRequest{
		Prompt: prompt, Images: images,
		ResponseSchema: llmclient.JSONSchema{Name: uiSetLayoutSchemaName, Schema: append(json.RawMessage(nil), uiSetLayoutJSONSchema...)},
	})
	if err != nil {
		return nil, fmt.Errorf("generator: analyze UI Set layout: %w", err)
	}
	if completion == nil {
		return nil, fmt.Errorf("%w: LLM returned no completion", ErrInvalidUISetLayout)
	}
	positions, err := decodeUISetLayout(completion.JSON, components, payload.Dimensions)
	if err != nil {
		return nil, err
	}
	result := make([]processedUISetComponent, len(components))
	for index, component := range components {
		component.Position = positions[component.Plan.Index]
		result[index] = component
	}
	return result, nil
}

func completeUISetLLM(
	ctx context.Context,
	service llmclient.LLMService,
	request *llmclient.CompletionRequest,
) (*llmclient.CompletionResult, error) {
	var lastErr error
	for attempt := 1; attempt <= maxUISetLLMAttempts; attempt++ {
		completion, err := service.Complete(ctx, request)
		if err == nil {
			return completion, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	return nil, lastErr
}

func decodeUISetLayout(
	raw []byte,
	components []processedUISetComponent,
	canvas assetdomain.Size,
) (map[uint]assetdomain.Position, error) {
	invalid := func(reason string) error { return fmt.Errorf("%w: %s", ErrInvalidUISetLayout, reason) }
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var response uiSetLayoutResponse
	if err := decoder.Decode(&response); err != nil {
		return nil, invalid(err.Error())
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, invalid("trailing data")
		}
		return nil, invalid(err.Error())
	}
	if response.Components == nil || len(*response.Components) != len(components) {
		return nil, invalid(fmt.Sprintf("expected %d Component positions", len(components)))
	}
	byIndex := make(map[uint]processedUISetComponent, len(components))
	for _, component := range components {
		byIndex[component.Plan.Index] = component
	}
	positions := make(map[uint]assetdomain.Position, len(components))
	for candidateIndex, candidate := range *response.Components {
		if candidate.Index == nil || *candidate.Index < 0 {
			return nil, invalid(fmt.Sprintf("position %d index is required", candidateIndex))
		}
		index := uint(*candidate.Index)
		component, known := byIndex[index]
		if !known {
			return nil, invalid(fmt.Sprintf("position %d has unknown Component index %d", candidateIndex, index))
		}
		if _, duplicate := positions[index]; duplicate {
			return nil, invalid(fmt.Sprintf("Component index %d is duplicated", index))
		}
		if candidate.Position == nil || candidate.Position.X == nil || candidate.Position.Y == nil {
			return nil, invalid(fmt.Sprintf("Component index %d position must contain x and y", index))
		}
		x, y := *candidate.Position.X, *candidate.Position.Y
		if !finite(x) || !finite(y) || x < 0 || y < 0 {
			return nil, invalid(fmt.Sprintf("Component index %d position must be finite and nonnegative", index))
		}
		if x+float64(component.Plan.Size.Width) > float64(canvas.Width) ||
			y+float64(component.Plan.Size.Height) > float64(canvas.Height) {
			return nil, invalid(fmt.Sprintf("Component index %d bounds exceed the UI Set canvas", index))
		}
		positions[index] = assetdomain.Position{X: x, Y: y}
	}
	return positions, nil
}

func (e *executor) persistUISet(
	ctx context.Context,
	payload CreateUISetPayload,
	components []processedUISetComponent,
) (json.RawMessage, error) {
	batchID, err := newUISetBatchID()
	if err != nil {
		return nil, fmt.Errorf("generator: create UI Set resource batch: %w", err)
	}
	objectKeys := make([]string, len(components))
	for index, component := range components {
		objectKeys[index] = fmt.Sprintf("projects/%d/uisets/%s/components/%d.png", payload.ProjectID, batchID, component.Plan.Index)
	}
	uploadedKeys, err := e.uploadUISetComponents(ctx, components, objectKeys)
	if err != nil {
		return nil, e.cleanupUISetResources(err, uploadedKeys)
	}
	cleanup := func(cause error) error { return e.cleanupUISetResources(cause, uploadedKeys) }

	contentComponents := make([]assetdomain.UIComponent, len(components))
	for index, component := range components {
		texture, encodeErr := json.Marshal(assetdomain.UITexture{URL: objectKeys[index]})
		if encodeErr != nil {
			return nil, cleanup(fmt.Errorf("generator: encode UI Component %d texture: %w", component.Plan.Index, encodeErr))
		}
		state, encodeErr := encodeUISetComponentState(component.Plan)
		if encodeErr != nil {
			return nil, cleanup(encodeErr)
		}
		contentComponents[index] = assetdomain.UIComponent{
			ID: component.Plan.Index + 1, Name: component.Plan.Name, Size: component.Plan.Size,
			Position: component.Position, Texture: texture, State: state,
			Metadata: map[string]any{"description": component.Plan.Description, "generated": true},
		}
	}
	content := assetdomain.NewAssetContent(assetdomain.AssetTypeUISet)
	content.Components = contentComponents
	encodedContent, err := assetdomain.EncodeContent(content)
	if err != nil {
		return nil, cleanup(fmt.Errorf("generator: encode UI Set content: %w", err))
	}
	dimensions, err := json.Marshal(payload.Dimensions)
	if err != nil {
		return nil, cleanup(fmt.Errorf("generator: encode UI Set dimensions: %w", err))
	}
	assetID, err := e.assets.CreateUISetAsset(ctx, &assetdomain.Asset{
		Name: strings.TrimSpace(payload.AssetName), ProjectID: payload.ProjectID, Type: assetdomain.AssetTypeUISet,
		Description: strings.TrimSpace(payload.CreativeBrief), Dimensions: dimensions, Content: encodedContent,
	})
	if err != nil {
		return nil, cleanup(fmt.Errorf("generator: create UI Set asset: %w", err))
	}
	if assetID == 0 {
		return nil, cleanup(fmt.Errorf("generator: create UI Set asset: empty result"))
	}
	return encodeExecutionResult(ExecutionResult{AssetID: assetID, Version: 1})
}

func encodeUISetComponentState(plan UISetComponentPlan) (json.RawMessage, error) {
	frames := make([]assetdomain.UIStateFrame, len(plan.States))
	for index, state := range plan.States {
		frames[index] = assetdomain.UIStateFrame{
			Name: state,
			Rect: assetdomain.UIRect{X: uint(index) * plan.Size.Width, Width: plan.Size.Width, Height: plan.Size.Height},
		}
	}
	state := assetdomain.UIComponentState{
		Kind: plan.Kind, TextureSize: assetdomain.Size{Width: plan.Size.Width * uint(len(plan.States)), Height: plan.Size.Height},
		Frames: frames,
	}
	if plan.Kind == "bar" {
		axis := "horizontal"
		if plan.Size.Height > plan.Size.Width {
			axis = "vertical"
		}
		state.RuntimeFill = &assetdomain.UIRuntimeFill{Enabled: true, Axis: axis}
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("generator: encode UI Component %d state: %w", plan.Index, err)
	}
	return encoded, nil
}

func (e *executor) uploadUISetComponents(
	ctx context.Context,
	components []processedUISetComponent,
	objectKeys []string,
) ([]string, error) {
	var mu sync.Mutex
	uploaded := make([]string, 0, len(components))
	err := runBoundedUISetJobs(ctx, len(components), maxUISetComponentConcurrency, func(jobCtx context.Context, index int) error {
		data, decodeErr := decodeUISetPNG(components[index].ImageBase64)
		if decodeErr != nil {
			return fmt.Errorf("generator: decode UI Component %d for upload: %w", components[index].Plan.Index, decodeErr)
		}
		if putErr := e.resources.PutObject(jobCtx, objectKeys[index], "image/png", data); putErr != nil {
			return fmt.Errorf("generator: upload UI Component %d: %w", components[index].Plan.Index, putErr)
		}
		mu.Lock()
		uploaded = append(uploaded, objectKeys[index])
		mu.Unlock()
		return nil
	})
	return uploaded, err
}

func (e *executor) cleanupUISetResources(cause error, keys []string) error {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), uiSetCleanupTTL)
	defer cancel()
	var cleanupErr error
	for _, key := range slices.Backward(keys) {
		if err := e.resources.DeleteObject(cleanupCtx, key); err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("generator: delete unreferenced UI Set resource %q: %w", key, err))
		}
	}
	if cleanupErr != nil {
		return errors.Join(cause, cleanupErr)
	}
	return cause
}

func runBoundedUISetJobs(
	ctx context.Context,
	count int,
	limit int,
	job func(context.Context, int) error,
) error {
	if count == 0 {
		return nil
	}
	jobCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan int)
	workers := min(count, limit)
	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error
	for range workers {
		wg.Go(func() {
			for index := range jobs {
				if err := job(jobCtx, index); err != nil {
					once.Do(func() {
						firstErr = err
						cancel()
					})
				}
			}
		})
	}
	for index := 0; index < count; index++ {
		select {
		case jobs <- index:
		case <-jobCtx.Done():
			index = count
		}
	}
	close(jobs)
	wg.Wait()
	if firstErr != nil {
		return firstErr
	}
	return ctx.Err()
}

func newUISetBatchID() (string, error) {
	value := make([]byte, uiSetBatchIDBytes)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}
