package repository_test

import (
	"context"
	"encoding/json"
	"testing"

	"gorm.io/datatypes"

	domain "github.com/1024XEngineer/Holonic-Asset/internal/model/asset"
	"github.com/1024XEngineer/Holonic-Asset/internal/repository"
	"github.com/1024XEngineer/Holonic-Asset/internal/repository/dao"
)

type jsonAssetDaoStub struct {
	dao.AssetDao
	asset       dao.Asset
	created     dao.Asset
	updatedID   uint
	updatedBody domain.AssetContent
}

func (s *jsonAssetDaoStub) GetAssetDetail(_ context.Context, _ uint) (dao.Asset, error) {
	return s.asset, nil
}

func (s *jsonAssetDaoStub) CreateAsset(_ context.Context, asset *dao.Asset) (dao.Asset, error) {
	s.created = *asset
	s.created.ID = 23
	return s.created, nil
}

func (s *jsonAssetDaoStub) UpdateContent(_ context.Context, id uint, content json.RawMessage) error {
	s.updatedID = id
	updated, err := (&domain.Asset{Content: content}).DecodeContent()
	if err != nil {
		return err
	}
	s.updatedBody = updated
	return nil
}

func TestAssetRepositoryReadsAssetContent(t *testing.T) {
	upURL := "https://cdn.example/prototype-up.png"
	downURL := "https://cdn.example/prototype-down.png"
	leftURL := "https://cdn.example/prototype-left.png"
	rightURL := "https://cdn.example/prototype-right.png"
	content := domain.NewAssetContent(domain.AssetTypeCharacter)
	content.ViewElements = []string{"up", "down", "left", "right"}
	content.Prototype.Status = domain.ContentStatusCompleted
	content.Prototype.Directions = map[string]domain.PrototypeDirection{
		"up": {
			Status: domain.ContentStatusCompleted,
			Image:  &domain.ImageResource{URL: &upURL, Status: domain.ContentStatusCompleted},
		},
		"down": {
			Status: domain.ContentStatusCompleted,
			Image:  &domain.ImageResource{URL: &downURL, Status: domain.ContentStatusCompleted},
		},
		"left": {
			Status: domain.ContentStatusCompleted,
			Image:  &domain.ImageResource{URL: &leftURL, Status: domain.ContentStatusCompleted},
		},
		"right": {
			Status: domain.ContentStatusCompleted,
			Image:  &domain.ImageResource{URL: &rightURL, Status: domain.ContentStatusCompleted},
		},
	}
	payload, err := domain.EncodeContent(content)
	if err != nil {
		t.Fatalf("encode content: %v", err)
	}

	repo := &repository.AssetRepositoryImpl{AssetDao: &jsonAssetDaoStub{asset: dao.Asset{
		ID:      7,
		Version: 2,
		Type:    string(domain.AssetTypeCharacter),
		Content: datatypes.JSON(payload),
	}}}
	asset, err := repo.GetAssetDetail(context.Background(), 7)
	if err != nil {
		t.Fatalf("get asset: %v", err)
	}
	decoded, err := asset.DecodeContent()
	if err != nil {
		t.Fatalf("decode asset: %v", err)
	}
	if decoded.Prototype == nil || decoded.Prototype.Status != domain.ContentStatusCompleted || len(decoded.Prototype.Directions) != 4 || decoded.Prototype.Directions["up"].Image == nil || decoded.Prototype.Directions["up"].Image.URL == nil || *decoded.Prototype.Directions["up"].Image.URL != "https://cdn.example/prototype-up.png" {
		t.Fatalf("unexpected asset content: %+v", decoded)
	}
}

func TestAssetRepositoryReturnsGeneratingAssetContent(t *testing.T) {
	content := domain.NewAssetContent(domain.AssetTypeCharacter)
	content.Prototype.Status = domain.ContentStatusProcessing
	payload, err := domain.EncodeContent(content)
	if err != nil {
		t.Fatalf("encode content: %v", err)
	}

	repo := &repository.AssetRepositoryImpl{AssetDao: &jsonAssetDaoStub{asset: dao.Asset{
		ID:      7,
		Version: 1,
		Type:    string(domain.AssetTypeCharacter),
		Content: datatypes.JSON(payload),
	}}}
	asset, err := repo.GetAssetDetail(context.Background(), 7)
	if err != nil {
		t.Fatalf("get asset: %v", err)
	}
	decoded, err := asset.DecodeContent()
	if err != nil {
		t.Fatalf("decode asset: %v", err)
	}
	if decoded.Prototype == nil || decoded.Prototype.Status != domain.ContentStatusProcessing {
		t.Fatalf("unexpected generating content: %+v", decoded)
	}
}

func TestAssetRepositoryCreatesCharacterWithPendingPrototype(t *testing.T) {
	daoStub := &jsonAssetDaoStub{}
	repo := &repository.AssetRepositoryImpl{AssetDao: daoStub}

	created, err := repo.CreateCharacterAsset(context.Background(), &domain.Asset{
		Name:      "hero",
		ProjectID: 42,
	})
	if err != nil {
		t.Fatalf("create character asset: %v", err)
	}
	if created == nil || created.ID != 23 {
		t.Fatalf("expected created asset ID 23, got %+v", created)
	}
	content, err := (&domain.Asset{Content: []byte(daoStub.created.Content), Type: domain.AssetTypeCharacter}).DecodeContent()
	if err != nil {
		t.Fatalf("decode created content: %v", err)
	}
	if content.Prototype == nil || content.Prototype.Status != domain.ContentStatusPending {
		t.Fatalf("expected pending prototype: %+v", content)
	}
}

func TestAssetRepositoryUpdatesDynamicAnimationDirection(t *testing.T) {
	content := domain.NewAssetContent(domain.AssetTypeCharacter)
	content.Animations = []domain.Animation{{
		ID:         9,
		Name:       "walk",
		Status:     domain.ContentStatusPending,
		Directions: map[string]domain.AnimationDirection{"left": {Status: domain.ContentStatusPending}},
	}}
	payload, err := domain.EncodeContent(content)
	if err != nil {
		t.Fatalf("encode content: %v", err)
	}
	daoStub := &jsonAssetDaoStub{asset: dao.Asset{
		ID:      7,
		Type:    string(domain.AssetTypeCharacter),
		Content: datatypes.JSON(payload),
	}}
	repo := &repository.AssetRepositoryImpl{AssetDao: daoStub}

	err = repo.UpdateAnimationDirection(
		context.Background(),
		7,
		9,
		"right",
		domain.ContentStatusCompleted,
		[]domain.Frame{{Status: domain.ContentStatusCompleted}},
	)
	if err != nil {
		t.Fatalf("update animation direction: %v", err)
	}
	if daoStub.updatedID != 7 {
		t.Fatalf("unexpected content update: id=%d", daoStub.updatedID)
	}
	animation := daoStub.updatedBody.Animations[0]
	if animation.Status != domain.ContentStatusProcessing || animation.Directions["right"].Status != domain.ContentStatusCompleted {
		t.Fatalf("unexpected animation state: %+v", animation)
	}
}

func TestAssetRepositoryCreatesAnimationInsideAssetContent(t *testing.T) {
	content := domain.NewAssetContent(domain.AssetTypeCharacter)
	payload, err := domain.EncodeContent(content)
	if err != nil {
		t.Fatalf("encode content: %v", err)
	}
	daoStub := &jsonAssetDaoStub{asset: dao.Asset{
		ID:      7,
		Type:    string(domain.AssetTypeCharacter),
		Content: datatypes.JSON(payload),
	}}
	repo := &repository.AssetRepositoryImpl{AssetDao: daoStub}

	animationID, err := repo.CreateAnimation(context.Background(), 7, "walk")
	if err != nil {
		t.Fatalf("create animation: %v", err)
	}
	if animationID != 1 {
		t.Fatalf("expected animation ID 1, got %d", animationID)
	}
	if len(daoStub.updatedBody.Animations) != 1 || daoStub.updatedBody.Animations[0].Name != "walk" || daoStub.updatedBody.Animations[0].Status != domain.ContentStatusPending {
		t.Fatalf("unexpected animation content: %+v", daoStub.updatedBody.Animations)
	}
}

func TestAssetRepositoryUpdatesOnePrototypeImagePerDirection(t *testing.T) {
	content := domain.NewAssetContent(domain.AssetTypeCharacter)
	content.ViewElements = []string{"up", "down", "left", "right"}
	payload, err := domain.EncodeContent(content)
	if err != nil {
		t.Fatalf("encode content: %v", err)
	}
	daoStub := &jsonAssetDaoStub{asset: dao.Asset{
		ID:      7,
		Type:    string(domain.AssetTypeCharacter),
		Content: datatypes.JSON(payload),
	}}
	repo := &repository.AssetRepositoryImpl{AssetDao: daoStub}

	err = repo.UpdatePrototypeImages(context.Background(), 7, map[string]domain.ImageResource{
		"up":    {Status: domain.ContentStatusCompleted},
		"down":  {Status: domain.ContentStatusCompleted},
		"left":  {Status: domain.ContentStatusCompleted},
		"right": {Status: domain.ContentStatusCompleted},
	})
	if err != nil {
		t.Fatalf("update prototype images: %v", err)
	}
	prototype := daoStub.updatedBody.Prototype
	if prototype == nil || prototype.Status != domain.ContentStatusCompleted || len(prototype.Directions) != 4 {
		t.Fatalf("unexpected prototype content: %+v", prototype)
	}
	for _, direction := range content.ViewElements {
		if prototype.Directions[direction].Image == nil {
			t.Fatalf("expected prototype image for direction %q", direction)
		}
	}
}
