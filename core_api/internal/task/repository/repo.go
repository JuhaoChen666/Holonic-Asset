package repository

import (
	"context"
	"fmt"

	"github.com/1024XEngineer/Holonic-Asset/internal/task/domain"
	"github.com/1024XEngineer/Holonic-Asset/internal/task/repository/dao"
)

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) (uint, error)
	FetchUndispatched(ctx context.Context, limit int) ([]domain.Task, error)
	Claim(ctx context.Context, taskID uint, jobID int64) (bool, error)

	ListByProjectID(ctx context.Context, projectID uint) ([]domain.Task, error)
}

type TaskRepositoryImpl struct {
	TaskDao dao.TaskDao
}

func NewTaskRepository(d dao.TaskDao) *TaskRepositoryImpl {
	return &TaskRepositoryImpl{TaskDao: d}
}

func (r *TaskRepositoryImpl) Create(ctx context.Context, task *domain.Task) (uint, error) {
	task.Status = domain.StatusWaiting

	daoTask := &dao.Task{
		Uid:         task.Uid,
		ProjectID:   task.ProjectID,
		Name:        task.Name,
		Description: task.Description,
		Type:        string(task.Type),
		Status:      uint(task.Status),
	}

	if err := r.TaskDao.Create(ctx, daoTask); err != nil {
		return 0, fmt.Errorf("repo: insert task: %w", err)
	}
	task.ID = daoTask.ID
	return task.ID, nil
}

func (r *TaskRepositoryImpl) FetchUndispatched(ctx context.Context, limit int) ([]domain.Task, error) {
	daoList, err := r.TaskDao.FetchUndispatched(ctx, limit)
	if err != nil {
		return nil, err
	}
	tasks := make([]domain.Task, len(daoList))
	for i, dt := range daoList {
		tasks[i] = domain.Task{
			ID:          dt.ID,
			Uid:         dt.Uid,
			ProjectID:   dt.ProjectID,
			JobID:       dt.JobID,
			Name:        dt.Name,
			Description: dt.Description,
			Type:        domain.TaskType(dt.Type),
			Status:      domain.Status(dt.Status),
			Metadata:    dt.Metadata,
		}
	}
	return tasks, nil
}

func (r *TaskRepositoryImpl) Claim(ctx context.Context, taskID uint, jobID int64) (bool, error) {
	return r.TaskDao.Claim(ctx, taskID, jobID)
}

func (r *TaskRepositoryImpl) ListByProjectID(ctx context.Context, projectID uint) ([]domain.Task, error) {
	_, err := r.TaskDao.ListByProjectID(ctx, projectID)
	return []domain.Task{}, err
}
