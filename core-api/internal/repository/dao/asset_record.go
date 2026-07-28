package dao

import (
	"context"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AssetVersion struct {
	ID      uint
	AssetID uint
	Version uint
	Content datatypes.JSON `gorm:"type:jsonb"`
}

type AssetVersionDao interface {
	CreateAssetVersion(ctx context.Context, version *AssetVersion) (uint, error)
	CreateAssetVersions(ctx context.Context, version []AssetVersion) error
	DeleteAssetVersion(ctx context.Context, assetID uint, version uint) error
	GetAssetVersionsByAssetID(ctx context.Context, assetID uint) ([]AssetVersion, error)
}

type AssetVersionDaoImpl struct {
	DB *gorm.DB
}

func (a *AssetVersionDaoImpl) CreateAssetVersion(ctx context.Context, version *AssetVersion) (uint, error) {
	if version == nil {
		return 0, fmt.Errorf("dao: asset version is nil")
	}
	if err := a.DB.WithContext(ctx).Create(version).Error; err != nil {
		return 0, fmt.Errorf("dao: create asset version: %w", err)
	}
	return version.ID, nil
}

func (a *AssetVersionDaoImpl) DeleteAssetVersion(ctx context.Context, assetID uint, version uint) error {
	result := a.DB.WithContext(ctx).
		Where("asset_id = ? AND version = ?", assetID, version).
		Delete(&AssetVersion{})
	if result.Error != nil {
		return fmt.Errorf("dao: delete asset version %d/%d: %w", assetID, version, result.Error)
	}
	return nil
}

func (a *AssetVersionDaoImpl) GetAssetVersionsByAssetID(ctx context.Context, assetID uint) ([]AssetVersion, error) {
	versions := make([]AssetVersion, 0)
	err := a.DB.WithContext(ctx).
		Where("asset_id = ?", assetID).
		Order("version ASC").
		Find(&versions).Error
	if err != nil {
		return nil, fmt.Errorf("dao: get asset versions for asset %d: %w", assetID, err)
	}
	return versions, nil
}
