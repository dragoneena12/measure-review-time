package repository

import (
	"context"
	"time"

	"github.com/dragoneena12/measure-review-time/domain/entity"
)

type PullRequestRepository interface {
	List(ctx context.Context, owner, repo string, opts ListOptions) ([]*entity.PullRequest, error)
	Get(ctx context.Context, owner, repo string, number int) (*entity.PullRequest, error)
}

type ListOptions struct {
	State     string
	Sort      string
	Direction string
	Since     *time.Time
	Until     *time.Time
	PerPage   int
}