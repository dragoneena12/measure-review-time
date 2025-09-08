package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/dragoneena12/measure-review-time/application/usecase"
	"github.com/dragoneena12/measure-review-time/domain/repository"
	"github.com/dragoneena12/measure-review-time/infra/github"
	"github.com/dragoneena12/measure-review-time/infra/printer"
)

func main() {
	var (
		owner  = flag.String("owner", "", "Repository owner (required)")
		repo   = flag.String("repo", "", "Repository name (required)")
		since  = flag.String("since", "", "Only PRs created after this date (YYYY-MM-DD)")
		format = flag.String("format", "table", "Output format (table, json, csv)")
		debug  = flag.Bool("debug", false, "Enable debug logging")
	)

	flag.StringVar(owner, "o", "", "Repository owner (short)")
	flag.StringVar(repo, "r", "", "Repository name (short)")
	flag.StringVar(format, "f", "table", "Output format (short)")

	flag.Parse()

	if *owner == "" {
		fmt.Fprintf(os.Stderr, "Error: Repository owner is required. Use -owner flag\n")
		os.Exit(1)
	}

	if *repo == "" {
		fmt.Fprintf(os.Stderr, "Error: Repository name is required. Use -repo flag\n")
		os.Exit(1)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Fprintf(os.Stderr, "Error: GITHUB_TOKEN environment variable is required\n")
		os.Exit(1)
	}

	ctx := context.Background()

	// Create logger
	logLevel := slog.LevelInfo
	if *debug {
		logLevel = slog.LevelDebug
	}
	
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})).With("component", "github_client")

	ghClient := github.NewClient(token, logger)
	measureUseCase := usecase.NewMeasureReviewTimeUseCase(ghClient)

	opts := usecase.MeasureOptions{
		Owner: *owner,
		Repo:  *repo,
		State: "closed",
	}

	if *since != "" {
		t, err := time.Parse("2006-01-02", *since)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid date format. Use YYYY-MM-DD\n")
			os.Exit(1)
		}
		opts.Since = &t
	}

	metrics, err := measureUseCase.Execute(ctx, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var p repository.Printer
	switch *format {
	case "json":
		p = printer.NewJSONPrinter()
	case "csv":
		p = printer.NewCSVPrinter()
	default:
		p = printer.NewTablePrinter()
	}

	if err := p.Print(*owner, *repo, metrics); err != nil {
		fmt.Fprintf(os.Stderr, "Error printing result: %v\n", err)
		os.Exit(1)
	}
}