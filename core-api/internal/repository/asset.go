package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"gorm.io/datatypes"

	domain "github.com/1024XEngineer/Holonic-Asset/internal/model/asset"
	"github.com/1024XEngineer/Holonic-Asset/internal/repository/dao"
)

type AssetRepository interface {
	GetAssetsByProjectID(ctx context.Context, projectID uint, filter domain.AssetListFilter) ([]domain.Asset, error)
	GetAssetDetail(ctx context.Context, id uint) (*domain.Asset, error)
	UpdateAsset(ctx context.Context, id uint, update *domain.AssetUpdate) (*domain.Asset, error)
	UpdateContent(ctx context.Context, assetID uint, content domain.AssetContent) error
	UpdateAnimationDirection(
		ctx context.Context,
		assetID uint,
		animationID uint,
		direction string,
		status domain.ContentStatus,
		frames []domain.Frame,
	) error

	CreateCharacterAsset(ctx context.Context, asset *domain.Asset) (*domain.Asset, error)
	CreateObjectAsset(ctx context.Context, asset *domain.Asset) (uint, error)
	CreateTileSetAsset(ctx context.Context, asset *domain.Asset) (uint, error)
	CreateUIAsset(ctx context.Context, asset *domain.Asset) (uint, error)
	CreateSceneryAsset(ctx context.Context, asset *domain.Asset) (uint, error)
	CreateAnimation(ctx context.Context, assetID uint, name string) (uint, error)
	UpdatePrototypeImages(ctx context.Context, assetID uint, images map[string]domain.ImageResource) error

	CreateRecord(ctx context.Context, version *domain.AssetVersion) (*domain.AssetVersion, error)
	RollBackRecord(ctx context.Context, assetID uint, version uint) (uint, error)
	Copy(ctx context.Context, assetID uint, version uint) (uint, error)
}

type AssetRepositoryImpl struct {
	AssetDao   dao.AssetDao
	VersionDao dao.AssetVersionDao
}

func NewAssetRepository(assetDao dao.AssetDao) AssetRepository {
	return &AssetRepositoryImpl{AssetDao: assetDao}
}

func (r *AssetRepositoryImpl) GetAssetsByProjectID(ctx context.Context, projectID uint, filter domain.AssetListFilter) ([]domain.Asset, error) {
	assets, err := r.AssetDao.GetAssetsByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Asset, 0, len(assets))
	for index := range assets {
		asset := convertAssetToDomain(assets[index])
		if !matchesAssetFilter(asset, filter) {
			continue
		}
		result = append(result, asset)
	}
	return result, nil
}

func matchesAssetFilter(asset domain.Asset, filter domain.AssetListFilter) bool {
	query := strings.TrimSpace(strings.ToLower(filter.Query))
	if query != "" && !strings.Contains(strings.ToLower(asset.Name), query) && !strings.Contains(strings.ToLower(asset.Description), query) {
		return false
	}
	if len(filter.Types) > 0 && !containsAssetType(filter.Types, asset.Type) {
		return false
	}
	for _, tag := range filter.Tags {
		if !containsTag(asset.Tags, tag) {
			return false
		}
	}
	return true
}

func containsAssetType(types []domain.AssetType, target domain.AssetType) bool {
	return slices.Contains(types, target)
}

func containsTag(tags []string, target string) bool {
	return slices.Contains(tags, target)
}

func (r *AssetRepositoryImpl) GetAssetDetail(ctx context.Context, id uint) (*domain.Asset, error) {
	asset, err := r.AssetDao.GetAssetDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	result := convertAssetToDomain(asset)
	return &result, nil
}

func (r *AssetRepositoryImpl) UpdateAsset(
	ctx context.Context,
	id uint,
	update *domain.AssetUpdate,
) (*domain.Asset, error) {
	if update == nil {
		return nil, fmt.Errorf("repository: asset update is nil")
	}

	value := &dao.AssetUpdate{
		Name:        update.Name,
		ProjectID:   update.ProjectID,
		Description: update.Description,
		Version:     update.Version,
	}
	if update.Type != nil {
		assetType := string(*update.Type)
		value.Type = &assetType
	}
	if update.Tags != nil {
		value.Tags = update.Tags
	}
	if update.Attributes != nil {
		value.Attributes = update.Attributes
	}

	asset, err := r.AssetDao.UpdateAsset(ctx, id, value)
	if err != nil {
		return nil, err
	}
	result := convertAssetToDomain(asset)
	return &result, nil
}

func convertAssetToDomain(asset dao.Asset) domain.Asset {
	return domain.Asset{
		ID:          asset.ID,
		Name:        asset.Name,
		ProjectID:   asset.ProjectID,
		Type:        domain.AssetType(asset.Type),
		Description: asset.Description,
		Tags:        append([]string(nil), asset.Tags...),
		Attributes:  append([]byte(nil), asset.Attributes...),
		Content:     append([]byte(nil), asset.Content...),
		Version:     asset.Version,
	}
}

func convertAssetToDAO(asset *domain.Asset, assetType domain.AssetType) (*dao.Asset, error) {
	if asset == nil {
		asset = &domain.Asset{}
	}
	content, err := asset.DecodeContent()
	if err != nil {
		return nil, fmt.Errorf("repository: decode asset content: %w", err)
	}
	if (assetType == domain.AssetTypeCharacter || assetType == domain.AssetTypeObject) && content.Prototype == nil {
		content.Prototype = &domain.Prototype{Status: domain.ContentStatusPending}
	}
	encoded, err := domain.EncodeContent(content)
	if err != nil {
		return nil, err
	}

	return &dao.Asset{
		ID:          asset.ID,
		Name:        asset.Name,
		ProjectID:   asset.ProjectID,
		Type:        string(assetType),
		Description: asset.Description,
		Tags:        append([]string(nil), asset.Tags...),
		Attributes:  append([]byte(nil), asset.Attributes...),
		Content:     datatypes.JSON(encoded),
		Version:     asset.Version,
	}, nil
}

func (r *AssetRepositoryImpl) UpdateContent(
	ctx context.Context,
	assetID uint,
	content domain.AssetContent,
) error {
	encoded, err := domain.EncodeContent(content)
	if err != nil {
		return fmt.Errorf("repository: encode asset %d content: %w", assetID, err)
	}
	return r.AssetDao.UpdateContent(ctx, assetID, encoded)
}

func (r *AssetRepositoryImpl) UpdateAnimationDirection(
	ctx context.Context,
	assetID uint,
	animationID uint,
	direction string,
	status domain.ContentStatus,
	frames []domain.Frame,
) error {
	if direction == "" {
		return fmt.Errorf("repository: animation direction is empty")
	}
	asset, err := r.GetAssetDetail(ctx, assetID)
	if err != nil {
		return err
	}
	content, err := asset.DecodeContent()
	if err != nil {
		return fmt.Errorf("repository: decode asset %d content: %w", assetID, err)
	}
	for index := range content.Animations {
		animation := &content.Animations[index]
		if animation.ID != animationID {
			continue
		}
		if animation.Directions == nil {
			animation.Directions = make(map[string]domain.AnimationDirection)
		}
		current := animation.Directions[direction]
		current.Status = status
		current.Frames = append([]domain.Frame(nil), frames...)
		animation.Directions[direction] = current
		animation.Status = aggregateAnimationStatus(animation.Directions)
		return r.UpdateContent(ctx, assetID, content)
	}
	return fmt.Errorf("repository: animation %d not found in asset %d", animationID, assetID)
}

func (r *AssetRepositoryImpl) createAsset(ctx context.Context, asset *domain.Asset, assetType domain.AssetType) (uint, error) {
	created, err := r.createAssetResult(ctx, asset, assetType)
	if err != nil {
		return 0, err
	}
	return created.ID, nil
}

func (r *AssetRepositoryImpl) createAssetResult(ctx context.Context, asset *domain.Asset, assetType domain.AssetType) (*domain.Asset, error) {
	value, err := convertAssetToDAO(asset, assetType)
	if err != nil {
		return nil, err
	}
	created, err := r.AssetDao.CreateAsset(ctx, value)
	if err != nil {
		return nil, err
	}
	result := convertAssetToDomain(created)
	return &result, nil
}

func (r *AssetRepositoryImpl) CreateCharacterAsset(ctx context.Context, asset *domain.Asset) (*domain.Asset, error) {
	return r.createAssetResult(ctx, asset, domain.AssetTypeCharacter)
}

func (r *AssetRepositoryImpl) CreateObjectAsset(ctx context.Context, asset *domain.Asset) (uint, error) {
	return r.createAsset(ctx, asset, domain.AssetTypeObject)
}

func (r *AssetRepositoryImpl) CreateTileSetAsset(ctx context.Context, asset *domain.Asset) (uint, error) {
	return r.createAsset(ctx, asset, domain.AssetTypeTileSet)
}

func (r *AssetRepositoryImpl) CreateUIAsset(ctx context.Context, asset *domain.Asset) (uint, error) {
	return r.createAsset(ctx, asset, domain.AssetTypeUI)
}

func (r *AssetRepositoryImpl) CreateSceneryAsset(ctx context.Context, asset *domain.Asset) (uint, error) {
	return r.createAsset(ctx, asset, domain.AssetTypeScenery)
}

func (r *AssetRepositoryImpl) CreateAnimation(ctx context.Context, assetID uint, name string) (uint, error) {
	if name == "" {
		return 0, fmt.Errorf("repository: animation name is empty")
	}
	asset, err := r.GetAssetDetail(ctx, assetID)
	if err != nil {
		return 0, err
	}
	content, err := asset.DecodeContent()
	if err != nil {
		return 0, err
	}
	animationID := nextAnimationID(content.Animations)
	content.Animations = append(content.Animations, domain.Animation{
		ID:         animationID,
		Name:       name,
		Status:     domain.ContentStatusPending,
		Directions: make(map[string]domain.AnimationDirection),
	})
	if err := r.UpdateContent(ctx, asset.ID, content); err != nil {
		return 0, err
	}
	return animationID, nil
}

func (r *AssetRepositoryImpl) UpdatePrototypeImages(
	ctx context.Context,
	assetID uint,
	images map[string]domain.ImageResource,
) error {
	asset, err := r.GetAssetDetail(ctx, assetID)
	if err != nil {
		return err
	}
	content, err := asset.DecodeContent()
	if err != nil {
		return err
	}
	if content.Prototype == nil {
		content.Prototype = &domain.Prototype{Status: domain.ContentStatusPending}
	}
	if content.Prototype.Directions == nil {
		content.Prototype.Directions = make(map[string]domain.PrototypeDirection)
	}
	for _, direction := range content.ViewElements {
		if _, ok := content.Prototype.Directions[direction]; !ok {
			content.Prototype.Directions[direction] = domain.PrototypeDirection{Status: domain.ContentStatusPending}
		}
	}
	for direction, image := range images {
		if direction == "" {
			return fmt.Errorf("repository: prototype direction is empty")
		}
		if len(content.ViewElements) > 0 && !containsViewElement(content.ViewElements, direction) {
			return fmt.Errorf("repository: prototype direction %q is not supported by asset %d", direction, assetID)
		}
		current := content.Prototype.Directions[direction]
		current.Status = image.Status
		current.Image = &image
		content.Prototype.Directions[direction] = current
	}
	content.Prototype.Status = aggregatePrototypeStatus(content.ViewElements, content.Prototype.Directions)
	return r.UpdateContent(ctx, asset.ID, content)
}

func (r *AssetRepositoryImpl) CreateRecord(ctx context.Context, version *domain.AssetVersion) (*domain.AssetVersion, error) {
	if version == nil {
		return nil, fmt.Errorf("repository: asset version is nil")
	}
	asset, err := r.GetAssetDetail(ctx, version.AssetID)
	if err != nil {
		return nil, err
	}
	if version.Version == 0 {
		version.Version = asset.Version + 1
	}
	snapshot := &domain.AssetVersion{
		AssetID: version.AssetID,
		Version: version.Version,
		Content: append([]byte(nil), asset.Content...),
	}
	if r.VersionDao != nil {
		id, err := r.VersionDao.CreateAssetVersion(ctx, &dao.AssetVersion{
			AssetID: snapshot.AssetID,
			Version: snapshot.Version,
			Content: datatypes.JSON(snapshot.Content),
		})
		if err != nil {
			return nil, err
		}
		snapshot.ID = id
	}
	if err := r.AssetDao.UpdateAssetVersion(ctx, asset.ID, snapshot.Version); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (r *AssetRepositoryImpl) RollBackRecord(ctx context.Context, assetID uint, version uint) (uint, error) {
	if r.VersionDao == nil {
		return 0, fmt.Errorf("repository: version storage is not configured")
	}
	versions, err := r.VersionDao.GetAssetVersionsByAssetID(ctx, assetID)
	if err != nil {
		return 0, err
	}
	for _, candidate := range versions {
		if candidate.Version != version {
			continue
		}
		content, err := decodeDAOContent(candidate.Content)
		if err != nil {
			return 0, err
		}
		if err := r.UpdateContent(ctx, assetID, content); err != nil {
			return 0, err
		}
		if err := r.AssetDao.UpdateAssetVersion(ctx, assetID, version); err != nil {
			return 0, err
		}
		return version, nil
	}
	return 0, fmt.Errorf("repository: asset %d version %d not found", assetID, version)
}

func (r *AssetRepositoryImpl) Copy(ctx context.Context, assetID uint, version uint) (uint, error) {
	original, err := r.GetAssetDetail(ctx, assetID)
	if err != nil {
		return 0, err
	}
	copyAsset := *original
	copyAsset.ID = 0
	if version > 0 {
		copyAsset.Version = version
	}
	created, err := convertAssetToDAO(&copyAsset, original.Type)
	if err != nil {
		return 0, err
	}
	created.ID = 0
	newAsset, err := r.AssetDao.CreateAsset(ctx, created)
	if err != nil {
		return 0, err
	}
	return newAsset.ID, nil
}

func decodeDAOContent(content []byte) (domain.AssetContent, error) {
	var result domain.AssetContent
	if err := json.Unmarshal(content, &result); err != nil {
		return domain.AssetContent{}, err
	}
	return result, nil
}

func nextAnimationID(animations []domain.Animation) uint {
	var id uint
	for _, animation := range animations {
		if animation.ID > id {
			id = animation.ID
		}
	}
	return id + 1
}

func containsViewElement(viewElements []string, direction string) bool {
	return slices.Contains(viewElements, direction)
}

func aggregatePrototypeStatus(
	viewElements []string,
	directions map[string]domain.PrototypeDirection,
) domain.ContentStatus {
	if len(directions) == 0 {
		return domain.ContentStatusPending
	}
	targets := viewElements
	if len(targets) == 0 {
		targets = make([]string, 0, len(directions))
		for direction := range directions {
			targets = append(targets, direction)
		}
	}
	completed := 0
	failed := 0
	for _, direction := range targets {
		value, ok := directions[direction]
		if !ok {
			continue
		}
		switch value.Status {
		case domain.ContentStatusCompleted:
			if value.Image != nil {
				completed++
			}
		case domain.ContentStatusFailed:
			failed++
		}
	}
	if completed == len(targets) {
		return domain.ContentStatusCompleted
	}
	if failed > 0 && completed > 0 {
		return domain.ContentStatusPartial
	}
	if failed == len(targets) {
		return domain.ContentStatusFailed
	}
	return domain.ContentStatusProcessing
}

func aggregateAnimationStatus(directions map[string]domain.AnimationDirection) domain.ContentStatus {
	if len(directions) == 0 {
		return domain.ContentStatusPending
	}
	completed := 0
	failed := 0
	for _, direction := range directions {
		switch direction.Status {
		case domain.ContentStatusFailed:
			failed++
		case domain.ContentStatusCompleted:
			completed++
		}
	}
	if completed == len(directions) {
		return domain.ContentStatusCompleted
	}
	if failed > 0 && completed > 0 {
		return domain.ContentStatusPartial
	}
	if failed == len(directions) {
		return domain.ContentStatusFailed
	}
	return domain.ContentStatusProcessing
}
