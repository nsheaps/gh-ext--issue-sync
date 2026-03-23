package sync

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// PushIssue updates a GitHub issue from local issue data.
func PushIssue(repo string, issue *Issue) error {
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
	// Milestone requires the milestone number, not title - skip for now
	// TODO: resolve milestone title to number

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
