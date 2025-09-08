package printer

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dragoneena12/measure-review-time/domain/entity"
	"github.com/dragoneena12/measure-review-time/domain/repository"
)

type CSVPrinter struct {
	writer io.Writer
}

func NewCSVPrinter() repository.Printer {
	return &CSVPrinter{
		writer: os.Stdout,
	}
}

func (p *CSVPrinter) Print(owner, repo string, metrics []*entity.ReviewMetrics) error {
	fmt.Fprintln(p.writer, "PR_Number,Title,Author,Created_At,Time_To_Review,Time_To_Approve")

	for _, metric := range metrics {
		pr := metric.PullRequest

		timeToReview := ""
		if metric.TimeToReview != nil {
			timeToReview = formatDuration(*metric.TimeToReview)
		}

		timeToApprove := ""
		if metric.TimeToApprove != nil {
			timeToApprove = formatDuration(*metric.TimeToApprove)
		}

		title := strings.ReplaceAll(pr.Title, ",", ";")
		title = strings.ReplaceAll(title, "\"", "'")

		fmt.Fprintf(p.writer, "%d,\"%s\",%s,%s,%s,%s\n",
			pr.Number,
			title,
			pr.Author,
			pr.CreatedAt.Format("2006-01-02 15:04:05"),
			timeToReview,
			timeToApprove,
		)
	}

	return nil
}