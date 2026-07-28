package dto

import (
	"encoding/json"

	domain "github.com/1024XEngineer/Holonic-Asset/internal/model/asset"
)

type AssetListItemResponse struct {
	AssetID     uint             `json:"assetId"`
	Name        string           `json:"name"`
	ProjectID   uint             `json:"projectId"`
	Type        domain.AssetType `json:"type"`
	Description string           `json:"description"`
	Tags        []string         `json:"tags"`
	Version     uint             `json:"version"`
}

type GetAssetsResponse struct {
	Assets []AssetListItemResponse `json:"assets"`
}

type GetAssetsRequest struct {
	Query string             `query:"query"`
	Tags  []string           `query:"tags"`
	Types []domain.AssetType `query:"types"`
}

type AssetDetailResponse struct {
	AssetID     uint             `json:"assetId"`
	Name        string           `json:"name"`
	ProjectID   uint             `json:"projectId"`
	Type        domain.AssetType `json:"type"`
	Description string           `json:"description"`
	Tags        []string         `json:"tags"`
	Attributes  json.RawMessage  `json:"attributes"`
	Content     json.RawMessage  `json:"content,omitempty"`
	Version     uint             `json:"version"`
}

type RecordAssetRequest struct {
	AssetID uint
}

type RecordAssetResponse struct {
}

type CopyAssetRequest struct {
	AssetID uint
}

type CopyAssetResponse struct {
	NewAssetID uint `json:"newAssetId"`
}

type UpdateAssetRequest struct {
	AssetID     uint              `json:"assetId"`
	Name        *string           `json:"name,omitempty"`
	ProjectID   *uint             `json:"projectId,omitempty"`
	Type        *domain.AssetType `json:"type,omitempty"`
	Description *string           `json:"description,omitempty"`
	Tags        *[]string         `json:"tags,omitempty"`
	Attributes  *json.RawMessage  `json:"attributes,omitempty"`
	Version     *uint             `json:"version,omitempty"`
}

type UpdateAssetResponse struct {
	AssetID     uint             `json:"assetId"`
	Name        string           `json:"name"`
	ProjectID   uint             `json:"projectId"`
	Type        domain.AssetType `json:"type"`
	Description string           `json:"description"`
	Tags        []string         `json:"tags"`
	Attributes  json.RawMessage  `json:"attributes"`
	Version     uint             `json:"version"`
}

type RollBackAssetRequest struct {
	AssetID uint
	Version uint
}

type RollBackAssetResponse struct {
	Asset *domain.Asset `json:"asset"`
}
