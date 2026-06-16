package agents

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

type GitHubClient struct {
	Token      string
	Owner      string
	Repo       string
	HTTPClient *http.Client
}

func NewGitHubClient(owner, repo string) *GitHubClient {
	return &GitHubClient{
		Token:      os.Getenv("GITHUB_TOKEN"),
		Owner:      owner,
		Repo:       repo,
		HTTPClient: &http.Client{},
	}
}

// CreateBranch creates a new branch from the default branch (e.g., main)
func (c *GitHubClient) CreateBranch(ctx context.Context, branchName string) error {
	// 1. Get default branch SHA
	// 2. Create new ref
	// For simplicity in this blueprint, we log the action.
	fmt.Printf("[GitHub] Creating branch: %s\n", branchName)
	return nil
}

// CommitFile creates or updates a file in the given branch
func (c *GitHubClient) CommitFile(ctx context.Context, branchName, path, content, message string) error {
	fmt.Printf("[GitHub] Committing to %s on branch %s\n", path, branchName)
	return nil
}

// CreatePullRequest opens a new PR and returns the PR number
func (c *GitHubClient) CreatePullRequest(ctx context.Context, title, body, head, base string) (int, error) {
	fmt.Printf("[GitHub] Creating PR: %s (%s -> %s)\n", title, head, base)
	// Returning a mock PR number for the Judge to use
	return 1, nil
}

// PostComment posts a comment on an issue or PR
func (c *GitHubClient) PostComment(ctx context.Context, prNumber int, body string) error {
	fmt.Printf("[GitHub] Posting comment on PR #%d: %s\n", prNumber, body)
	return nil
}
