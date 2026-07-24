package task

import (
	"context"
	"log"

	"github.com/1024XEngineer/Holonic-Asset/internal/task/repository"
	"github.com/1024XEngineer/Holonic-Asset/pkg/queue"
)

type Dispatcher struct {
	repo      repository.TaskRepository
	publisher queue.Publisher
}

func NewDispatcher(repo repository.TaskRepository, p queue.Publisher) *Dispatcher {
	return &Dispatcher{repo: repo, publisher: p}
}

func (d *Dispatcher) Run(ctx context.Context, batchSize int) (int, error) {
	tasks, err := d.repo.FetchUndispatched(ctx, batchSize)
	if err != nil {
		return 0, err
	}

	dispatched := 0
	for i := range tasks {
		t := &tasks[i]

		job, err := BuildJob(t)
		if err != nil {
			log.Printf("task dispatcher: skip %d: %v", t.ID, err)
			continue
		}

		jobID, err := d.publisher.Publish(ctx, job)
		if err != nil {
			log.Printf("task dispatcher: publish %d (%s): %v", t.ID, job.Kind(), err)
			continue
		}

		claimed, err := d.repo.Claim(ctx, t.ID, jobID)
		if err != nil {
			log.Printf("task dispatcher: claim %d: %v", t.ID, err)
			continue
		}
		if claimed {
			dispatched++
		}
	}

	return dispatched, nil
}
