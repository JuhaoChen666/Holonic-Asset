package domain

type Status uint
type TaskType string

const (
	StatusWaiting Status = iota
	StatusPending
	StatusProcessing
	StatusCompleted
	StatusFailed
	StatusCancelled
)

const (
	GenerateCharacterProtoType   TaskType = "generateCharacterProtoType"
	GenerateCharacterAnimation   TaskType = "generateCharacterAnimation"
	RegenerateCharacterProtoType TaskType = "regenerateCharacterProtoType"
	RegenerateCharacterAnimation TaskType = "regenerateCharacterAnimation"
	RegenerateCharacterFrames    TaskType = "regenerateCharacterFrames"

	GenerateObjectProtoType   TaskType = "generateObjectProtoType"
	GenerateObjectAnimation   TaskType = "generateObjectAnimation"
	RegenerateObjectProtoType TaskType = "regenerateObjectProtoType"
	RegenerateObjectAnimation TaskType = "regenerateObjectAnimation"
	RegenerateObjectFrames    TaskType = "regenerateObjectFrames"

	GenerateTileSet TaskType = "generateTileSet"
	RegenerateItem  TaskType = "regenerateItem"
	RegenerateTiles TaskType = "regenerateTiles"
)
