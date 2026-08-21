package prompts

import (
	"fmt"
	"strings"
)

const sceneryVisualConstraints = `Visual contract:
- The final scene and every independently generated layer MUST use classic low-resolution 2D pixel art, matching the character and object asset pipeline.
- Use deliberate pixel clusters, hard aliased edges, a restricted palette, no antialiasing, no smooth gradients, no painterly brushwork, no vector rendering, no 3D rendering, and no photorealism.
- Keep one consistent pixel density, palette, lighting model, material language, and perspective across every layer.
- Scenery should contain no characters, people, humanoids, animals, or creatures by default. If and only if the user's creative brief explicitly requests one of them, follow that explicit request and ignore this preference.
- Priority to User Intent & Scene Cohesion: Always respect the user's overall vision above all else. Construct each layer's position, framing, and depth attributes based on the holistic scene description.
- Autonomous Layer Refinement: If the user does not describe each layer in detail, autonomously expand and refine each layer's prompt to be vivid and visually rich while remaining strictly faithful to the original semantics without over-packaging.
- Seamless Scene Flow & Grounding: Overall visual smoothness and environmental harmony are the top priority. Prevent compositing disasters where individual layers clash jarringly with the background theme or style. Unless explicitly requested by the user, grounded structures (buildings, roads, terrain, platforms) must remain solidly grounded and perspective-anchored, never floating unnaturally in mid-air.
- SCENERY ASSET IS ALWAYS A 2D PANORAMIC BACKDROP PLATE (SIDE-ON / OVERLOOK VISTA): A "scenery" asset in 2D games is strictly a panoramic scenic vista and parallax environmental backdrop (like Stardew Valley's Summit cutscene, Hilltop Farm cliff overlook, or mountain/ocean backdrops). It is displayed behind the gameplay area or revealed beyond cliff edges. NEVER generate an interactive top-down playable tilemap, floor plan, or satellite grid. Even if the project game type is a Top-Down farming sim or RPG, the scenery asset itself is ALWAYS rendered looking horizontally out toward the panoramic horizon (distant rolling hills, mountain peaks, valley floor, skies, rivers, sunset vistas, framing silhouette trees), NOT a walkable tile surface!`

const sceneryPlanTemplate = `Plan the independently generated image layers for one complete layered 2D game scenery.

%s

Rules:
- Decide the number of layers needed to express the complete scene. Use 2 to 4 layers total (typically 1 backmost backdrop, plus 1 to 3 discrete midground and foreground layers).
- Return layers in back-to-front compositing order.
- STRICT ELEMENT MUTUAL EXCLUSIVITY & ORTHOGONALITY:
  * Every distinct scene entity or structure (e.g. the specific farmhouse building, the river, the tilled crops, the mountain ridge, the fence line) MUST belong to EXACTLY ONE layer.
  * NEVER duplicate or redraw an element across multiple layers.
- PROGRESSIVE DEPTH SPECIALIZATION FOR PANORAMIC VISTAS: Layer assignments must progress logically from back to front:
  * Layer 1 (Backmost Base Backdrop): The single fully opaque base backdrop containing global sky, clouds, celestial bodies, and distant mountain silhouettes spanning 75%% ~ 85%% of the canvas height.
  * Intermediate Layers (Midground Vista Landscape): Distant valley floor, rolling hills, winding river, distant farm plots, or intermediate terrain seen looking horizontally across the landscape.
  * Foremost Layers (Foreground Framing & Overlook Elements): Framing foliage (cliff edges, overlook fences, silhouette trees, foreground building profiles) framing the sides/bottom of the panoramic vista.
    - ANTI-DIORAMA / NO FLOATING ISLANDS: Foremost overlay layers MUST NOT generate self-contained circular/rectangular terrain islands or duplicate mini-ground patches. They must render only discrete framing structures on flat uniform chroma matte (#00ff00).
- Subsequent Overlay Layers: Must each represent isolated, discrete midground or foreground structures (e.g. water bodies, terrain, platform ground, trees, buildings). Overlay layers MUST NOT duplicate the sky, distant mountain ranges, or backdrop scenery belonging to Layer 1. Each overlay layer MUST leave the canvas region above or around it open for chroma-key transparency so that the backdrop layer remains visible.
  * Framing & Open Space: Size and position each overlay layer so that it harmonizes with the overall scene composition, anchoring the lower or side regions while leaving generous open space for the underlying layers.
- HOLISTIC SCENE HARMONY & CONTEXTUAL MATERIALITY: Derive each layer's opacity, lighting, perspective, and scale directly from the user's creative brief. Ensure all layers combine into one seamless, balanced composition where all key narrative elements have clear visual presence without accidental occlusion or unnatural ghosting.
- Give every layer a short unique name and a self-contained image-generation brief.
- Each layer brief must describe only that layer's visual content and its intended placement, framing, scale, depth, and relationship to the full canvas.
- SPATIAL POSITIONING & HARMONY: In each layer brief, define clear, explicit canvas coordinates and framing (e.g. "distant farmhouse in the valley at center-left X=35%%~65%%, Y=40%%~65%%", "river winding through valley at X=10%%~35%%", "foreground framing pine trees at left edge X=0%%~20%%"). Ensure that all elements interlock seamlessly in 2D space without clashing, blocking, or overlapping awkwardly.
- Coordinate silhouettes, overlaps, palette, lighting, perspective, and level of detail across all layer briefs so separately generated images form one coherent scene.
%s
%s
- Do not add IDs; the backend assigns stable IDs from response order.
- Return only the fields defined by the supplied JSON schema. Do not return explanations, coordinates, resources, or metadata.

Scenery asset:
<asset_name>%s</asset_name>
<creative_brief>%s</creative_brief>
<dimensions width="%d" height="%d" />
<perspective>%s</perspective>`

type SceneryPlanInput struct {
	AssetName           string
	CreativeBrief       string
	Perspective         string
	Width               uint
	Height              uint
	HasProjectReference bool
	ReviewFeedback      string
}

func SceneryPlan(input SceneryPlanInput) string {
	visualGuidance := `- Visual Guidance: No reference images are supplied. Autonomously derive and refine each layer's aesthetic nuances (such as lighting mood, color harmony, material textures, and environmental tone) to form a cohesive, rich 2D pixel art scene based on the user's creative brief.`
	if input.HasProjectReference {
		visualGuidance = `- Visual Prototype Guidance: Reference Image 1 is the project's visual prototype / style reference. Visually inspect its palette, lighting model, contrast, pixel granularity, outline treatment, material language, and visual world conventions. Decompose the scene into harmonious layers and autonomously refine each layer's generation brief to strictly adhere to the visual art style demonstrated in this prototype.`
	}

	feedbackSection := ""
	if strings.TrimSpace(input.ReviewFeedback) != "" {
		feedbackSection = fmt.Sprintf("- PREVIOUS REVIEW CORRECTIONS: The previous generation was rejected with these notes:\n%s\nYou MUST explicitly fix these issues in this plan.", strings.TrimSpace(input.ReviewFeedback))
	}

	return fmt.Sprintf(
		sceneryPlanTemplate,
		sceneryVisualConstraints,
		visualGuidance,
		feedbackSection,
		strings.TrimSpace(input.AssetName),
		strings.TrimSpace(input.CreativeBrief),
		input.Width,
		input.Height,
		strings.TrimSpace(input.Perspective),
	)
}

const sceneryLayerTemplate = `Create exactly one production-ready image layer for a layered 2D game scenery.

%s

Pipeline requirements:
%s
- STRICT LAYER ISOLATION & ANTI-DIORAMA: Generate ONLY the elements explicitly described in <layer_creative_brief>. DO NOT generate a complete scene or elements belonging to other layers.
- 2D PANORAMIC SIDE-ON VISTA PERSPECTIVE: Render this layer strictly as a slice of a 2D horizontal panoramic scenic vista (Side-On / Overlook perspective). NEVER draw a top-down aerial map or floor grid.
- For overlay layers (non-backmost): DO NOT draw background sky, sun, clouds, or entire landscapes. DO NOT draw self-contained floating island dioramas, separate floating ground patches, or duplicate rivers/roads. The entire space outside and around the requested subject MUST remain the flat solid uniform matte color (#00ff00) so that underlying layers remain fully visible.
- PROGRESSIVE REFERENCE COGNITION & PHYSICAL ALIGNMENT: When a reference image showing preceding layers (such as foreground structures or midground terrain) is supplied, you MUST visually inspect where every structure (doorways, walls, fences, paths, rivers, trees, and light sources) is placed in Reference Image 1:
  * Connect and physically align your layer with those structures (e.g. if the reference image shows a house, draw paths leading directly to its doorway and wrap terrain around its foundation; if drawing a backdrop, draw horizons and mountains fitting naturally behind the existing structures).
  * Align light reflections, highlights, and shadow directions so the whole scene feels unified under the same lighting.
- Keep the complete layer artwork inside the canvas with clean separation from the matte background.
- Do not add text, labels, logos, watermarks, borders, frames, UI, or a preview of the assembled scene.
- Match the shared scenery brief, perspective, palette, lighting, pixel density, and material treatment so this independently generated layer can be composited seamlessly with every other layer.
- Treat any supplied reference image as visual-language guidance only. Do not copy its composition or recognizable content.

Scenery asset:
<asset_name>
%s
</asset_name>

Shared creative brief:
<creative_brief>
%s
</creative_brief>

Requested layer:
<layer_id>%d</layer_id>
<layer_name>%s</layer_name>
<layer_creative_brief>
%s
</layer_creative_brief>

Final canvas:
<dimensions width="%d" height="%d" />
<perspective>%s</perspective>

Project visual context:
<project_name>%s</project_name>
<game_type>%s</game_type>
<target_platform>%s</target_platform>
<project_description>%s</project_description>
<reference_supplied>%t</reference_supplied>`

type SceneryLayerInput struct {
	AssetName          string
	CreativeBrief      string
	Perspective        string
	ProjectName        string
	GameType           string
	TargetPlatform     string
	ProjectDescription string
	Width              uint
	Height             uint
	LayerID            uint
	LayerName          string
	LayerCreativeBrief string
	HasReference       bool
	IsBackmost         bool
}

func SceneryLayer(input SceneryLayerInput, backgroundConstraint string) string {
	backgroundContract := strings.TrimSpace(backgroundConstraint)
	if input.IsBackmost {
		backgroundContract = `- This is the backmost scenery layer (the background backdrop). It may cover the complete canvas edge to edge and may be fully opaque.
- Paint continuous sky, atmospheric lighting, and distant horizon scenery. Give prominent background features (mountains, giant landmarks, celestial bodies) grand vertical scale (typically spanning 75%%~85%% of the frame height) to ensure ample clearance above lower overlays.
- Do not leave empty margins, frames, borders, or solid green matte on this backmost layer.`
	} else {
		backgroundContract += `
- This is an overlay layer. Every pixel outside the requested artwork must remain the exact solid matte colour.
- Leave a continuous matte-only border around the canvas edge for isolated elements, while grounded structures must span continuously across the canvas and anchor firmly to the bottom of the frame.
- For continuous terrain / ground layers: The terrain surface (grass, soil, water, paths) must fill the lower 55%%~75%% of the frame with fully opaque, rich terrain pixels. Apply the uniform #00ff00 green matte ONLY to the upper 25%%~45%% region where the sky/backdrop will show through.
- Generate ONLY the subject described in the layer brief. DO NOT draw a self-contained floating island diorama or duplicate ground patches.
- If drawing upright objects (e.g. a building, fence, or tree), draw ONLY the object itself resting cleanly on the canvas bottom/ground plane against pure solid green matte (#00ff00).
- Position and frame the requested subject according to the layer brief, leaving the surrounding canvas area as flat solid matte colour (#00ff00) so underlying layers remain visible.
- If a reference image of a preceding layer is supplied, visually align all light reflection paths, highlights, and contact shadows with the light sources in that reference image.
- Treat material opacity and transparency contextually based on the physical and atmospheric nature of the scene brief.
- Do not return the matte-only input unchanged; the requested layer must contain clearly non-matte opaque subject pixels.
- Grounding alignment: Grounded elements (such as terrain, roads, platforms, or building bases) must be framed and anchored firmly in the lower area of the canvas.`
	}
	return fmt.Sprintf(
		sceneryLayerTemplate,
		sceneryVisualConstraints,
		backgroundContract,
		strings.TrimSpace(input.AssetName),
		strings.TrimSpace(input.CreativeBrief),
		input.LayerID,
		strings.TrimSpace(input.LayerName),
		strings.TrimSpace(input.LayerCreativeBrief),
		input.Width,
		input.Height,
		strings.TrimSpace(input.Perspective),
		strings.TrimSpace(input.ProjectName),
		strings.TrimSpace(input.GameType),
		strings.TrimSpace(input.TargetPlatform),
		strings.TrimSpace(input.ProjectDescription),
		input.HasReference,
	)
}

const sceneryLayoutAnalysisTemplate = `Inspect every attached processed image, critically review the composited scene for flaws, and propose the final layout for one layered 2D game scenery.

%s

Critical Review & Calibration:
- Always inspect the attached layer artwork and composite relationships with a rigorous, critical eye.
- STRICT REJECTION CRITERIA (Set "approved": false if ANY of the following defects are present):
  1. Duplicate major elements across layers: e.g. multiple different farmhouses, conflicting multiple rivers/roads, or redundant buildings drawn on more than one layer.
  2. Floating island diorama & center blockage: An overlay layer generated its own self-contained terrain island or miniature scene that sits in the center and occludes more than 40%% of the underlying midground/background.
  3. Floating/unanchored structures: roads, bridges, platforms, or buildings hovering in the air with gaps exposing the background below.
  4. Broken or truncated spans: horizontal structures cut off abruptly before reaching the canvas boundary, or rectangular/box-like cutout boundaries.
  5. Redundant or clashing backgrounds: overlay layers that accidentally drew their own sky, sun, or duplicate scenery instead of cleanly isolated subjects.
  6. Perspective, horizon, and depth alignment: conflicting perspective grids, vanishing points, or clashing scales/styles between layers. Ensure background focal elements maintain a strong, balanced visual presence above foreground/midground horizons.
  7. Context-inappropriate transparency or ghosting: Solid physical terrain and structures should avoid accidental semi-transparency that creates bizarre submerged illusions.
- Calibrate layout parameters (position X/Y, scale, and zIndex) to correct and align any minor flaws.
- If ANY layer has unfixable generation flaws or violates the strict rejection criteria, set "approved": false.
- If and only if the scene composition is harmonious, well-grounded, cleanly separated, free of duplicate entities, and satisfies the user's creative brief (or after calibrations), set "approved": true.
- Summarize your visual assessment and any adjustments made in "review_notes". When rejecting ("approved": false), provide clear, actionable feedback in "review_notes" specifying which layers duplicated elements or caused clashing occlusions.

Layout rules:
- Return exactly one layout for every supplied layer ID. Do not invent, omit, or duplicate IDs.
- Every attached image is already registered to the complete final canvas at the requested dimensions (%d x %d). Transparent pixels are intentional padding, and visible pixels already express the planned global placement.
- Default to position (0, 0), scale (1, 1), and rotation 0. Change these only when inspection of the actual attached pixels proves a correction is necessary; never transform already-correct content out of its intended canvas region.
- DO NOT return the coordinate of where the subject appears inside the image as the position offset. Leaving position at (0, 0) keeps the artwork exactly where it was drawn. Only return a non-zero position offset if you specifically intend to shift/translate the layer to fix an alignment flaw.
- The first attached image is the authoritative opaque full-canvas backdrop. Keep it at position (0, 0), scale (1, 1), rotation 0, opacity 1, and give it the unique lowest zIndex so it can never cover another layer.
- Derive opacity, scale, and placement from the actual attached pixels, physical perspective, and creative intent. Default to opacity 1.0 for solid elements, adjusting only when atmospheric effects or specific creative intent warrants translucency.
- Ensure foreground terrain, roads, highways, and platform elements sit naturally in the lower region of the canvas to anchor the scene and create clean depth over the background.
- Use canvas pixels with the canvas top-left as (0, 0), positive X to the right, and positive Y downward.
- Position is the top-left offset of the scaled layer before rotation. Rotation is clockwise in degrees around the scaled layer center.
- Scale X and Y must be finite and greater than zero. Opacity must be from 0 through 1. zIndex must be an integer.
- Keep every transformed layer at least partially intersecting the canvas.
- Use the actual attached pixels, shared creative intent, perspective, and depth relationships to choose placement, scale, rotation, opacity, and stacking order.
- Return only the fields defined by the supplied JSON schema.

Scenery asset:
<asset_name>%s</asset_name>
<creative_brief>%s</creative_brief>
<dimensions width="%d" height="%d" />
<perspective>%s</perspective>

Project visual context:
<project_name>%s</project_name>
<game_type>%s</game_type>
<target_platform>%s</target_platform>
<project_description>%s</project_description>

Attached layers:
%s`

type SceneryLayoutLayerInput struct {
	ID   uint
	Name string
}

type SceneryLayoutAnalysisInput struct {
	AssetName          string
	CreativeBrief      string
	Perspective        string
	ProjectName        string
	GameType           string
	TargetPlatform     string
	ProjectDescription string
	Width              uint
	Height             uint
	Layers             []SceneryLayoutLayerInput
}

func SceneryLayoutAnalysis(input SceneryLayoutAnalysisInput) string {
	var layerList strings.Builder
	for index, layer := range input.Layers {
		fmt.Fprintf(&layerList, "Attached image %d corresponds to layer ID %d named %q.\n", index+1, layer.ID, strings.TrimSpace(layer.Name))
	}
	return fmt.Sprintf(
		sceneryLayoutAnalysisTemplate,
		sceneryVisualConstraints,
		input.Width, input.Height,
		strings.TrimSpace(input.AssetName), strings.TrimSpace(input.CreativeBrief), input.Width, input.Height,
		strings.TrimSpace(input.Perspective), strings.TrimSpace(input.ProjectName),
		strings.TrimSpace(input.GameType), strings.TrimSpace(input.TargetPlatform), strings.TrimSpace(input.ProjectDescription),
		strings.TrimSpace(layerList.String()),
	)
}
