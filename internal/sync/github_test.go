package sync

import (
	"testing"
	"time"
)

func TestAPIIssueToIssue(t *testing.T) {
	t.Run("basic conversion", func(t *testing.T) {
		ai := &ghAPIIssue{
			Number:    42,
			Title:     "Fix login bug",
			State:     "open",
			Body:      "The login form crashes.",
			CreatedAt: "2026-01-15T10:30:00Z",
			UpdatedAt: "2026-03-20T14:22:00Z",
		}
		ai.User.Login = "octocat"
		ai.Labels = []struct {
			Name string `json:"name"`
		}{
			{Name: "bug"},
			{Name: "priority/high"},
		}
		ai.Assignees = []struct {
			Login string `json:"login"`
		}{
			{Login: "alice"},
		}
		ai.Milestone = &struct {
			Title  string `json:"title"`
			Number int    `json:"number"`
		}{
			Title:  "v1.0",
			Number: 1,
		}

		issue, err := apiIssueToIssue(ai)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if issue.Number != 42 {
			t.Errorf("number: got %d, want 42", issue.Number)
		}
		if issue.Title != "Fix login bug" {
			t.Errorf("title: got %q, want %q", issue.Title, "Fix login bug")
		}
		if issue.State != "open" {
			t.Errorf("state: got %q, want %q", issue.State, "open")
		}
		if issue.Body != "The login form crashes." {
			t.Errorf("body: got %q", issue.Body)
		}
		if issue.Author != "octocat" {
			t.Errorf("author: got %q, want %q", issue.Author, "octocat")
		}
		if len(issue.Labels) != 2 {
			t.Fatalf("labels: got %v, want 2 items", issue.Labels)
		}
		if issue.Labels[0] != "bug" || issue.Labels[1] != "priority/high" {
			t.Errorf("labels: got %v", issue.Labels)
		}
		if len(issue.Assignees) != 1 || issue.Assignees[0] != "alice" {
			t.Errorf("assignees: got %v", issue.Assignees)
		}
		if issue.Milestone != "v1.0" {
			t.Errorf("milestone: got %q, want %q", issue.Milestone, "v1.0")
		}
		if issue.CreatedAt != (time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)) {
			t.Errorf("created_at: got %v", issue.CreatedAt)
		}
		if issue.UpdatedAt != (time.Date(2026, 3, 20, 14, 22, 0, 0, time.UTC)) {
			t.Errorf("updated_at: got %v", issue.UpdatedAt)
		}
	})

	t.Run("nil milestone", func(t *testing.T) {
		ai := &ghAPIIssue{
			Number:    1,
			Title:     "Test",
			State:     "open",
			CreatedAt: "2026-01-01T00:00:00Z",
			UpdatedAt: "2026-01-01T00:00:00Z",
		}
		ai.User.Login = "bob"

		issue, err := apiIssueToIssue(ai)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if issue.Milestone != "" {
			t.Errorf("milestone: got %q, want empty", issue.Milestone)
		}
	})

	t.Run("empty labels and assignees", func(t *testing.T) {
		ai := &ghAPIIssue{
			Number:    2,
			Title:     "Test",
			State:     "closed",
			CreatedAt: "2026-01-01T00:00:00Z",
			UpdatedAt: "2026-01-01T00:00:00Z",
		}
		ai.User.Login = "bob"

		issue, err := apiIssueToIssue(ai)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if issue.Labels != nil {
			t.Errorf("labels: got %v, want nil", issue.Labels)
		}
		if issue.Assignees != nil {
			t.Errorf("assignees: got %v, want nil", issue.Assignees)
		}
	})

	t.Run("invalid created_at", func(t *testing.T) {
		ai := &ghAPIIssue{
			Number:    3,
			Title:     "Test",
			State:     "open",
			CreatedAt: "not-a-date",
			UpdatedAt: "2026-01-01T00:00:00Z",
		}
		_, err := apiIssueToIssue(ai)
		if err == nil {
			t.Error("expected error for invalid created_at")
		}
	})

	t.Run("invalid updated_at", func(t *testing.T) {
		ai := &ghAPIIssue{
			Number:    3,
			Title:     "Test",
			State:     "open",
			CreatedAt: "2026-01-01T00:00:00Z",
			UpdatedAt: "not-a-date",
		}
		_, err := apiIssueToIssue(ai)
		if err == nil {
			t.Error("expected error for invalid updated_at")
		}
	})
}
