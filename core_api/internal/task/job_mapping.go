package task

import (
	"fmt"

	"github.com/1024XEngineer/Holonic-Asset/internal/common"
	"github.com/1024XEngineer/Holonic-Asset/internal/task/domain"
	"github.com/1024XEngineer/Holonic-Asset/pkg/queue"
)

// BuildJob maps a domain task to its corresponding job struct.
// Job types live in internal/common so every module can reference them.
func BuildJob(task *domain.Task) (queue.Job, error) {
	aid := assetIDFromMeta(task.Metadata)

	switch task.Type {
	case domain.GenerateCharacterProtoType:
		return common.GenerateCharacterProtoTypeJob{
			TaskID: task.ID, ProjectID: task.ProjectID, AssetID: aid,
		}, nil
	case domain.GenerateCharacterAnimation:
		return common.GenerateCharacterAnimationJob{
			TaskID: task.ID, ProjectID: task.ProjectID, AssetID: aid,
		}, nil
	case domain.RegenerateCharacterProtoType:
		return common.RegenerateCharacterProtoTypeJob{
			TaskID: task.ID, ProjectID: task.ProjectID, AssetID: aid,
		}, nil
	case domain.RegenerateCharacterAnimation:
		return common.RegenerateCharacterAnimationJob{
			TaskID: task.ID, ProjectID: task.ProjectID, AssetID: aid,
		}, nil
	case domain.RegenerateCharacterFrames:
		return common.RegenerateCharacterFramesJob{
			TaskID: task.ID, ProjectID: task.ProjectID, AssetID: aid,
		}, nil
	case domain.GenerateObjectProtoType:
		return common.GenerateObjectProtoTypeJob{
			TaskID: task.ID, ProjectID: task.ProjectID, AssetID: aid,
		}, nil
	case domain.GenerateObjectAnimation:
		return common.GenerateObjectAnimationJob{
			TaskID: task.ID, ProjectID: task.ProjectID, AssetID: aid,
		}, nil
	case domain.RegenerateObjectProtoType:
		return common.RegenerateObjectProtoTypeJob{
			TaskID: task.ID, ProjectID: task.ProjectID, AssetID: aid,
		}, nil
	case domain.RegenerateObjectAnimation:
		return common.RegenerateObjectAnimationJob{
			TaskID: task.ID, ProjectID: task.ProjectID, AssetID: aid,
		}, nil
	case domain.RegenerateObjectFrames:
		return common.RegenerateObjectFramesJob{
			TaskID: task.ID, ProjectID: task.ProjectID, AssetID: aid,
		}, nil
	case domain.GenerateTileSet:
		return common.GenerateTileSetJob{
			TaskID: task.ID, ProjectID: task.ProjectID,
		}, nil
	case domain.RegenerateItem:
		return common.RegenerateItemJob{
			TaskID:    task.ID,
			ProjectID: task.ProjectID,
			AssetID:   aid,
			ItemIndex: itemIndexFromMeta(task.Metadata),
		}, nil
	case domain.RegenerateTiles:
		return common.RegenerateTilesJob{
			TaskID: task.ID, ProjectID: task.ProjectID, AssetID: aid,
		}, nil
	default:
		return nil, fmt.Errorf("unknown task type %q", task.Type)
	}
}

func assetIDFromMeta(m map[string]any) uint {
	if m == nil {
		return 0
	}
	switch v := m["asset_id"].(type) {
	case float64:
		return uint(v)
	case uint:
		return v
	case int:
		return uint(v)
	}
	return 0
}

func itemIndexFromMeta(m map[string]any) int {
	if m == nil {
		return 0
	}
	switch v := m["item_index"].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return 0
}
