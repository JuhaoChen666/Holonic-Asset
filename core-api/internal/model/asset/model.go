package asset

import (
	"encoding/json"
)

type Asset struct {
	ID          uint
	Name        string
	ProjectID   uint
	Type        AssetType
	Description string
	Tags        []string        `json:"tags"`
	Attributes  json.RawMessage `json:"attributes,omitempty"`
	Content     json.RawMessage `json:"content,omitempty"`
	Version     uint
}

type AssetListFilter struct {
	Query string
	Tags  []string
	Types []AssetType
}

type AssetUpdate struct {
	Name        *string
	ProjectID   *uint
	Type        *AssetType
	Description *string
	Tags        *[]string
	Attributes  *json.RawMessage
	Version     *uint
}

type ContentStatus string
type ViewMode string

const (
	ContentStatusPending    ContentStatus = "pending"
	ContentStatusProcessing ContentStatus = "processing"
	ContentStatusPartial    ContentStatus = "partial"
	ContentStatusCompleted  ContentStatus = "completed"
	ContentStatusFailed     ContentStatus = "failed"
	ContentStatusCancelled  ContentStatus = "cancelled"
)

const (
	ViewModeSideOn  ViewMode = "side_on"
	ViewModeTopDown ViewMode = "top_down"
)

type AssetContent struct {
	ViewMode     ViewMode       `json:"viewMode,omitempty"`
	ViewElements []string       `json:"viewElements,omitempty"`
	Prototype    *Prototype     `json:"prototype,omitempty"`
	Animations   []Animation    `json:"animations,omitempty"`
	TileSize     *TileSize      `json:"tileSize,omitempty"`
	Items        []TileSetItem  `json:"items,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

type Prototype struct {
	Status     ContentStatus                 `json:"status"`
	Directions map[string]PrototypeDirection `json:"directions,omitempty"`
}

type PrototypeDirection struct {
	Status ContentStatus  `json:"status"`
	Image  *ImageResource `json:"image,omitempty"`
}

type Animation struct {
	ID         uint                          `json:"id"`
	Name       string                        `json:"name"`
	Status     ContentStatus                 `json:"status"`
	Directions map[string]AnimationDirection `json:"directions,omitempty"`
}

type AnimationDirection struct {
	Status ContentStatus `json:"status"`
	Frames []Frame       `json:"frames,omitempty"`
}

type ImageResource struct {
	URL      *string         `json:"url,omitempty"`
	Status   ContentStatus   `json:"status"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

type Frame struct {
	URL      *string         `json:"url,omitempty"`
	Status   ContentStatus   `json:"status"`
	Duration uint            `json:"duration,omitempty"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

type TileSize struct {
	Width  uint `json:"width"`
	Height uint `json:"height"`
}

type TileSetItem struct {
	Name   string        `json:"name"`
	Status ContentStatus `json:"status"`
	Tiles  []Tile        `json:"tiles,omitempty"`
}

type Tile struct {
	Name     string          `json:"name"`
	URL      *string         `json:"url,omitempty"`
	Status   ContentStatus   `json:"status"`
	Position TilePosition    `json:"position"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

type TilePosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func NewAssetContent(assetType AssetType) AssetContent {
	content := AssetContent{}
	if assetType == AssetTypeCharacter || assetType == AssetTypeObject {
		content.Prototype = &Prototype{Status: ContentStatusPending}
	}
	return content
}

func (a Asset) DecodeContent() (AssetContent, error) {
	if len(a.Content) == 0 {
		return NewAssetContent(a.Type), nil
	}

	var content AssetContent
	if err := json.Unmarshal(a.Content, &content); err != nil {
		return AssetContent{}, err
	}
	return content, nil
}

func EncodeContent(content AssetContent) (json.RawMessage, error) {
	content.normalizePrototypeDirections()
	value, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(value), nil
}

func (content *AssetContent) normalizePrototypeDirections() {
	if content.Prototype == nil || len(content.ViewElements) == 0 {
		return
	}
	if content.Prototype.Directions == nil {
		content.Prototype.Directions = make(map[string]PrototypeDirection, len(content.ViewElements))
	}
	supported := make(map[string]struct{}, len(content.ViewElements))
	for _, direction := range content.ViewElements {
		if direction == "" {
			continue
		}
		supported[direction] = struct{}{}
		if _, ok := content.Prototype.Directions[direction]; !ok {
			content.Prototype.Directions[direction] = PrototypeDirection{Status: ContentStatusPending}
		}
	}
	for direction := range content.Prototype.Directions {
		if _, ok := supported[direction]; !ok {
			delete(content.Prototype.Directions, direction)
		}
	}
}

type AssetVersion struct {
	ID        uint
	AssetID   uint
	Version   uint
	CreatedAt int64
	Content   json.RawMessage
}
