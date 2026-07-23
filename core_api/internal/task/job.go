// Package task defines job types and handlers for task-related asynchronous work.
//
// Job structs implement queue.Job (and therefore satisfy river.JobArgs via
// structural typing) WITHOUT importing any queue infrastructure package.
// This allows the task module to remain fully decoupled from River, Kafka,
// or any specific queue technology.
package task

// GenerateCharacterProtoTypeJob triggers AI generation of a character prototype.
type GenerateCharacterProtoTypeJob struct {
	TaskID    uint `json:"task_id"`
	ProjectID uint `json:"project_id"`
	AssetID   uint `json:"asset_id"`
}

func (GenerateCharacterProtoTypeJob) Kind() string { return "generate_character_prototype" }

// GenerateCharacterAnimationJob triggers AI generation of a character animation.
type GenerateCharacterAnimationJob struct {
	TaskID    uint `json:"task_id"`
	ProjectID uint `json:"project_id"`
	AssetID   uint `json:"asset_id"`
}

func (GenerateCharacterAnimationJob) Kind() string { return "generate_character_animation" }

// RegenerateCharacterProtoTypeJob triggers re-generation of a character prototype.
type RegenerateCharacterProtoTypeJob struct {
	TaskID    uint `json:"task_id"`
	ProjectID uint `json:"project_id"`
	AssetID   uint `json:"asset_id"`
}

func (RegenerateCharacterProtoTypeJob) Kind() string { return "regenerate_character_prototype" }

// RegenerateCharacterAnimationJob triggers re-generation of a character animation.
type RegenerateCharacterAnimationJob struct {
	TaskID    uint `json:"task_id"`
	ProjectID uint `json:"project_id"`
	AssetID   uint `json:"asset_id"`
}

func (RegenerateCharacterAnimationJob) Kind() string { return "regenerate_character_animation" }

// RegenerateCharacterFramesJob triggers re-generation of character frames.
type RegenerateCharacterFramesJob struct {
	TaskID    uint `json:"task_id"`
	ProjectID uint `json:"project_id"`
	AssetID   uint `json:"asset_id"`
}

func (RegenerateCharacterFramesJob) Kind() string { return "regenerate_character_frames" }

// GenerateObjectProtoTypeJob triggers AI generation of an object prototype.
type GenerateObjectProtoTypeJob struct {
	TaskID    uint `json:"task_id"`
	ProjectID uint `json:"project_id"`
	AssetID   uint `json:"asset_id"`
}

func (GenerateObjectProtoTypeJob) Kind() string { return "generate_object_prototype" }

// GenerateObjectAnimationJob triggers AI generation of an object animation.
type GenerateObjectAnimationJob struct {
	TaskID    uint `json:"task_id"`
	ProjectID uint `json:"project_id"`
	AssetID   uint `json:"asset_id"`
}

func (GenerateObjectAnimationJob) Kind() string { return "generate_object_animation" }

// RegenerateObjectProtoTypeJob triggers re-generation of an object prototype.
type RegenerateObjectProtoTypeJob struct {
	TaskID    uint `json:"task_id"`
	ProjectID uint `json:"project_id"`
	AssetID   uint `json:"asset_id"`
}

func (RegenerateObjectProtoTypeJob) Kind() string { return "regenerate_object_prototype" }

// RegenerateObjectAnimationJob triggers re-generation of an object animation.
type RegenerateObjectAnimationJob struct {
	TaskID    uint `json:"task_id"`
	ProjectID uint `json:"project_id"`
	AssetID   uint `json:"asset_id"`
}

func (RegenerateObjectAnimationJob) Kind() string { return "regenerate_object_animation" }

// RegenerateObjectFramesJob triggers re-generation of object frames.
type RegenerateObjectFramesJob struct {
	TaskID    uint `json:"task_id"`
	ProjectID uint `json:"project_id"`
	AssetID   uint `json:"asset_id"`
}

func (RegenerateObjectFramesJob) Kind() string { return "regenerate_object_frames" }

// GenerateTileSetJob triggers AI generation of a tile set.
type GenerateTileSetJob struct {
	TaskID    uint `json:"task_id"`
	ProjectID uint `json:"project_id"`
}

func (GenerateTileSetJob) Kind() string { return "generate_tileset" }

// RegenerateItemJob triggers re-generation of a single item.
type RegenerateItemJob struct {
	TaskID    uint `json:"task_id"`
	ProjectID uint `json:"project_id"`
	AssetID   uint `json:"asset_id"`
	ItemIndex int  `json:"item_index"`
}

func (RegenerateItemJob) Kind() string { return "regenerate_item" }

// RegenerateTilesJob triggers re-generation of tiles.
type RegenerateTilesJob struct {
	TaskID    uint `json:"task_id"`
	ProjectID uint `json:"project_id"`
	AssetID   uint `json:"asset_id"`
}

func (RegenerateTilesJob) Kind() string { return "regenerate_tiles" }
