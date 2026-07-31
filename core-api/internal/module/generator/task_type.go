package generator

type TaskType string

const (
	GenerateCharacterProtoType   TaskType = "generate_character_prototype"
	GenerateAnimation            TaskType = "generate_animation"
	RegenerateCharacterProtoType TaskType = "regenerate_character_prototype"
	RegenerateAnimation          TaskType = "regenerate_animation"
	RegenerateFrames             TaskType = "regenerate_frames"
	GenerateObjectProtoType      TaskType = "generate_object_prototype"
	RegenerateObjectProtoType    TaskType = "regenerate_object_prototype"
	GenerateTileSet              TaskType = "generate_tileset"
	RegenerateItem               TaskType = "regenerate_item"
	RegenerateTiles              TaskType = "regenerate_tiles"
)

func ProjectLevelTaskTypes() []TaskType {
	return []TaskType{
		GenerateCharacterProtoType,
		GenerateObjectProtoType,
		GenerateTileSet,
	}
}

func TaskTypes() []TaskType {
	return []TaskType{
		GenerateCharacterProtoType,
		GenerateAnimation,
		RegenerateCharacterProtoType,
		RegenerateAnimation,
		RegenerateFrames,
		GenerateObjectProtoType,
		RegenerateObjectProtoType,
		GenerateTileSet,
		RegenerateItem,
		RegenerateTiles,
	}
}
