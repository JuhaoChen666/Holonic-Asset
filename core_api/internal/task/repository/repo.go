package repository

import (
	"context"

	"github.com/1024XEngineer/Holonic-Asset/internal/task/domain"
	"github.com/1024XEngineer/Holonic-Asset/internal/task/repository/dao"
)

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) (uint, error)
	ListByProjectID(ctx context.Context, projectID uint) ([]domain.Task, error)
}

type TaskRepositoryImpl struct {
	TaskDao dao.TaskDao
}

func (r *TaskRepositoryImpl) Create(ctx context.Context, task *domain.Task) (uint, error) {
	return 0, nil
}

func (r *TaskRepositoryImpl) ListByProjectID(ctx context.Context, projectID uint) ([]domain.Task, error) {
	_, err := r.TaskDao.ListByProjectID(ctx, projectID)
	return []domain.Task{}, err
}
