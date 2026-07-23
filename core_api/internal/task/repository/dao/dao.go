package dao

import (
	"context"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Task struct {
	ID          uint
	Uid         uint
	ProjectID   uint
	JobID       uint
	Name        string
	Description string
	Type        string
	Status      uint
	Metadata    datatypes.JSONMap `gorm:"type:jsonb;default:'{}'"`
}

type TaskDao interface {
	Create(ctx context.Context, task *Task) error

	// ListByProjectID returns tasks belonging to the specified project.
	ListByProjectID(ctx context.Context, projectID uint) ([]*Task, error)

	// GetDetail returns the current state and details of a task.
	GetDetail(ctx context.Context, taskID uint) (*Task, error)

	// Transition applies a guarded task state transition.
	Transition(ctx context.Context, from, to uint) error

	// Cancel requests cancellation of a task and its runnable steps.
	Cancel(ctx context.Context, taskID uint) error
}

type TaskDaoImpl struct {
	DB *gorm.DB
}

func NewTaskDao(db *gorm.DB) *TaskDaoImpl {
	return &TaskDaoImpl{
		DB: db,
	}
}

func (d *TaskDaoImpl) Create(ctx context.Context, task *Task) error {
	return nil
}

func (d *TaskDaoImpl) ListByProjectID(ctx context.Context, projectID uint) ([]*Task, error) {
	return []*Task{}, nil
}
