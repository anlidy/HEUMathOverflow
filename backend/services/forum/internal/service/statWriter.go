package service

import (
	"MathOverflow/common/utils"
	"context"
	"time"
)

const (
	defaultStatWriterQueueSize = 1024
	defaultStatWriterWorkers   = 2
	defaultStatWriteTimeout    = 2 * time.Second
)

type statWriteJob struct {
	name string
	fn   func(ctx context.Context) error
}

type statWriter struct {
	serviceName string
	jobs        chan statWriteJob
}

func newStatWriter(serviceName string) *statWriter {
	w := &statWriter{
		serviceName: serviceName,
		jobs:        make(chan statWriteJob, defaultStatWriterQueueSize),
	}
	for i := 0; i < defaultStatWriterWorkers; i++ {
		go w.run()
	}
	return w
}

func (w *statWriter) Submit(name string, fn func(ctx context.Context) error) {
	if w == nil || fn == nil {
		return
	}
	job := statWriteJob{name: name, fn: fn}
	select {
	case w.jobs <- job:
	default:
		utils.Logger().WithField("service", w.serviceName).WithField("job", name).Warn("stat write queue is full, dropping job")
	}
}

func (w *statWriter) run() {
	for job := range w.jobs {
		ctx, cancel := context.WithTimeout(context.Background(), defaultStatWriteTimeout)
		err := job.fn(ctx)
		cancel()
		if err != nil {
			utils.WithContext(ctx).WithField("service", w.serviceName).WithField("job", job.name).WithError(err).Error("async stat write failed")
		}
	}
}
