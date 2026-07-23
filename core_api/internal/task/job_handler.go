package task

import (
	"context"
	"encoding/json"
	"fmt"
)

// GenerateCharacterProtoTypeHandler processes GenerateCharacterProtoTypeJob.
// It contains pure business logic — no River, Kafka, or queue dependency.
type GenerateCharacterProtoTypeHandler struct {
	// svc *service.TaskService  // inject when business logic is implemented
}

func NewGenerateCharacterProtoTypeHandler() *GenerateCharacterProtoTypeHandler {
	return &GenerateCharacterProtoTypeHandler{}
}

func (h *GenerateCharacterProtoTypeHandler) JobKind() string {
	return GenerateCharacterProtoTypeJob{}.Kind()
}

func (h *GenerateCharacterProtoTypeHandler) Handle(ctx context.Context, payload []byte) error {
	var job GenerateCharacterProtoTypeJob
	if err := json.Unmarshal(payload, &job); err != nil {
		return fmt.Errorf("task: unmarshal generate_character_prototype job: %w", err)
	}
	// TODO: implement business logic — call AI provider, update task status, etc.
	return nil
}

// GenerateCharacterAnimationHandler processes GenerateCharacterAnimationJob.
type GenerateCharacterAnimationHandler struct{}

func NewGenerateCharacterAnimationHandler() *GenerateCharacterAnimationHandler {
	return &GenerateCharacterAnimationHandler{}
}

func (h *GenerateCharacterAnimationHandler) JobKind() string {
	return GenerateCharacterAnimationJob{}.Kind()
}

func (h *GenerateCharacterAnimationHandler) Handle(ctx context.Context, payload []byte) error {
	var job GenerateCharacterAnimationJob
	if err := json.Unmarshal(payload, &job); err != nil {
		return fmt.Errorf("task: unmarshal generate_character_animation job: %w", err)
	}
	// TODO: implement business logic
	return nil
}

// GenerateObjectProtoTypeHandler processes GenerateObjectProtoTypeJob.
type GenerateObjectProtoTypeHandler struct{}

func NewGenerateObjectProtoTypeHandler() *GenerateObjectProtoTypeHandler {
	return &GenerateObjectProtoTypeHandler{}
}

func (h *GenerateObjectProtoTypeHandler) JobKind() string {
	return GenerateObjectProtoTypeJob{}.Kind()
}

func (h *GenerateObjectProtoTypeHandler) Handle(ctx context.Context, payload []byte) error {
	var job GenerateObjectProtoTypeJob
	if err := json.Unmarshal(payload, &job); err != nil {
		return fmt.Errorf("task: unmarshal generate_object_prototype job: %w", err)
	}
	// TODO: implement business logic
	return nil
}

// GenerateObjectAnimationHandler processes GenerateObjectAnimationJob.
type GenerateObjectAnimationHandler struct{}

func NewGenerateObjectAnimationHandler() *GenerateObjectAnimationHandler {
	return &GenerateObjectAnimationHandler{}
}

func (h *GenerateObjectAnimationHandler) JobKind() string {
	return GenerateObjectAnimationJob{}.Kind()
}

func (h *GenerateObjectAnimationHandler) Handle(ctx context.Context, payload []byte) error {
	var job GenerateObjectAnimationJob
	if err := json.Unmarshal(payload, &job); err != nil {
		return fmt.Errorf("task: unmarshal generate_object_animation job: %w", err)
	}
	// TODO: implement business logic
	return nil
}

// GenerateTileSetHandler processes GenerateTileSetJob.
type GenerateTileSetHandler struct{}

func NewGenerateTileSetHandler() *GenerateTileSetHandler {
	return &GenerateTileSetHandler{}
}

func (h *GenerateTileSetHandler) JobKind() string {
	return GenerateTileSetJob{}.Kind()
}

func (h *GenerateTileSetHandler) Handle(ctx context.Context, payload []byte) error {
	var job GenerateTileSetJob
	if err := json.Unmarshal(payload, &job); err != nil {
		return fmt.Errorf("task: unmarshal generate_tileset job: %w", err)
	}
	// TODO: implement business logic
	return nil
}
