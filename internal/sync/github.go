package sync

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ghAPIIssue is the JSON shape returned by the GitHub API.
type ghAPIIssue struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
	Assignees []struct {
		Login string `json:"login"`
	} `json:"assignees"`
	Milestone *struct {
		Title string `json:"title"`
	} `json:"milestone"`
}

// ResolveRepo determines the current GitHub repo from git remotes via gh.
func ResolveRepo() (string, error) {
	cmd := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolving repo: %w (is gh authenticated and in a git repo?)", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// FetchIssue fetches a single issue from GitHub using the gh CLI.
func FetchIssue(repo string, number int) (*Issue, error) {
	cmd := exec.Command("gh", "api",
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

	return apiIssueToIssue(&apiIssue)
}

// FetchAllIssues fetches all open issues from GitHub using the gh CLI.
func FetchAllIssues(repo string, state string) ([]*Issue, error) {
	if state == "" {
		state = "open"
	}

	cmd := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/issues", repo),
		"--header", "Accept: application/vnd.github+json",
		"--method", "GET",
		"-f", fmt.Sprintf("state=%s", state),
		"-f", "per_page=100",
		"--paginate",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("fetching issues: %w", err)
	}

	var apiIssues []ghAPIIssue
	if err := json.Unmarshal(out, &apiIssues); err != nil {
		return nil, fmt.Errorf("parsing issues: %w", err)
	}

	var issues []*Issue
	for _, ai := range apiIssues {
		// Skip pull requests (GitHub API returns them as issues too)
		issue, err := apiIssueToIssue(&ai)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}

	return issues, nil
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
