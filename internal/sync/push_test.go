package sync

import "testing"

func TestPushIssueViaGH_Validation(t *testing.T) {
	t.Run("invalid number", func(t *testing.T) {
		err := pushIssueViaGH("owner/repo", &Issue{Number: 0, Title: "Test", State: "open"})
		if err == nil {
			t.Error("expected error for zero issue number")
		}
	})

	t.Run("negative number", func(t *testing.T) {
		err := pushIssueViaGH("owner/repo", &Issue{Number: -1, Title: "Test", State: "open"})
		if err == nil {
			t.Error("expected error for negative issue number")
		}
	})

	t.Run("empty title", func(t *testing.T) {
		err := pushIssueViaGH("owner/repo", &Issue{Number: 1, Title: "", State: "open"})
		if err == nil {
			t.Error("expected error for empty title")
		}
	})

	t.Run("invalid state", func(t *testing.T) {
		err := pushIssueViaGH("owner/repo", &Issue{Number: 1, Title: "Test", State: "invalid"})
		if err == nil {
			t.Error("expected error for invalid state")
		}
	})

	t.Run("empty state", func(t *testing.T) {
		err := pushIssueViaGH("owner/repo", &Issue{Number: 1, Title: "Test", State: ""})
		if err == nil {
			t.Error("expected error for empty state")
		}
	})
}
