package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/dragoneena12/measure-review-time/domain/entity"
	"github.com/dragoneena12/measure-review-time/domain/repository"
)

type MeasureReviewTimeUseCase struct {
	prRepo repository.PullRequestRepository
}

func NewMeasureReviewTimeUseCase(prRepo repository.PullRequestRepository) *MeasureReviewTimeUseCase {
	return &MeasureReviewTimeUseCase{
		prRepo: prRepo,
	}
}

type MeasureOptions struct {
	Owner string
	Repo  string
	State string
	Since *time.Time
	Until *time.Time
}

func (u *MeasureReviewTimeUseCase) Execute(ctx context.Context, opts MeasureOptions) ([]*entity.ReviewMetrics, error) {
	metrics := make([]*entity.ReviewMetrics, 0)

	listOpts := repository.ListOptions{
		State:     opts.State,
		Sort:      "created",
		Direction: "desc",
		Since:     opts.Since,
		Until:     opts.Until,
		PerPage:   100,
	}

	prs, err := u.prRepo.List(ctx, opts.Owner, opts.Repo, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	for _, pr := range prs {
		metric := pr.CalculateMetrics()
		metrics = append(metrics, metric)
	}

	return metrics, nil
}
