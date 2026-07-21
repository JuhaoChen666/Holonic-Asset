package module

import (
	"context"

	interfaces "../Interface"
)

// TaxonomyModule describes the public tag, association, search, and filtering capabilities.
type TaxonomyModule interface {
	// RegisterAsset provides asset identity, ownership, and project scope.
	RegisterAsset(asset AssetModule)

	// RegisterTagService registers tag management and asset-tag assignment capabilities.
	RegisterTagService(service interfaces.TagService)

	// RegisterAssetAssociationService registers asset association capabilities.
	RegisterAssetAssociationService(service interfaces.AssetAssociationService)

	// RegisterAssetDiscoveryService registers asset search and filtering capabilities.
	RegisterAssetDiscoveryService(service interfaces.AssetDiscoveryService)

	// CreateTag creates a tag for asset classification.
	CreateTag(ctx context.Context, tag *interfaces.Tag) error

	// ListTags returns tags used by assets in the specified project.
	ListTags(ctx context.Context, projectID uint) ([]*interfaces.Tag, error)

	// GetTag returns the details of a tag.
	GetTag(ctx context.Context, tagID uint) (*interfaces.Tag, error)

	// UpdateTag updates mutable fields of a tag.
	UpdateTag(ctx context.Context, tag *interfaces.Tag) error

	// DeleteTag deletes a tag and removes its asset assignments.
	DeleteTag(ctx context.Context, tagID uint) error

	// AssignTag assigns a tag to an asset.
	AssignTag(ctx context.Context, assetID uint, tagID uint) error

	// RemoveTag removes a tag assignment from an asset.
	RemoveTag(ctx context.Context, assetID uint, tagID uint) error

	// ListAssetTags returns the tags assigned to an asset.
	ListAssetTags(ctx context.Context, assetID uint) ([]*interfaces.Tag, error)

	// AssociateAssets creates an association between two assets.
	AssociateAssets(ctx context.Context, association *interfaces.AssetAssociation) error

	// RemoveAssetAssociation removes an existing asset association.
	RemoveAssetAssociation(ctx context.Context, associationID uint) error

	// ListAssetAssociations returns the associations of an asset.
	ListAssetAssociations(ctx context.Context, assetID uint) ([]*interfaces.AssetAssociation, error)

	// SearchAssets searches assets by text and taxonomy criteria.
	SearchAssets(
		ctx context.Context,
		request *interfaces.SearchAssetsRequest,
	) (*interfaces.AssetSearchResult, error)

	// FilterAssets filters assets by structured taxonomy criteria.
	FilterAssets(
		ctx context.Context,
		request *interfaces.FilterAssetsRequest,
	) (*interfaces.AssetSearchResult, error)
}
