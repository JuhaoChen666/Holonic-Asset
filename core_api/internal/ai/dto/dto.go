package dto

// GenerateCharacterRequest identifies the Character prototype to generate.
type GenerateCharacterRequest struct {
	AssetID         uint     `json:"assetId"`
	AssetResourceID uint     `json:"assetResourceId"`
	Prompt          string   `json:"prompt"`
	ReferenceURLs   []string `json:"referenceUrls,omitempty"`
}

type GenerateCharacterResponse struct {
	TaskID uint `json:"taskId"`
}

// EditCharacterRequest identifies the existing Character prototype to edit.
type EditCharacterRequest struct {
	AssetID         uint   `json:"assetId"`
	AssetResourceID uint   `json:"assetResourceId"`
	Prompt          string `json:"prompt"`
}

type EditCharacterResponse struct {
	TaskID uint `json:"taskId"`
}

// GenerateTileSetItemRequest identifies the TileSet Item to generate.
type GenerateTileSetItemRequest struct {
	AssetID         uint     `json:"assetId"`
	AssetResourceID uint     `json:"assetResourceId"`
	Prompt          string   `json:"prompt"`
	ReferenceURLs   []string `json:"referenceUrls,omitempty"`
}

type GenerateTileSetItemResponse struct {
	TaskID uint `json:"taskId"`
}

// EditTileSetItemRequest identifies the existing TileSet Item to edit.
type EditTileSetItemRequest struct {
	AssetID         uint   `json:"assetId"`
	AssetResourceID uint   `json:"assetResourceId"`
	Prompt          string `json:"prompt"`
}

type EditTileSetItemResponse struct {
	TaskID uint `json:"taskId"`
}

// EditUIComponentRequest identifies the existing UI component to edit.
type EditUIComponentRequest struct {
	AssetID         uint   `json:"assetId"`
	AssetResourceID uint   `json:"assetResourceId"`
	Prompt          string `json:"prompt"`
}

type EditUIComponentResponse struct {
	TaskID uint `json:"taskId"`
}

// EditFrameRequest identifies the existing animation frame to edit.
type EditFrameRequest struct {
	AssetID         uint   `json:"assetId"`
	AssetResourceID uint   `json:"assetResourceId"`
	Prompt          string `json:"prompt"`
}

type EditFrameResponse struct {
	TaskID uint `json:"taskId"`
}

// GenerateObjectRequest identifies the Object prototype to generate.
type GenerateObjectRequest struct {
	AssetID         uint     `json:"assetId"`
	AssetResourceID uint     `json:"assetResourceId"`
	Prompt          string   `json:"prompt"`
	ReferenceURLs   []string `json:"referenceUrls,omitempty"`
}

type GenerateObjectResponse struct {
	TaskID uint `json:"taskId"`
}

// EditObjectRequest identifies the existing Object prototype to edit.
type EditObjectRequest struct {
	AssetID         uint   `json:"assetId"`
	AssetResourceID uint   `json:"assetResourceId"`
	Prompt          string `json:"prompt"`
}

type EditObjectResponse struct {
	TaskID uint `json:"taskId"`
}

// GenerateProjectPreviewRequest identifies the Project used to build the preview context.
type GenerateProjectPreviewRequest struct {
	ProjectID uint   `json:"projectId"`
	Prompt    string `json:"prompt"`
}

type GenerateProjectPreviewResponse struct {
	TaskID uint `json:"taskId"`
}

// GenerateSceneryLayerRequest identifies the Scenery layer to generate.
type GenerateSceneryLayerRequest struct {
	AssetID         uint     `json:"assetId"`
	AssetResourceID uint     `json:"assetResourceId"`
	Prompt          string   `json:"prompt"`
	ReferenceURLs   []string `json:"referenceUrls,omitempty"`
}

type GenerateSceneryLayerResponse struct {
	TaskID uint `json:"taskId"`
}

// EditSceneryLayerRequest identifies the existing Scenery layer to edit.
type EditSceneryLayerRequest struct {
	AssetID         uint   `json:"assetId"`
	AssetResourceID uint   `json:"assetResourceId"`
	Prompt          string `json:"prompt"`
}

type EditSceneryLayerResponse struct {
	TaskID uint `json:"taskId"`
}

// GenerateAnimationRequest identifies the Animation and its first frame resource.
type GenerateAnimationRequest struct {
	AssetID                   uint   `json:"assetId"`
	AssetResourceID           uint   `json:"assetResourceId"`
	FirstFrameAssetResourceID uint   `json:"firstFrameAssetResourceId"`
	Prompt                    string `json:"prompt"`
	FrameCount                uint   `json:"frameCount"`
	KeepFirstFrame            bool   `json:"keepFirstFrame"`
}

type GenerateAnimationResponse struct {
	TaskID uint `json:"taskId"`
}

// GenerateUIRequest identifies the UI Asset and output resource to generate.
type GenerateUIRequest struct {
	AssetID         uint     `json:"assetId"`
	AssetResourceID uint     `json:"assetResourceId"`
	Prompt          string   `json:"prompt"`
	ReferenceURLs   []string `json:"referenceUrls,omitempty"`
}

type GenerateUIResponse struct {
	TaskID uint `json:"taskId"`
}
