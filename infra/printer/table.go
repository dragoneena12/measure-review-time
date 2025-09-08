package printer

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/dragoneena12/measure-review-time/domain/entity"
	"github.com/dragoneena12/measure-review-time/domain/repository"
)

type TablePrinter struct {
	writer io.Writer
}

func NewTablePrinter() repository.Printer {
	return &TablePrinter{
		writer: os.Stdout,
	}
}

func (p *TablePrinter) Print(owner, repo string, metrics []*entity.ReviewMetrics) error {
	if len(metrics) == 0 {
		fmt.Fprintln(p.writer, "No pull requests found")
		return nil
	}

	fmt.Fprintf(p.writer, "\n=== PR Review Time Report for %s/%s ===\n\n", owner, repo)

	// Create a new tabwriter
	w := tabwriter.NewWriter(p.writer, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print table header
	fmt.Fprintln(w, "PR #\tAuthor\tCreated\tTime to Review\tTime to Approve\tTitle")
	fmt.Fprintln(w, "----\t------\t-------\t--------------\t---------------\t-----")

	// Print each PR
	for _, metric := range metrics {
		pr := metric.PullRequest
		
		title := pr.Title
		if len(title) > 60 {
			title = title[:57] + "..."
		}

		timeToReview := "N/A"
		if metric.TimeToReview != nil {
			timeToReview = formatDuration(*metric.TimeToReview)
		}

		timeToApprove := "N/A"
		if metric.TimeToApprove != nil {
			timeToApprove = formatDuration(*metric.TimeToApprove)
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			pr.Number,
			truncateString(pr.Author, 20),
			pr.CreatedAt.Format("2006-01-02 15:04"),
			timeToReview,
			timeToApprove,
			title,
		)
	}

	fmt.Fprintln(p.writer)
	return nil
}