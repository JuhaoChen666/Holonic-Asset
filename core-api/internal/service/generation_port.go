package service

import (
	"context"

	"github.com/1024XEngineer/Holonic-Asset/internal/repository"
)

// Transaction exposes generation persistence and module services bound to the
// same database transaction.
type Transaction interface {
	Repository() repository.Repository
}

type UnitOfWork interface {
	WithinTransaction(ctx context.Context, work func(context.Context, Transaction) error) error
}
