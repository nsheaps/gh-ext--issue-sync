package sync

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// pushIssueViaGH updates a GitHub issue using the gh CLI.
func pushIssueViaGH(repo string, issue *Issue) error {
	if issue.Number <= 0 {
		return fmt.Errorf("invalid issue number: %d", issue.Number)
	}
	if issue.Title == "" {
		return fmt.Errorf("issue #%d: title cannot be empty", issue.Number)
	}

	// Build the update payload - only mutable fields
	payload := map[string]interface{}{
		"title": issue.Title,
		"body":  issue.Body,
		"state": issue.State,
	}

	if issue.Labels != nil {
		payload["labels"] = issue.Labels
	}
	if issue.Assignees != nil {
		payload["assignees"] = issue.Assignees
	}

	// Resolve milestone title to number if set
	if issue.Milestone != "" {
		milestoneNum, err := resolveMilestoneNumber(repo, issue.Milestone)
		if err != nil {
			return fmt.Errorf("issue #%d: %w", issue.Number, err)
		}
		payload["milestone"] = milestoneNum
	} else {
		// Explicitly clear milestone
		payload["milestone"] = nil
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling update payload: %w", err)
	}

	cmd := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/issues/%d", repo, issue.Number),
		"--method", "PATCH",
		"--header", "Accept: application/vnd.github+json",
		"--input", "-",
	)
	cmd.Stdin = strings.NewReader(string(payloadJSON))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("updating issue #%d: %w\n%s", issue.Number, err, string(out))
	}

	return nil
}

// resolveMilestoneNumber looks up a milestone number by its title.
func resolveMilestoneNumber(repo string, title string) (int, error) {
	cmd := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/milestones", repo),
		"--header", "Accept: application/vnd.github+json",
		"--method", "GET",
		"-f", "state=all",
		"-f", "per_page=100",
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("fetching milestones: %w", err)
	}

	var milestones []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
	}
	if err := json.Unmarshal(out, &milestones); err != nil {
		return 0, fmt.Errorf("parsing milestones: %w", err)
	}

	for _, m := range milestones {
		if m.Title == title {
			return m.Number, nil
		}
	}

	return 0, fmt.Errorf("milestone %q not found", title)
}
