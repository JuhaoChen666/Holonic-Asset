package service_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	domain "github.com/1024XEngineer/Holonic-Asset/internal/model/generation"
	"github.com/1024XEngineer/Holonic-Asset/internal/repository"
	"github.com/1024XEngineer/Holonic-Asset/internal/service"
)

type generationReaderStub struct {
	filter *domain.RunListFilter
	page   *domain.RunListPage
}

func (*generationReaderStub) GetRun(context.Context, domain.RunID) (*domain.GenerationRun, error) {
	return &domain.GenerationRun{}, nil
}

func (s *generationReaderStub) ListRuns(
	_ context.Context,
	filter *domain.RunListFilter,
) (*domain.RunListPage, error) {
	s.filter = filter
	return s.page, nil
}

func (*generationReaderStub) GetStep(context.Context, domain.StepID) (*domain.Step, error) {
	return &domain.Step{}, nil
}

func (*generationReaderStub) ListSteps(context.Context, domain.RunID) ([]domain.Step, error) {
	return nil, nil
}

var _ repository.Reader = (*generationReaderStub)(nil)

func TestListBuildsProjectScopeFilter(t *testing.T) {
	reader := &generationReaderStub{page: &domain.RunListPage{}}
	generationService := service.NewGenerationService(reader, nil, nil)

	_, err := generationService.List(context.Background(), &domain.RunListQuery{
		ProjectID: 42,
		Status:    domain.RunListStatusActive,
		Limit:     10,
		Cursor:    "cursor",
	})
	if err != nil {
		t.Fatalf("list generation runs: %v", err)
	}

	if reader.filter == nil {
		t.Fatal("expected repository filter")
	}
	if reader.filter.ProjectID != 42 || reader.filter.AssetID != nil ||
		reader.filter.Limit != 10 || reader.filter.Cursor != "cursor" {
		t.Fatalf("unexpected filter: %+v", reader.filter)
	}
	if !reflect.DeepEqual(reader.filter.Lifecycles, domain.ActiveRunLifecycles()) {
		t.Fatalf("unexpected lifecycles: %v", reader.filter.Lifecycles)
	}
	if !reflect.DeepEqual(reader.filter.IncludeTaskTypes, domain.ProjectLevelTaskTypes()) {
		t.Fatalf("unexpected project task types: %v", reader.filter.IncludeTaskTypes)
	}
	if len(reader.filter.ExcludeTaskTypes) != 0 {
		t.Fatalf("project filter must not exclude task types: %v", reader.filter.ExcludeTaskTypes)
	}
}

func TestListBuildsAssetScopeFilter(t *testing.T) {
	assetID := uint(9)
	reader := &generationReaderStub{page: &domain.RunListPage{}}
	generationService := service.NewGenerationService(reader, nil, nil)

	_, err := generationService.List(context.Background(), &domain.RunListQuery{
		ProjectID: 42,
		AssetID:   &assetID,
	})
	if err != nil {
		t.Fatalf("list generation runs: %v", err)
	}

	if reader.filter == nil || reader.filter.AssetID == nil || *reader.filter.AssetID != assetID {
		t.Fatalf("unexpected asset filter: %+v", reader.filter)
	}
	if len(reader.filter.IncludeTaskTypes) != 0 {
		t.Fatalf("asset filter must not include project task types: %v", reader.filter.IncludeTaskTypes)
	}
	if !reflect.DeepEqual(reader.filter.ExcludeTaskTypes, domain.ProjectLevelTaskTypes()) {
		t.Fatalf("unexpected excluded task types: %v", reader.filter.ExcludeTaskTypes)
	}
	if reader.filter.Limit != 20 {
		t.Fatalf("expected default limit 20, got %d", reader.filter.Limit)
	}
}

func TestListRejectsUnsupportedStatus(t *testing.T) {
	generationService := service.NewGenerationService(&generationReaderStub{}, nil, nil)

	_, err := generationService.List(context.Background(), &domain.RunListQuery{Status: "completed"})
	if !errors.Is(err, service.ErrInvalidRunListStatus) {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}
