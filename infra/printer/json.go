package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dragoneena12/measure-review-time/domain/entity"
	"github.com/dragoneena12/measure-review-time/domain/repository"
)

type JSONPrinter struct {
	writer io.Writer
}

func NewJSONPrinter() repository.Printer {
	return &JSONPrinter{
		writer: os.Stdout,
	}
}

func (p *JSONPrinter) Print(owner, repo string, metrics []*entity.ReviewMetrics) error {
	if len(metrics) == 0 {
		fmt.Fprintln(p.writer, "[]")
		return nil
	}

	output := map[string]any{
		"repository": fmt.Sprintf("%s/%s", owner, repo),
		"pull_requests": []map[string]any{},
	}

	pullRequests := []map[string]any{}
	for _, metric := range metrics {
		pr := metric.PullRequest
		prMap := map[string]any{
			"number":     pr.Number,
			"title":      pr.Title,
			"author":     pr.Author,
			"created_at": pr.CreatedAt.Format(time.RFC3339),
		}

		if metric.TimeToReview != nil {
			prMap["time_to_review"] = formatDuration(*metric.TimeToReview)
		}
		if metric.TimeToApprove != nil {
			prMap["time_to_approve"] = formatDuration(*metric.TimeToApprove)
		}

		pullRequests = append(pullRequests, prMap)
	}
	output["pull_requests"] = pullRequests

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}
	
	fmt.Fprintln(p.writer, string(jsonBytes))
	return nil
}