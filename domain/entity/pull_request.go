package entity

import (
	"time"
)

type PullRequest struct {
	ID                   int64
	Number               int
	Title                string
	Author               string
	State                string
	CreatedAt            time.Time
	MergedAt             *time.Time
	ClosedAt             *time.Time
	FirstReviewRequestAt *time.Time
	FirstReviewAt        *time.Time
	FirstApproveAt       *time.Time
	ReviewDuration       *time.Duration
}

type ReviewMetrics struct {
	PullRequest   *PullRequest
	TimeToReview  *time.Duration
	TimeToApprove *time.Duration
	TotalDuration *time.Duration
}

func (pr *PullRequest) CalculateMetrics() *ReviewMetrics {
	metrics := &ReviewMetrics{
		PullRequest: pr,
	}

	// Use FirstReviewRequestAt as baseline if available, otherwise use CreatedAt
	baseTime := pr.CreatedAt
	if pr.FirstReviewRequestAt != nil {
		baseTime = *pr.FirstReviewRequestAt
	}

	if pr.FirstReviewAt != nil {
		duration := pr.FirstReviewAt.Sub(baseTime)
		metrics.TimeToReview = &duration
	}

	if pr.FirstApproveAt != nil {
		duration := pr.FirstApproveAt.Sub(baseTime)
		metrics.TimeToApprove = &duration
	}

	if pr.MergedAt != nil {
		duration := pr.MergedAt.Sub(baseTime)
		metrics.TotalDuration = &duration
	} else if pr.ClosedAt != nil {
		duration := pr.ClosedAt.Sub(baseTime)
		metrics.TotalDuration = &duration
	}

	return metrics
}
