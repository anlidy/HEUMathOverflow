package service

import (
	"MathOverflow/common/utils"
	"MathOverflow/services/forum/internal/repository"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

const (
	defaultFileCleanupQueueSize = 256
	defaultFileCleanupWorkers   = 2
	defaultFileDeleteTimeout    = 10 * time.Second
)

type fileDeleteJob struct {
	name      string
	bucket    string
	urls      []string
	urlIsPath bool
}

type fileCleanup struct {
	serviceName string
	fileRepo    repository.FileRepo
	jobs        chan fileDeleteJob
}

func newFileCleanup(serviceName string, fileRepo repository.FileRepo) *fileCleanup {
	c := &fileCleanup{
		serviceName: serviceName,
		fileRepo:    fileRepo,
		jobs:        make(chan fileDeleteJob, defaultFileCleanupQueueSize),
	}
	for i := 0; i < defaultFileCleanupWorkers; i++ {
		go c.run()
	}
	return c
}

func (c *fileCleanup) SubmitObjectURLs(name, bucket string, urls []string) {
	c.submit(fileDeleteJob{name: name, bucket: bucket, urls: cloneNonEmptyStrings(urls)})
}

func (c *fileCleanup) SubmitRelativePaths(name, bucket string, paths []string) {
	c.submit(fileDeleteJob{name: name, bucket: bucket, urls: cloneNonEmptyStrings(paths), urlIsPath: true})
}

func (c *fileCleanup) submit(job fileDeleteJob) {
	if c == nil || c.fileRepo == nil || len(job.urls) == 0 {
		return
	}
	select {
	case c.jobs <- job:
	default:
		utils.Logger().WithField("service", c.serviceName).WithField("job", job.name).Warn("file cleanup queue is full, dropping job")
	}
}

func (c *fileCleanup) run() {
	for job := range c.jobs {
		ctx, cancel := context.WithTimeout(context.Background(), defaultFileDeleteTimeout)
		c.handleJob(ctx, job)
		cancel()
	}
}

func (c *fileCleanup) handleJob(ctx context.Context, job fileDeleteJob) {
	for _, item := range job.urls {
		filename := item
		if !job.urlIsPath {
			filename = strings.TrimPrefix(item, fmt.Sprintf("/api/v1/%s/file/", job.bucket))
		}
		if strings.TrimSpace(filename) == "" {
			continue
		}
		if err := c.fileRepo.DeleteFile(ctx, filename); err != nil {
			if minio.ToErrorResponse(err).Code == "NoSuchKey" {
				utils.WithContext(ctx).WithField("service", c.serviceName).WithField("filename", filename).Info("minio file not found when deleting file")
				continue
			}
			utils.WithContext(ctx).WithField("service", c.serviceName).WithField("job", job.name).WithField("filename", filename).WithError(err).Error("delete file failed")
		}
	}
}

func cloneNonEmptyStrings(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		result = append(result, item)
	}
	return result
}
