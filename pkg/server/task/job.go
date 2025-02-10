package task

import (
	"context"
)

type Job func(context.Context)
type JobWithError func(context.Context) error
