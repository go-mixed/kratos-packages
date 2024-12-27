package worker

import (
	"context"
	"github.com/robfig/cron/v3"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/job"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/server/schedule"
	"time"
)

type IWorker interface {
	WithContext(ctx context.Context) IWorker
	OnceForCluster(key string) IWorker
	Submit(job job.Job)
	SubmitWait(job job.Job)
	SubmitAfter(delay time.Duration, job job.Job) *time.Timer
	SubmitTicker(interval time.Duration, job job.Job) *time.Ticker
	SubmitWithError(job job.JobWithError) error
	Cron(spec any, job job.Job) (cron.EntryID, error)
	CronWith(job job.Job) schedule.Spec
}
