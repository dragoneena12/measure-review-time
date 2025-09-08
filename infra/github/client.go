package github

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/dragoneena12/measure-review-time/domain/entity"
	"github.com/dragoneena12/measure-review-time/domain/repository"
	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	logger *slog.Logger
}

func NewClient(token string, logger *slog.Logger) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	logger.Info("GitHub client initialized")

	return &Client{
		client: client,
		logger: logger,
	}
}

func (c *Client) List(ctx context.Context, owner, repo string, opts repository.ListOptions) ([]*entity.PullRequest, error) {
	// Build search query
	query := fmt.Sprintf("repo:%s/%s is:pr", owner, repo)
	
	// Add state filter
	if opts.State != "" {
		if opts.State == "closed" {
			query += " is:closed"
		} else if opts.State == "open" {
			query += " is:open"
		}
	}
	
	// Add date filters
	if opts.Since != nil && opts.Until != nil {
		// When both are specified, use range syntax
		query += fmt.Sprintf(" created:%s..%s", opts.Since.Format("2006-01-02"), opts.Until.Format("2006-01-02"))
	} else if opts.Since != nil {
		// Only since is specified
		query += fmt.Sprintf(" created:>=%s", opts.Since.Format("2006-01-02"))
	} else if opts.Until != nil {
		// Only until is specified
		query += fmt.Sprintf(" created:<=%s", opts.Until.Format("2006-01-02"))
	}
	
	// Initialize result collection
	var allIssues []*github.Issue
	page := 1
	perPage := opts.PerPage
	if perPage == 0 {
		perPage = 100 // Default per page
	}
	
	// Fetch all pages
	for {
		searchOpts := &github.SearchOptions{
			Sort:  opts.Sort,
			Order: opts.Direction,
			ListOptions: github.ListOptions{
				PerPage: perPage,
				Page:    page,
			},
		}

		logAttrs := []any{
			slog.String("owner", owner),
			slog.String("repo", repo),
			slog.String("query", query),
			slog.Int("page", page),
			slog.Int("per_page", perPage),
		}
		if opts.Since != nil {
			logAttrs = append(logAttrs, slog.Time("since", *opts.Since))
		}
		if opts.Until != nil {
			logAttrs = append(logAttrs, slog.Time("until", *opts.Until))
		}
		c.logger.Info("Searching pull requests", logAttrs...)

		searchResult, resp, err := c.client.Search.Issues(ctx, query, searchOpts)
		if err != nil {
			c.logger.Error("Failed to search pull requests",
				slog.String("owner", owner),
				slog.String("repo", repo),
				slog.String("query", query),
				slog.Int("page", page),
				slog.String("error", err.Error()),
			)
			return nil, err
		}

		c.logger.Info("Successfully searched pull requests page",
			slog.String("owner", owner),
			slog.String("repo", repo),
			slog.Int("page", page),
			slog.Int("count", len(searchResult.Issues)),
			slog.Int("total_count", *searchResult.Total),
		)

		allIssues = append(allIssues, searchResult.Issues...)
		
		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	c.logger.Info("Fetched all pull requests",
		slog.String("owner", owner),
		slog.String("repo", repo),
		slog.Int("total_issues", len(allIssues)),
	)

	result := make([]*entity.PullRequest, 0, len(allIssues))
	totalIssues := len(allIssues)
	
	for i, issue := range allIssues {
		// Display progress
		c.logger.Info("Processing pull request",
			slog.String("progress", fmt.Sprintf("%d/%d", i+1, totalIssues)),
			slog.Int("number", *issue.Number),
		)
		
		// Get full PR details
		pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, *issue.Number)
		if err != nil {
			c.logger.Error("Failed to get PR details",
				slog.String("owner", owner),
				slog.String("repo", repo),
				slog.Int("number", *issue.Number),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
		
		pullRequest := c.convertToDomainEntity(pr)
		
		// Get review request time
		firstReviewRequestTime, err := c.getFirstReviewRequestTime(ctx, owner, repo, *pr.Number)
		if err != nil {
			c.logger.Warn("Failed to get review request time",
				slog.String("owner", owner),
				slog.String("repo", repo),
				slog.Int("number", *pr.Number),
				slog.String("error", err.Error()),
			)
			// Continue without review request time
		}
		pullRequest.FirstReviewRequestAt = firstReviewRequestTime
		
		firstReviewTime, firstApproveTime, err := c.getReviews(ctx, owner, repo, *pr.Number, firstReviewRequestTime)
		if err != nil {
			return nil, err
		}
		pullRequest.FirstReviewAt = firstReviewTime
		pullRequest.FirstApproveAt = firstApproveTime

		result = append(result, pullRequest)
	}

	return result, nil
}

func (c *Client) Get(ctx context.Context, owner, repo string, number int) (*entity.PullRequest, error) {
	c.logger.Info("Fetching single pull request",
		slog.String("owner", owner),
		slog.String("repo", repo),
		slog.Int("number", number),
	)

	pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		c.logger.Error("Failed to fetch pull request",
			slog.String("owner", owner),
			slog.String("repo", repo),
			slog.Int("number", number),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	c.logger.Info("Successfully fetched pull request",
		slog.String("owner", owner),
		slog.String("repo", repo),
		slog.Int("number", number),
	)

	pullRequest := c.convertToDomainEntity(pr)

	// Get review request time
	firstReviewRequestTime, err := c.getFirstReviewRequestTime(ctx, owner, repo, number)
	if err != nil {
		c.logger.Warn("Failed to get review request time",
			slog.String("owner", owner),
			slog.String("repo", repo),
			slog.Int("number", number),
			slog.String("error", err.Error()),
		)
		// Continue without review request time
	}
	pullRequest.FirstReviewRequestAt = firstReviewRequestTime

	firstReviewTime, firstApproveTime, err := c.getReviews(ctx, owner, repo, number, firstReviewRequestTime)
	if err != nil {
		return nil, err
	}
	pullRequest.FirstReviewAt = firstReviewTime
	pullRequest.FirstApproveAt = firstApproveTime

	return pullRequest, nil
}

func (c *Client) getReviews(ctx context.Context, owner, repo string, number int, firstReviewRequestAt *time.Time) (firstReviewTime, firstApproveTime *time.Time, err error) {
	c.logger.Debug("Fetching reviews for pull request",
		slog.String("owner", owner),
		slog.String("repo", repo),
		slog.Int("number", number),
	)

	reviews, _, err := c.client.PullRequests.ListReviews(ctx, owner, repo, number, nil)
	if err != nil {
		c.logger.Error("Failed to fetch reviews",
			slog.String("owner", owner),
			slog.String("repo", repo),
			slog.Int("number", number),
			slog.String("error", err.Error()),
		)
		return nil, nil, err
	}

	type reviewInfo struct {
		state string
		time  time.Time
	}

	var reviewList []reviewInfo
	for _, review := range reviews {
		if review.State != nil && *review.State != "" && *review.State != "PENDING" {
			// Skip reviews from GitHub Apps (bots)
			if review.User != nil && review.User.Type != nil && *review.User.Type == "Bot" {
				c.logger.Debug("Skipping review from GitHub App",
					slog.String("user", review.User.GetLogin()),
					slog.Int("pr_number", number),
				)
				continue
			}
			
			submittedAt := review.GetSubmittedAt().Time
			
			// Skip reviews that occurred before the review request
			if firstReviewRequestAt != nil && submittedAt.Before(*firstReviewRequestAt) {
				c.logger.Debug("Skipping review before review request",
					slog.Time("review_time", submittedAt),
					slog.Time("request_time", *firstReviewRequestAt),
					slog.Int("pr_number", number),
				)
				continue
			}
			
			reviewList = append(reviewList, reviewInfo{
				state: review.GetState(),
				time:  submittedAt,
			})
		}
	}

	c.logger.Debug("Successfully fetched reviews",
		slog.String("owner", owner),
		slog.String("repo", repo),
		slog.Int("number", number),
		slog.Int("review_count", len(reviewList)),
	)

	// Sort by time
	sort.Slice(reviewList, func(i, j int) bool {
		return reviewList[i].time.Before(reviewList[j].time)
	})

	// Find first review and first approve
	for _, r := range reviewList {
		if firstReviewTime == nil {
			firstReviewTime = &r.time
		}
		if firstApproveTime == nil && r.state == "APPROVED" {
			firstApproveTime = &r.time
		}
	}

	return firstReviewTime, firstApproveTime, nil
}

func (c *Client) getFirstReviewRequestTime(ctx context.Context, owner, repo string, number int) (*time.Time, error) {
	c.logger.Debug("Fetching timeline events for pull request",
		slog.String("owner", owner),
		slog.String("repo", repo),
		slog.Int("number", number),
	)

	var allEvents []*github.Timeline
	page := 1
	
	// Fetch all timeline events (handling pagination)
	for {
		opts := &github.ListOptions{
			Page:    page,
			PerPage: 100,
		}
		
		events, resp, err := c.client.Issues.ListIssueTimeline(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, err
		}
		
		allEvents = append(allEvents, events...)
		
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	c.logger.Debug("Fetched timeline events",
		slog.String("owner", owner),
		slog.String("repo", repo),
		slog.Int("number", number),
		slog.Int("event_count", len(allEvents)),
	)

	var firstReviewRequestTime *time.Time
	for _, event := range allEvents {
		// Check for review_requested event
		if event.Event != nil && *event.Event == "review_requested" {
			if event.CreatedAt != nil {
				if firstReviewRequestTime == nil || event.CreatedAt.Before(*firstReviewRequestTime) {
					firstReviewRequestTime = &event.CreatedAt.Time
				}
			}
		}
	}

	if firstReviewRequestTime != nil {
		c.logger.Debug("Found first review request time",
			slog.String("owner", owner),
			slog.String("repo", repo),
			slog.Int("number", number),
			slog.Time("review_requested_at", *firstReviewRequestTime),
		)
	} else {
		c.logger.Debug("No review request events found",
			slog.String("owner", owner),
			slog.String("repo", repo),
			slog.Int("number", number),
			slog.Int("total_events", len(allEvents)),
		)
	}

	return firstReviewRequestTime, nil
}

func (c *Client) convertToDomainEntity(pr *github.PullRequest) *entity.PullRequest {
	pullRequest := &entity.PullRequest{
		ID:        pr.GetID(),
		Number:    pr.GetNumber(),
		Title:     pr.GetTitle(),
		Author:    pr.GetUser().GetLogin(),
		State:     pr.GetState(),
		CreatedAt: pr.GetCreatedAt().Time,
	}

	if pr.MergedAt != nil {
		mergedAt := pr.GetMergedAt().Time
		pullRequest.MergedAt = &mergedAt
	}

	if pr.ClosedAt != nil {
		closedAt := pr.GetClosedAt().Time
		pullRequest.ClosedAt = &closedAt
	}

	return pullRequest
}