package prompts

import (
	"fmt"
	"strings"
)

type UISetComponentInput struct {
	Index       int
	Name        string
	Description string
}

type UISetPlanInput struct {
	AssetName          string
	CreativeBrief      string
	Style              string
	ProjectStyle       string
	ProjectName        string
	GameType           string
	TargetPlatform     string
	ProjectDescription string
	Width              uint
	Height             uint
	Components         []UISetComponentInput
}

const uiSetPlanTemplate = `Plan a comprehensive, cohesive collection of independently generated UI Components for one 2D game project.

AUTHORITATIVE PROJECT CONTEXT:
- The Project art style, game type, and game description are the source of truth. Every Component must belong to that same game and visual system.
- Use the requested UI style only to refine the Project style. If they conflict, preserve the Project style.
- Infer all practical Components and variants needed by the gameplay, menus, HUD, controls, and feedback explicitly supported by the game description. Be comprehensive without inventing unrelated mechanics or generic genre clutter.

COMPONENT AND STATE RULES:
- The requested Components are mandatory seeds. Return them first, in the same order, with request_index equal to their supplied index. Preserve their supplied names and descriptions.
- After the requested Components, add useful inferred Components with request_index -1. Give every Component a unique short name and self-contained description.
- Classify each Component as panel, button, icon, indicator, bar, slot, cursor, badge, or other.
- Size is the in-game display size of one state frame. Assign a positive integer width and height that fits within the complete UI Set canvas.
- Enumerate all meaningful visual states in display order. Examples include button normal/hover/pressed/disabled and heart health full/damaged/empty. Do not split states into separate Components.
- A bar is an engine-composited frame: it must return kind bar and exactly one state named empty. Its generated image must never contain health, mana, stamina, experience, progress, liquid, or another dynamic fill.
- For non-bar Components, include every visually distinct state that a game engine would select at runtime; use one default state only when no alternate visual state is meaningful.
- Keep state strips practical: at most eight states and no repeated state names.
- Do not render words, letters, numerals, labels, or pseudo-text into image assets. Plan empty text containers and icon-only controls where applicable; real text is rendered by the engine.

OUTPUT RULES:
- Each planned entry defines one independent PNG resource. Its states will be placed left-to-right in one horizontal sprite strip, one equal-sized frame per state.
- A UI Set has no Tiles, Tile size, grid, shape, footprint, occupied cells, atlas, or image-splitting layer.
- Do not choose positions. Layout is handled later without changing display sizes, states, content, IDs, or order.
- Return only fields defined by the supplied JSON schema.

UI Set:
<asset_name>%s</asset_name>
<creative_brief>%s</creative_brief>
<requested_ui_style>%s</requested_ui_style>
<canvas width="%d" height="%d" />

Project:
<project_name>%s</project_name>
<game_type>%s</game_type>
<target_platform>%s</target_platform>
<project_style>%s</project_style>
<project_description>%s</project_description>

Mandatory requested Components:
%s`

func UISetPlan(input UISetPlanInput) string {
	components := make([]string, len(input.Components))
	for index, component := range input.Components {
		components[index] = fmt.Sprintf(
			`<component request_index="%d"><name>%s</name><description>%s</description></component>`,
			component.Index,
			strings.TrimSpace(component.Name),
			strings.TrimSpace(component.Description),
		)
	}
	return fmt.Sprintf(
		uiSetPlanTemplate,
		strings.TrimSpace(input.AssetName), strings.TrimSpace(input.CreativeBrief), strings.TrimSpace(input.Style),
		input.Width, input.Height, strings.TrimSpace(input.ProjectName), strings.TrimSpace(input.GameType),
		strings.TrimSpace(input.TargetPlatform), strings.TrimSpace(input.ProjectStyle), strings.TrimSpace(input.ProjectDescription),
		strings.Join(components, "\n"),
	)
}

type UISetComponentGenerationInput struct {
	AssetName          string
	CreativeBrief      string
	Style              string
	ProjectStyle       string
	ProjectName        string
	GameType           string
	TargetPlatform     string
	ProjectDescription string
	ComponentName      string
	ComponentBrief     string
	Kind               string
	States             []string
	FrameWidth         uint
	FrameHeight        uint
	HasReference       bool
}

const uiSetComponentTemplate = `Create exactly one production-ready 2D game UI Component sprite strip on a solid pure green #00ff00 matte.

NON-OVERRIDABLE OUTPUT CONTRACT:
- Return one image containing exactly %d equal frames in one horizontal row. Each frame is %dx%d pixels in the final asset, and the complete strip is %dx%d pixels.
- Frames from left to right are: %s.
- Draw exactly one visual state in each frame. Keep the Component's silhouette, scale, materials, lighting, pixel density, and alignment consistent across every state.
- Mentally divide the strip into the stated equal cells before drawing. Every cell must contain clearly visible non-green Component pixels centered within that cell; never leave a cell blank, merge adjacent states, or return fewer states than listed.
- Keep every visible Component pixel inside its frame. Use only pure green #00ff00 outside the Component; do not use that colour in the Component itself.
- Produce no extra variants, preview canvas, labels, separators, guides, borders around frame cells, scenery, characters, logos, watermarks, words, letters, numerals, or pseudo-text.
- Match the Project art style as the authority. The requested UI style may refine but never replace the Project's palette, rendering language, or level of detail.
- Treat supplied references only as visual-language context. Do not copy text, layout, logos, or recognizable content.

STATE AND ENGINE CONTRACT:
- This single image is the complete resource for Component %q, kind %q.
- Component brief: %s
- If this is a bar, render only an empty reusable frame/background with a transparent or visually empty interior. Do not render any health, mana, stamina, experience, progress, liquid, colour fill, or fill percentage; the game engine supplies dynamic fill.
- If this is an indicator such as a heart, clearly render every listed state, including its full, intermediate/damaged, and empty/disabled appearance where listed.

UI Set:
<asset_name>%s</asset_name>
<creative_brief>%s</creative_brief>
<requested_ui_style>%s</requested_ui_style>

Project:
<project_name>%s</project_name>
<game_type>%s</game_type>
<target_platform>%s</target_platform>
<project_style>%s</project_style>
<project_description>%s</project_description>
<reference_supplied>%t</reference_supplied>`

func UISetComponent(input UISetComponentGenerationInput) string {
	stateList := make([]string, len(input.States))
	for index, state := range input.States {
		stateList[index] = fmt.Sprintf("%d=%s", index, strings.TrimSpace(state))
	}
	return fmt.Sprintf(
		uiSetComponentTemplate,
		len(input.States), input.FrameWidth, input.FrameHeight, input.FrameWidth*uint(len(input.States)), input.FrameHeight,
		strings.Join(stateList, ", "), strings.TrimSpace(input.ComponentName), strings.TrimSpace(input.Kind),
		strings.TrimSpace(input.ComponentBrief), strings.TrimSpace(input.AssetName), strings.TrimSpace(input.CreativeBrief),
		strings.TrimSpace(input.Style), strings.TrimSpace(input.ProjectName), strings.TrimSpace(input.GameType),
		strings.TrimSpace(input.TargetPlatform), strings.TrimSpace(input.ProjectStyle), strings.TrimSpace(input.ProjectDescription),
		input.HasReference,
	)
}

type UISetLayoutComponentInput struct {
	Index  uint
	Name   string
	Kind   string
	Width  uint
	Height uint
}

type UISetLayoutInput struct {
	AssetName          string
	CreativeBrief      string
	Style              string
	ProjectStyle       string
	GameType           string
	ProjectDescription string
	Width              uint
	Height             uint
	Components         []UISetLayoutComponentInput
}

const uiSetLayoutTemplate = `Inspect every attached UI Component sprite strip and assign its in-game top-left position on one UI Set canvas.

Rules:
- Return exactly one position for every supplied Component index. Do not omit, duplicate, or invent indexes.
- Use canvas pixels with top-left (0,0), positive X rightward, and positive Y downward.
- Positions must be finite and nonnegative. The Component's one-state display width and height must remain fully inside the %dx%d canvas.
- Attached images may contain multiple horizontal state frames; layout uses only the supplied one-state display size, never the full strip width.
- Preserve Component order, content, state frames, sizes, and IDs. Return only index and position.
- Overlap is allowed and may be intentional for layered UI composition.
- Build a usable game interface hierarchy that matches the Project style, game type, description, and requested UI brief. Avoid covering the whole playfield unless the brief calls for a full-screen menu.

UI Set: %s
Creative brief: %s
Requested UI style: %s
Project style: %s
Game type: %s
Project description: %s

Attached Components:
%s`

func UISetLayout(input UISetLayoutInput) string {
	components := make([]string, len(input.Components))
	for index, component := range input.Components {
		components[index] = fmt.Sprintf(
			"Attached image %d: index=%d, name=%q, kind=%s, display_size=%dx%d.",
			index+1, component.Index, strings.TrimSpace(component.Name), strings.TrimSpace(component.Kind), component.Width, component.Height,
		)
	}
	return fmt.Sprintf(
		uiSetLayoutTemplate, input.Width, input.Height, strings.TrimSpace(input.AssetName), strings.TrimSpace(input.CreativeBrief),
		strings.TrimSpace(input.Style), strings.TrimSpace(input.ProjectStyle), strings.TrimSpace(input.GameType),
		strings.TrimSpace(input.ProjectDescription), strings.Join(components, "\n"),
	)
}

type UISetComponentEditInput struct {
	CreativeBrief string
	Name          string
	Kind          string
	States        []string
	FrameWidth    uint
	FrameHeight   uint
	ProjectStyle  string
	GameType      string
	ProjectBrief  string
}

const uiSetComponentEditTemplate = `Regenerate exactly one existing 2D game UI Component sprite strip.

NON-OVERRIDABLE RULES:
- The first reference image is the authoritative current Component. Preserve its horizontal frame count, frame order, per-frame %dx%d display size, complete %dx%d strip canvas, alignment, transparency silhouette, Project art style, and visual identity.
- Frames from left to right are: %s.
- Apply only this edit: %s
- Component name: %s
- Component kind: %s
- Project style: %s
- Game type: %s
- Project description: %s
- Return one image on a solid pure green #00ff00 matte, with no extra variants, labels, separators, text, numerals, pseudo-text, logos, watermarks, or scenery.
- Other supplied references are surrounding UI Set or user references. Use them only to preserve visual cohesion.
- If kind is bar, preserve an empty frame/background only. Never add dynamic fill, liquid, percentage, health, mana, stamina, experience, or progress content.`

func UISetComponentEdit(input UISetComponentEditInput) string {
	return fmt.Sprintf(
		uiSetComponentEditTemplate,
		input.FrameWidth, input.FrameHeight, input.FrameWidth*uint(len(input.States)), input.FrameHeight,
		strings.Join(input.States, ", "), strings.TrimSpace(input.CreativeBrief), strings.TrimSpace(input.Name),
		strings.TrimSpace(input.Kind), strings.TrimSpace(input.ProjectStyle), strings.TrimSpace(input.GameType), strings.TrimSpace(input.ProjectBrief),
	)
}
