package interfaces

import "context"

// TagService defines tag management and tag assignments for assets.
type TagService interface {
	// CreateTag creates a tag for asset classification.
	CreateTag(ctx context.Context, tag *Tag) error

	// ListTags returns tags used by assets in the specified project.
	ListTags(ctx context.Context, projectID uint) ([]*Tag, error)

	// GetTag returns the details of a tag.
	GetTag(ctx context.Context, tagID uint) (*Tag, error)

	// UpdateTag updates mutable fields of a tag.
	UpdateTag(ctx context.Context, tag *Tag) error

	// DeleteTag deletes a tag and removes its asset assignments.
	DeleteTag(ctx context.Context, tagID uint) error

	// AssignTag assigns a tag to an asset.
	AssignTag(ctx context.Context, assetID uint, tagID uint) error

	// RemoveTag removes a tag assignment from an asset.
	RemoveTag(ctx context.Context, assetID uint, tagID uint) error

	// ListAssetTags returns the tags assigned to an asset.
	ListAssetTags(ctx context.Context, assetID uint) ([]*Tag, error)
}

// AssetAssociationService defines relationships between assets in the same project.
type AssetAssociationService interface {
	// AssociateAssets creates an association between two assets.
	AssociateAssets(ctx context.Context, association *AssetAssociation) error

	// RemoveAssetAssociation removes an existing asset association.
	RemoveAssetAssociation(ctx context.Context, associationID uint) error

	// ListAssetAssociations returns the associations of an asset.
	ListAssetAssociations(ctx context.Context, assetID uint) ([]*AssetAssociation, error)
}

// AssetDiscoveryService defines project-scoped asset search and filtering.
type AssetDiscoveryService interface {
	// SearchAssets searches assets by text and taxonomy criteria.
	SearchAssets(
		ctx context.Context,
		request *SearchAssetsRequest,
	) (*AssetSearchResult, error)

	// FilterAssets filters assets by structured taxonomy criteria.
	FilterAssets(
		ctx context.Context,
		request *FilterAssetsRequest,
	) (*AssetSearchResult, error)
}
