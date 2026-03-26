package crawler

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

// GitHubClient wraps the GitHub API with rate-limit awareness.
type GitHubClient struct {
	client *github.Client
	token  string
}

// NewGitHubClient creates an authenticated GitHub API client.
// If token is empty, the client uses unauthenticated requests (60 req/hr limit).
func NewGitHubClient(token string) *GitHubClient {
	var httpClient *http.Client

	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient = oauth2.NewClient(context.Background(), ts)
	}

	return &GitHubClient{
		client: github.NewClient(httpClient),
		token:  token,
	}
}

// CodeSearchResult holds a single code search match.
type CodeSearchResult struct {
	HTMLURL   string
	RepoOwner string
	RepoName  string
	FilePath  string
	Stars     int
	Forks     int
	Watchers  int
	UpdatedAt *time.Time
}

// SearchCode searches GitHub Code for files matching the query.
// Returns up to maxResults results, respecting rate limits.
func (g *GitHubClient) SearchCode(ctx context.Context, query string, maxResults int) ([]CodeSearchResult, error) {
	var results []CodeSearchResult
	page := 1
	perPage := 100 // GitHub max

	for len(results) < maxResults {
		opts := &github.SearchOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		}

		resp, httpResp, err := g.client.Search.Code(ctx, query, opts)
		if err != nil {
			// Handle rate limit
			if _, ok := err.(*github.RateLimitError); ok {
				return results, fmt.Errorf("github rate limit hit: %w", err)
			}
			return results, fmt.Errorf("github search error: %w", err)
		}
		if httpResp != nil {
			httpResp.Body.Close()
		}

		for _, item := range resp.CodeResults {
			if item.Repository == nil || item.HTMLURL == nil {
				continue
			}

			stars := 0
			forks := 0
			if item.Repository.StargazersCount != nil {
				stars = *item.Repository.StargazersCount
			}
			if item.Repository.ForksCount != nil {
				forks = *item.Repository.ForksCount
			}
			watchers := 0
			if item.Repository.WatchersCount != nil {
				watchers = *item.Repository.WatchersCount
			}

			owner := ""
			repoName := ""
			if item.Repository.Owner != nil && item.Repository.Owner.Login != nil {
				owner = *item.Repository.Owner.Login
			}
			if item.Repository.Name != nil {
				repoName = *item.Repository.Name
			}

			var updatedAt *time.Time
			if item.Repository.UpdatedAt != nil {
				t := item.Repository.UpdatedAt.Time
				updatedAt = &t
			}

			filePath := ""
			if item.Path != nil {
				filePath = *item.Path
			}

			results = append(results, CodeSearchResult{
				HTMLURL:   *item.HTMLURL,
				RepoOwner: owner,
				RepoName:  repoName,
				FilePath:  filePath,
				Stars:     stars,
				Forks:     forks,
				Watchers:  watchers,
				UpdatedAt: updatedAt,
			})

			if len(results) >= maxResults {
				break
			}
		}

		// Check if there are more pages
		if resp.GetIncompleteResults() || len(resp.CodeResults) < perPage {
			break
		}
		page++

		// Respect secondary rate limits — 1 req/sec for search
		time.Sleep(1 * time.Second)
	}

	return results, nil
}

// GetFileContent fetches the raw content of a file from a GitHub repo.
func (g *GitHubClient) GetFileContent(ctx context.Context, owner, repo, path string) (string, *time.Time, error) {
	file, _, resp, err := g.client.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		return "", nil, fmt.Errorf("get file content: %w", err)
	}
	if resp != nil {
		resp.Body.Close()
	}
	if file == nil {
		return "", nil, fmt.Errorf("empty file response")
	}

	var content string
	if file.Content != nil {
		decoded, err := base64.StdEncoding.DecodeString(*file.Content)
		if err != nil {
			return "", nil, fmt.Errorf("decode content: %w", err)
		}
		content = string(decoded)
	}

	// Get last commit date for the file
	commits, _, resp2, err := g.client.Repositories.ListCommits(ctx, owner, repo, &github.CommitsListOptions{
		Path: path,
		ListOptions: github.ListOptions{PerPage: 1},
	})
	if resp2 != nil {
		resp2.Body.Close()
	}

	var lastUpdated *time.Time
	if err == nil && len(commits) > 0 && commits[0].Commit != nil && commits[0].Commit.Author != nil {
		t := commits[0].Commit.Author.Date.Time
		lastUpdated = &t
	}

	return content, lastUpdated, nil
}

// GetRateLimit returns the current GitHub API rate limit status.
func (g *GitHubClient) GetRateLimit(ctx context.Context) (remaining, limit int, err error) {
	rl, _, err := g.client.RateLimit.Get(ctx)
	if err != nil {
		return 0, 0, err
	}
	if rl.Core != nil {
		return rl.Core.Remaining, rl.Core.Limit, nil
	}
	return 0, 0, nil
}
