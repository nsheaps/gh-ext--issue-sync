package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// milestoneCache maps "repo:title" to milestone number to avoid repeated API calls.
var milestoneCache = map[string]int{}

// pushIssueViaGH updates a GitHub issue using the gh CLI.
func pushIssueViaGH(repo string, issue *Issue) error {
	if issue.Number <= 0 {
		return fmt.Errorf("invalid issue number: %d", issue.Number)
	}
	if issue.Title == "" {
		return fmt.Errorf("issue #%d: title cannot be empty", issue.Number)
	}
	if issue.State != "open" && issue.State != "closed" {
		return fmt.Errorf("issue #%d: invalid state %q (must be \"open\" or \"closed\")", issue.Number, issue.State)
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

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", "api",
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
// Results are cached per repo to avoid repeated API calls during batch pushes.
func resolveMilestoneNumber(repo string, title string) (int, error) {
	cacheKey := repo + ":" + title
	if num, ok := milestoneCache[cacheKey]; ok {
		return num, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", "api",
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

	// Cache all milestones for this repo
	for _, m := range milestones {
		milestoneCache[repo+":"+m.Title] = m.Number
	}

	if num, ok := milestoneCache[cacheKey]; ok {
		return num, nil
	}

	var available []string
	for _, m := range milestones {
		available = append(available, m.Title)
	}
	if len(available) > 0 {
		return 0, fmt.Errorf("milestone %q not found (available: %s)", title, strings.Join(available, ", "))
	}
	return 0, fmt.Errorf("milestone %q not found (no milestones exist in this repo)", title)
}
