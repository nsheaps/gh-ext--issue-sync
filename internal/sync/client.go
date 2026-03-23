package sync

// Client abstracts GitHub operations so they can be mocked in tests.
type Client interface {
	// ResolveRepo determines the current GitHub repo (e.g., "owner/repo").
	ResolveRepo() (string, error)
	// FetchIssue fetches a single issue by number.
	FetchIssue(repo string, number int) (*Issue, error)
	// FetchAllIssues fetches issues matching the given state filter.
	FetchAllIssues(repo string, state string) ([]*Issue, error)
	// PushIssue updates a GitHub issue from local data.
	PushIssue(repo string, issue *Issue) error
}
