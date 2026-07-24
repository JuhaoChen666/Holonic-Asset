package dao

import (
	"context"
	"fmt"

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

	FetchUndispatched(ctx context.Context, limit int) ([]*Task, error)
	Claim(ctx context.Context, taskID uint, jobID int64) (bool, error)

	ListByProjectID(ctx context.Context, projectID uint) ([]*Task, error)
	GetDetail(ctx context.Context, taskID uint) (*Task, error)
	Transition(ctx context.Context, from, to uint) error
	Cancel(ctx context.Context, taskID uint) error
}

type TaskDaoImpl struct {
	DB *gorm.DB
}

func NewTaskDao(db *gorm.DB) *TaskDaoImpl {
	return &TaskDaoImpl{DB: db}
}

func (d *TaskDaoImpl) Create(ctx context.Context, task *Task) error {
	return d.DB.WithContext(ctx).Create(task).Error
}

func (d *TaskDaoImpl) FetchUndispatched(ctx context.Context, limit int) ([]*Task, error) {
	var tasks []*Task
	err := d.DB.WithContext(ctx).
		Where("status = 0 AND job_id = 0").
		Order("id ASC").
		Limit(limit).
		Find(&tasks).Error
	if err != nil {
		return nil, fmt.Errorf("dao: fetch undispatched: %w", err)
	}
	return tasks, nil
}

func (d *TaskDaoImpl) Claim(ctx context.Context, taskID uint, jobID int64) (bool, error) {
	result := d.DB.WithContext(ctx).
		Model(&Task{}).
		Where("id = ? AND status = 0", taskID).
		Updates(map[string]any{
			"status": 1, // StatusPending
			"job_id": jobID,
		})
	if result.Error != nil {
		return false, fmt.Errorf("dao: claim task %d: %w", taskID, result.Error)
	}
	return result.RowsAffected > 0, nil
}

func (d *TaskDaoImpl) ListByProjectID(ctx context.Context, projectID uint) ([]*Task, error) {
	return []*Task{}, nil
}
