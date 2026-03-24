package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// commandTimeout is the maximum time to wait for a gh CLI command.
const commandTimeout = 2 * time.Minute

// ghAPIIssue is the JSON shape returned by the GitHub API.
type ghAPIIssue struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	State       string `json:"state"`
	Body        string `json:"body"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	PullRequest *struct {
		URL string `json:"url"`
	} `json:"pull_request"`
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
	Assignees []struct {
		Login string `json:"login"`
	} `json:"assignees"`
	Milestone *struct {
		Title  string `json:"title"`
		Number int    `json:"number"`
	} `json:"milestone"`
}

// GHClient implements Client using the gh CLI.
type GHClient struct{}

// NewGHClient returns a Client backed by the gh CLI.
func NewGHClient() Client {
	return &GHClient{}
}

func (c *GHClient) ResolveRepo() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolving repo: %w (is gh authenticated and in a git repo?)", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *GHClient) FetchIssue(repo string, number int) (*Issue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", "api",
		fmt.Sprintf("repos/%s/issues/%d", repo, number),
		"--header", "Accept: application/vnd.github+json",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("fetching issue #%d: %w", number, err)
	}

	var apiIssue ghAPIIssue
	if err := json.Unmarshal(out, &apiIssue); err != nil {
		return nil, fmt.Errorf("parsing issue #%d: %w", number, err)
	}

	if apiIssue.PullRequest != nil {
		return nil, fmt.Errorf("issue #%d is a pull request, not an issue", number)
	}

	return apiIssueToIssue(&apiIssue)
}

func (c *GHClient) FetchAllIssues(repo string, state string) ([]*Issue, error) {
	if state == "" {
		state = "open"
	}

	// gh api --paginate concatenates JSON arrays: [{...}][{...}]
	// Use --jq '.[]' to emit newline-delimited JSON objects instead.
	// Use a longer timeout for paginated requests.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", "api",
		fmt.Sprintf("repos/%s/issues", repo),
		"--header", "Accept: application/vnd.github+json",
		"--method", "GET",
		"-f", fmt.Sprintf("state=%s", state),
		"-f", "per_page=100",
		"--paginate",
		"--jq", ".[]",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("fetching issues: %w", err)
	}

	if len(strings.TrimSpace(string(out))) == 0 {
		return nil, nil
	}

	// Parse newline-delimited JSON objects
	var issues []*Issue
	dec := json.NewDecoder(strings.NewReader(string(out)))
	for dec.More() {
		var ai ghAPIIssue
		if err := dec.Decode(&ai); err != nil {
			return nil, fmt.Errorf("parsing issues: %w", err)
		}
		// Skip pull requests (GitHub API returns them as issues)
		if ai.PullRequest != nil {
			continue
		}
		issue, err := apiIssueToIssue(&ai)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

func (c *GHClient) PushIssue(repo string, issue *Issue) error {
	return pushIssueViaGH(repo, issue)
}

func apiIssueToIssue(ai *ghAPIIssue) (*Issue, error) {
	createdAt, err := time.Parse(time.RFC3339, ai.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, ai.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("parsing updated_at: %w", err)
	}

	var labels []string
	for _, l := range ai.Labels {
		labels = append(labels, l.Name)
	}

	var assignees []string
	for _, a := range ai.Assignees {
		assignees = append(assignees, a.Login)
	}

	var milestone string
	if ai.Milestone != nil {
		milestone = ai.Milestone.Title
	}

	return &Issue{
		Number:    ai.Number,
		Title:     ai.Title,
		State:     ai.State,
		Body:      ai.Body,
		Labels:    labels,
		Assignees: assignees,
		Milestone: milestone,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Author:    ai.User.Login,
	}, nil
}
