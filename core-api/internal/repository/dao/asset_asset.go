package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Asset struct {
	ID          uint `gorm:"primaryKey"`
	Name        string
	ProjectID   uint `gorm:"index"`
	Type        string
	Description string
	Tags        []string        `json:"tags" gorm:"serializer:json"`
	Attributes  json.RawMessage `json:"attributes" gorm:"serializer:json"`
	Content     datatypes.JSON  `json:"content" gorm:"type:jsonb"`
	Version     uint
}

type AssetUpdate struct {
	Name        *string
	ProjectID   *uint
	Type        *string
	Description *string
	Tags        *[]string
	Attributes  *json.RawMessage
	Version     *uint
}

type AssetDao interface {
	CreateAsset(ctx context.Context, asset *Asset) (Asset, error)
	GetAssetsByProjectID(ctx context.Context, projectID uint) ([]Asset, error)
	GetAssetDetail(ctx context.Context, id uint) (Asset, error)
	UpdateAsset(ctx context.Context, id uint, update *AssetUpdate) (Asset, error)
	UpdateAssetVersion(ctx context.Context, id uint, version uint) error
	UpdateContent(ctx context.Context, id uint, content json.RawMessage) error
}

type AssetDaoImpl struct {
	DB *gorm.DB
}

func (a *AssetDaoImpl) GetAssetsByProjectID(ctx context.Context, projectID uint) ([]Asset, error) {
	assets := make([]Asset, 0)
	err := a.DB.WithContext(ctx).
		Where("project_id = ?", projectID).
		Select("id, name, project_id, type, description, tags, version").
		Order("id ASC").
		Find(&assets).Error
	return assets, err
}

func (a *AssetDaoImpl) CreateAsset(ctx context.Context, asset *Asset) (Asset, error) {
	if asset == nil {
		return Asset{}, fmt.Errorf("dao: asset is nil")
	}
	if asset.Version == 0 {
		asset.Version = 1
	}
	if err := a.DB.WithContext(ctx).Create(asset).Error; err != nil {
		return Asset{}, fmt.Errorf("dao: create asset: %w", err)
	}
	return *asset, nil
}

func (a *AssetDaoImpl) GetAssetDetail(ctx context.Context, id uint) (Asset, error) {
	var asset Asset
	err := a.DB.WithContext(ctx).First(&asset, id).Error
	return asset, err
}

func (a *AssetDaoImpl) UpdateAsset(ctx context.Context, id uint, update *AssetUpdate) (Asset, error) {
	if update == nil {
		return Asset{}, fmt.Errorf("dao: asset update is nil")
	}

	values := make(map[string]any)
	if update.Name != nil {
		values["name"] = *update.Name
	}
	if update.ProjectID != nil {
		values["project_id"] = *update.ProjectID
	}
	if update.Type != nil {
		values["type"] = *update.Type
	}
	if update.Description != nil {
		values["description"] = *update.Description
	}
	if update.Tags != nil {
		values["tags"] = *update.Tags
	}
	if update.Attributes != nil {
		values["attributes"] = *update.Attributes
	}
	if update.Version != nil {
		values["version"] = *update.Version
	}

	query := a.DB.WithContext(ctx).Model(&Asset{}).Where("id = ?", id)
	if len(values) > 0 {
		if result := query.Updates(values); result.Error != nil {
			return Asset{}, fmt.Errorf("dao: update asset %d: %w", id, result.Error)
		} else if result.RowsAffected == 0 {
			return Asset{}, fmt.Errorf("dao: asset %d not found", id)
		}
	}

	var asset Asset
	if err := a.DB.WithContext(ctx).First(&asset, id).Error; err != nil {
		return Asset{}, fmt.Errorf("dao: get updated asset %d: %w", id, err)
	}
	return asset, nil
}

func (a *AssetDaoImpl) UpdateAssetVersion(ctx context.Context, id uint, version uint) error {
	result := a.DB.WithContext(ctx).
		Model(&Asset{}).
		Where("id = ?", id).
		Update("version", version)
	if result.Error != nil {
		return fmt.Errorf("dao: update version for asset %d: %w", id, result.Error)
	}
	return nil
}

func (a *AssetDaoImpl) UpdateContent(
	ctx context.Context,
	id uint,
	content json.RawMessage,
) error {
	query := a.DB.WithContext(ctx).
		Model(&Asset{}).
		Where("id = ?", id).
		Update("content", datatypes.JSON(content))
	if query.Error != nil {
		return fmt.Errorf("dao: update content for asset %d: %w", id, query.Error)
	}
	if query.RowsAffected == 0 {
		return fmt.Errorf("dao: asset %d not found", id)
	}
	return nil
}
