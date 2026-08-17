package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/1024XEngineer/Holonic-Asset/internal/module/generator/imageclient"
	"github.com/1024XEngineer/Holonic-Asset/internal/module/generator/prompts"
	imageprocessor "github.com/1024XEngineer/Holonic-Asset/internal/module/processor/image"
	assetdomain "github.com/1024XEngineer/Holonic-Asset/internal/module/workspace/asset"
	projectdomain "github.com/1024XEngineer/Holonic-Asset/internal/module/workspace/project"
)

const maxUISetEditReferences = 8

type uiSetEditTarget struct {
	ComponentIndex uint
	Component      assetdomain.UIComponent
	Plan           UISetComponentPlan
	TextureURL     string
}

func (e *executor) editUISetComponents(
	ctx context.Context,
	payload EditUISetComponentsPayload,
) (json.RawMessage, error) {
	indexes, err := parseUISetComponentPaths(payload.TargetAssetPaths)
	if err != nil {
		return nil, err
	}
	asset, err := e.assets.GetDetail(ctx, payload.AssetID)
	if err != nil {
		return nil, fmt.Errorf("generator: load UI Set asset %d: %w", payload.AssetID, err)
	}
	if asset.ID != payload.AssetID {
		return nil, invalidTaskPayload("UI Set asset %d was not found", payload.AssetID)
	}
	if asset.ProjectID != payload.ProjectID {
		return nil, invalidTaskPayload("UI Set asset %d does not belong to project %d", payload.AssetID, payload.ProjectID)
	}
	if asset.Type != assetdomain.AssetTypeUISet {
		return nil, invalidTaskPayload("asset %d must have type uiset", payload.AssetID)
	}
	if asset.Version == 0 {
		return nil, invalidTaskPayload("UI Set asset %d has no current version", payload.AssetID)
	}
	content, err := asset.DecodeContent()
	if err != nil {
		return nil, fmt.Errorf("generator: decode UI Set asset %d content: %w", payload.AssetID, err)
	}
	project, err := e.projects.GetDetail(ctx, payload.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("generator: load project %d for UI Set edit: %w", payload.ProjectID, err)
	}
	if project == nil {
		return nil, invalidTaskPayload("project %d was not found", payload.ProjectID)
	}

	targets, err := resolveUISetEditTargets(indexes, content.Components)
	if err != nil {
		return nil, err
	}
	componentReferences, err := e.resolveUISetComponentReferences(ctx, content.Components)
	if err != nil {
		return nil, err
	}
	sharedReferences, err := e.resolveUISetReferences(ctx, payload.Reference, project.Reference)
	if err != nil {
		return nil, err
	}

	replacements := make([]processedUISetComponent, len(targets))
	err = runBoundedUISetJobs(ctx, len(targets), maxUISetComponentConcurrency, func(jobCtx context.Context, index int) error {
		target := targets[index]
		references := uiSetEditReferenceOrder(target.ComponentIndex, componentReferences, sharedReferences)
		replacement, generateErr := e.generateUISetEditReplacement(jobCtx, payload, project, target, references)
		if generateErr != nil {
			return generateErr
		}
		replacements[index] = replacement
		return nil
	})
	if err != nil {
		return nil, err
	}

	batchID, err := newUISetBatchID()
	if err != nil {
		return nil, fmt.Errorf("generator: create edited UI Set resource batch: %w", err)
	}
	objectKeys := make([]string, len(targets))
	for index, target := range targets {
		objectKeys[index] = fmt.Sprintf(
			"projects/%d/uisets/%d/revisions/%s/components/%d.png",
			payload.ProjectID, payload.AssetID, batchID, target.ComponentIndex,
		)
	}
	uploadedKeys, err := e.uploadUISetComponents(ctx, replacements, objectKeys)
	if err != nil {
		return nil, e.cleanupUISetResources(err, uploadedKeys)
	}
	cleanup := func(cause error) error { return e.cleanupUISetResources(cause, uploadedKeys) }
	for index, target := range targets {
		texture, replaceErr := replaceUISetTextureURL(content.Components[target.ComponentIndex].Texture, objectKeys[index])
		if replaceErr != nil {
			return nil, cleanup(fmt.Errorf("generator: update UI Component %d texture: %w", target.ComponentIndex, replaceErr))
		}
		content.Components[target.ComponentIndex].Texture = texture
	}
	encoded, err := assetdomain.EncodeContent(content)
	if err != nil {
		return nil, cleanup(fmt.Errorf("generator: encode edited UI Set asset %d content: %w", payload.AssetID, err))
	}
	record, err := e.assets.CreateRecord(ctx, &assetdomain.AssetRecord{AssetID: payload.AssetID, Content: encoded}, asset.Version)
	if err != nil {
		return nil, cleanup(fmt.Errorf("generator: revise UI Set asset %d: %w", payload.AssetID, err))
	}
	if record == nil || record.AssetID != payload.AssetID || record.Version == 0 {
		return nil, cleanup(fmt.Errorf("generator: revise UI Set asset %d: empty result", payload.AssetID))
	}
	return encodeExecutionResult(ExecutionResult{AssetID: payload.AssetID, Version: record.Version})
}

func resolveUISetEditTargets(
	indexes []uint,
	components []assetdomain.UIComponent,
) ([]uiSetEditTarget, error) {
	targets := make([]uiSetEditTarget, len(indexes))
	for targetIndex, componentIndex := range indexes {
		if componentIndex >= uint(len(components)) {
			return nil, invalidTaskPayload("UI Set component index %d is out of range", componentIndex)
		}
		component := components[componentIndex]
		textureURL, err := decodeUISetTextureURL(component.Texture)
		if err != nil {
			return nil, invalidTaskPayload("UI Set component %d texture: %v", componentIndex, err)
		}
		plan, err := decodePersistedUISetComponentPlan(componentIndex, component)
		if err != nil {
			return nil, invalidTaskPayload("UI Set component %d state: %v", componentIndex, err)
		}
		targets[targetIndex] = uiSetEditTarget{
			ComponentIndex: componentIndex, Component: component, Plan: plan, TextureURL: textureURL,
		}
	}
	return targets, nil
}

func decodePersistedUISetComponentPlan(index uint, component assetdomain.UIComponent) (UISetComponentPlan, error) {
	if strings.TrimSpace(component.Name) == "" || component.Size.Width == 0 || component.Size.Height == 0 {
		return UISetComponentPlan{}, fmt.Errorf("name and positive size are required")
	}
	plan := UISetComponentPlan{
		Index: index, Name: component.Name, Kind: "other", States: []string{"default"}, Size: component.Size,
	}
	if len(component.State) == 0 || string(component.State) == "null" {
		return plan, nil
	}
	state, typedState := decodeTypedUISetComponentState(component.State)
	if !typedState {
		return plan, nil // Legacy arbitrary state JSON remains editable as a single default frame.
	}
	if !validUISetComponentKind(strings.ToLower(strings.TrimSpace(state.Kind))) || len(state.Frames) == 0 {
		return plan, nil
	}
	plan.Kind = strings.ToLower(strings.TrimSpace(state.Kind))
	plan.States = make([]string, len(state.Frames))
	for frameIndex, frame := range state.Frames {
		if strings.TrimSpace(frame.Name) == "" || frame.Rect.X != uint(frameIndex)*component.Size.Width ||
			frame.Rect.Y != 0 || frame.Rect.Width != component.Size.Width || frame.Rect.Height != component.Size.Height {
			return UISetComponentPlan{}, fmt.Errorf("frame %d does not match the Component display size and horizontal state order", frameIndex)
		}
		plan.States[frameIndex] = strings.TrimSpace(frame.Name)
	}
	wantTexture := assetdomain.Size{Width: component.Size.Width * uint(len(state.Frames)), Height: component.Size.Height}
	if state.TextureSize != wantTexture {
		return UISetComponentPlan{}, fmt.Errorf("textureSize does not match the state strip")
	}
	if _, err := validateUISetPlanStates(int(index), plan.Kind, plan.States); err != nil {
		return UISetComponentPlan{}, err
	}
	return plan, nil
}

func decodeTypedUISetComponentState(raw json.RawMessage) (assetdomain.UIComponentState, bool) {
	var state assetdomain.UIComponentState
	err := json.Unmarshal(raw, &state)
	return state, err == nil
}

func decodeUISetTextureURL(raw json.RawMessage) (string, error) {
	var texture assetdomain.UITexture
	if err := json.Unmarshal(raw, &texture); err != nil {
		return "", err
	}
	texture.URL = strings.TrimSpace(texture.URL)
	if texture.URL == "" {
		return "", fmt.Errorf("url is required")
	}
	return texture.URL, nil
}

func replaceUISetTextureURL(raw json.RawMessage, objectKey string) (json.RawMessage, error) {
	var texture map[string]json.RawMessage
	if err := json.Unmarshal(raw, &texture); err != nil || texture == nil {
		texture = make(map[string]json.RawMessage)
	}
	encodedURL, err := json.Marshal(objectKey)
	if err != nil {
		return nil, err
	}
	texture["url"] = encodedURL
	encoded, err := json.Marshal(texture)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func (e *executor) resolveUISetComponentReferences(
	ctx context.Context,
	components []assetdomain.UIComponent,
) (map[uint]string, error) {
	result := make(map[uint]string, len(components))
	for index, component := range components {
		url, err := decodeUISetTextureURL(component.Texture)
		if err != nil {
			continue // Unmodeled legacy Components are preserved but cannot provide visual context.
		}
		resolved, err := e.references.ResolveReference(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("generator: resolve UI Component %d texture: %w", index, err)
		}
		if strings.TrimSpace(resolved) == "" {
			return nil, fmt.Errorf("generator: resolve UI Component %d texture: empty result", index)
		}
		result[uint(index)] = resolved
	}
	return result, nil
}

func uiSetEditReferenceOrder(
	target uint,
	componentReferences map[uint]string,
	shared []string,
) []string {
	result := make([]string, 0, maxUISetEditReferences)
	seen := make(map[string]struct{}, maxUISetEditReferences)
	appendValue := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || len(result) >= maxUISetEditReferences {
			return
		}
		if _, duplicate := seen[value]; duplicate {
			return
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	appendValue(componentReferences[target])
	for _, value := range shared {
		appendValue(value)
	}
	for _, index := range slices.Sorted(maps.Keys(componentReferences)) {
		if index != target {
			appendValue(componentReferences[index])
		}
	}
	return result
}

func (e *executor) generateUISetEditReplacement(
	ctx context.Context,
	payload EditUISetComponentsPayload,
	project *projectdomain.Project,
	target uiSetEditTarget,
	references []string,
) (processedUISetComponent, error) {
	if len(references) == 0 {
		return processedUISetComponent{}, fmt.Errorf("generator: edit UI Component %d: current texture reference is required", target.ComponentIndex)
	}
	plan := target.Plan
	textureWidth := plan.Size.Width * uint(len(plan.States))
	prompt := prompts.UISetComponentEdit(prompts.UISetComponentEditInput{
		CreativeBrief: payload.CreativeBrief, Name: plan.Name, Kind: plan.Kind, States: plan.States,
		FrameWidth: plan.Size.Width, FrameHeight: plan.Size.Height, ProjectStyle: project.Style,
		GameType: project.GameType, ProjectBrief: project.Description,
	})
	generated, err := e.images.Generate(ctx, &imageclient.GenerateRequest{
		Prompt: prompt, ReferenceImages: references, Size: fmt.Sprintf("%dx%d", textureWidth, plan.Size.Height),
	})
	if err != nil {
		return processedUISetComponent{}, fmt.Errorf("generator: regenerate UI Component %d: %w", target.ComponentIndex, err)
	}
	if generated == nil || len(generated.Images) != 1 || strings.TrimSpace(generated.Images[0].Base64) == "" {
		return processedUISetComponent{}, fmt.Errorf("generator: regenerate UI Component %d: expected exactly one image", target.ComponentIndex)
	}
	removed, err := e.processor.RemoveBackground(ctx, &imageprocessor.RemoveBackgroundRequest{
		ImageBase64: generated.Images[0].Base64, MatteColor: imageprocessor.DefaultMatteColor,
	})
	if err != nil {
		return processedUISetComponent{}, fmt.Errorf("generator: remove edited UI Component %d background: %w", target.ComponentIndex, err)
	}
	if removed == nil || strings.TrimSpace(removed.ImageBase64) == "" {
		return processedUISetComponent{}, fmt.Errorf("generator: remove edited UI Component %d background: empty result", target.ComponentIndex)
	}
	normalizedBase64, normalizedMediaType, err := e.normalizeUISetStateStrip(ctx, removed.ImageBase64, plan, uiSetRasterMode(project.Style))
	if err != nil {
		return processedUISetComponent{}, fmt.Errorf("generator: normalize edited UI Component %d: %w", target.ComponentIndex, err)
	}
	verified, err := e.processor.Verify(ctx, &imageprocessor.VerifyRequest{
		ImageBase64: normalizedBase64, Profile: imageprocessor.ProfileGeneric,
		ExpectedMatteColor: imageprocessor.DefaultMatteColor,
	})
	if err != nil {
		return processedUISetComponent{}, fmt.Errorf("generator: verify edited UI Component %d: %w", target.ComponentIndex, err)
	}
	if verified == nil || !verified.Passed {
		return processedUISetComponent{}, fmt.Errorf("generator: verify edited UI Component %d failed", target.ComponentIndex)
	}
	if err := validateUISetSpriteStrip(normalizedBase64, plan); err != nil {
		return processedUISetComponent{}, err
	}
	return processedUISetComponent{Plan: plan, ImageBase64: normalizedBase64, MediaType: normalizedMediaType}, nil
}
