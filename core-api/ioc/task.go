// Package ioc is the composition root — it wires business handlers to
// infrastructure modules.
package ioc

import (
	"context"
	"fmt"

	"github.com/1024XEngineer/Holonic-Asset/config"
	"github.com/1024XEngineer/Holonic-Asset/internal/module/task"
)

func InitTask(
	ctx context.Context,
	cfg config.QueueConfig,
) (task.Queue, error) {
	queue, err := task.NewQueue(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("ioc: initialize task queue: %w", err)
	}
	return queue, nil
}
