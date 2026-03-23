package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nsheaps/gh-ext--issue-sync/internal/sync"
)

// mockClient implements sync.Client for testing.
type mockClient struct {
	repo   string
	issues map[int]*sync.Issue
	// pushCalls tracks which issues were pushed (number -> issue)
	pushCalls map[int]*sync.Issue
	// errors to inject
	resolveErr  error
	fetchErr    error
	fetchAllErr error
	pushErr     error
}

func newMockClient(repo string) *mockClient {
	return &mockClient{
		repo:      repo,
		issues:    make(map[int]*sync.Issue),
		pushCalls: make(map[int]*sync.Issue),
	}
}

func (m *mockClient) ResolveRepo() (string, error) {
	if m.resolveErr != nil {
		return "", m.resolveErr
	}
	return m.repo, nil
}

func (m *mockClient) FetchIssue(_ string, number int) (*sync.Issue, error) {
	if m.fetchErr != nil {
		return nil, m.fetchErr
	}
	issue, ok := m.issues[number]
	if !ok {
		return nil, fmt.Errorf("issue #%d not found", number)
	}
	return issue, nil
}

func (m *mockClient) FetchAllIssues(_ string, state string) ([]*sync.Issue, error) {
	if m.fetchAllErr != nil {
		return nil, m.fetchAllErr
	}
	var result []*sync.Issue
	for _, issue := range m.issues {
		if state == "all" || issue.State == state {
			result = append(result, issue)
		}
	}
	return result, nil
}

func (m *mockClient) PushIssue(_ string, issue *sync.Issue) error {
	if m.pushErr != nil {
		return m.pushErr
	}
	m.pushCalls[issue.Number] = issue
	return nil
}

func (m *mockClient) addIssue(issue *sync.Issue) {
	m.issues[issue.Number] = issue
}

var testTime = time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

func makeIssue(number int, title, state, body string) *sync.Issue {
	return &sync.Issue{
		Number:    number,
		Title:     title,
		State:     state,
		Body:      body,
		Labels:    []string{"bug"},
		Assignees: []string{"alice"},
		CreatedAt: testTime,
		UpdatedAt: testTime,
		Author:    "octocat",
	}
}

// setupTestClient installs a mock client and returns it with cleanup.
func setupTestClient(t *testing.T) *mockClient {
	t.Helper()
	mc := newMockClient("owner/repo")
	oldClient := ghClient
	ghClient = mc
	t.Cleanup(func() { ghClient = oldClient })
	return mc
}

// --- Pull tests ---

func TestPullSingleIssue(t *testing.T) {
	mc := setupTestClient(t)
	mc.addIssue(makeIssue(42, "Fix login bug", "open", "The login form crashes."))

	dir := t.TempDir()
	pullDir = dir
	pullAll = false

	err := pullCmd.RunE(pullCmd, []string{"42"})
	if err != nil {
		t.Fatalf("pull: %v", err)
	}

	// Verify file was created
	filename := filepath.Join(dir, "00042.md")
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "title: Fix login bug") {
		t.Error("file should contain title")
	}
	if !strings.Contains(content, "number: 42") {
		t.Error("file should contain number")
	}
	if !strings.Contains(content, "The login form crashes.") {
		t.Error("file should contain body")
	}
}

func TestPullAllIssues(t *testing.T) {
	mc := setupTestClient(t)
	mc.addIssue(makeIssue(1, "First issue", "open", "Body 1"))
	mc.addIssue(makeIssue(2, "Second issue", "open", "Body 2"))

	dir := t.TempDir()
	pullDir = dir
	pullAll = true
	pullState = "open"

	err := pullCmd.RunE(pullCmd, []string{})
	if err != nil {
		t.Fatalf("pull --all: %v", err)
	}

	// Verify both files exist
	for _, num := range []int{1, 2} {
		filename := filepath.Join(dir, fmt.Sprintf("%05d.md", num))
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", filename)
		}
	}
}

func TestPullInvalidNumber(t *testing.T) {
	setupTestClient(t)
	pullAll = false

	err := pullCmd.RunE(pullCmd, []string{"abc"})
	if err == nil {
		t.Error("expected error for non-numeric issue number")
	}
}

func TestPullNegativeNumber(t *testing.T) {
	setupTestClient(t)
	pullAll = false

	err := pullCmd.RunE(pullCmd, []string{"-1"})
	if err == nil {
		t.Error("expected error for negative issue number")
	}
}

func TestPullNoArgs(t *testing.T) {
	setupTestClient(t)
	pullAll = false

	err := pullCmd.RunE(pullCmd, []string{})
	if err == nil {
		t.Error("expected error when no args and no --all")
	}
}

// --- Push tests ---

func TestPushSingleIssue(t *testing.T) {
	mc := setupTestClient(t)
	mc.addIssue(makeIssue(42, "Fix login bug", "open", "Original body"))

	dir := t.TempDir()
	pushDir = dir

	// Write a file to push
	issue := makeIssue(42, "Updated title", "closed", "Updated body")
	if err := writeIssueFile(issue, dir); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	err := pushCmd.RunE(pushCmd, []string{"42"})
	if err != nil {
		t.Fatalf("push: %v", err)
	}

	pushed, ok := mc.pushCalls[42]
	if !ok {
		t.Fatal("issue #42 was not pushed")
	}
	if pushed.Title != "Updated title" {
		t.Errorf("title: got %q, want %q", pushed.Title, "Updated title")
	}
	if pushed.State != "closed" {
		t.Errorf("state: got %q, want %q", pushed.State, "closed")
	}
	if pushed.Body != "Updated body" {
		t.Errorf("body: got %q, want %q", pushed.Body, "Updated body")
	}
}

func TestPushAllIssues(t *testing.T) {
	mc := setupTestClient(t)

	dir := t.TempDir()
	pushDir = dir
	pushAll = true

	// Write two files
	for i := 1; i <= 2; i++ {
		issue := makeIssue(i, fmt.Sprintf("Issue %d", i), "open", fmt.Sprintf("Body %d", i))
		if err := writeIssueFile(issue, dir); err != nil {
			t.Fatalf("writing test file: %v", err)
		}
	}

	err := pushCmd.RunE(pushCmd, []string{})
	if err != nil {
		t.Fatalf("push --all: %v", err)
	}

	if len(mc.pushCalls) != 2 {
		t.Errorf("expected 2 push calls, got %d", len(mc.pushCalls))
	}
}

func TestPushMissingFile(t *testing.T) {
	setupTestClient(t)

	pushDir = t.TempDir()
	pushAll = false

	err := pushCmd.RunE(pushCmd, []string{"999"})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestPushNoArgs(t *testing.T) {
	setupTestClient(t)
	pushAll = false

	err := pushCmd.RunE(pushCmd, []string{})
	if err == nil {
		t.Error("expected error when no args and no --all")
	}
}

// --- Status tests ---

func TestStatusShowsModified(t *testing.T) {
	mc := setupTestClient(t)
	mc.addIssue(makeIssue(42, "Remote title", "open", "Remote body"))

	dir := t.TempDir()
	statusDir = dir

	// Write a file with different title
	issue := makeIssue(42, "Local title", "open", "Remote body")
	if err := writeIssueFile(issue, dir); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runStatus("owner/repo")

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("status: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "M #42") {
		t.Errorf("expected modified marker, got: %s", output)
	}
	if !strings.Contains(output, "title") {
		t.Errorf("expected 'title' in changes, got: %s", output)
	}
}

func TestStatusShowsUnchanged(t *testing.T) {
	mc := setupTestClient(t)
	issue := makeIssue(42, "Same title", "open", "Same body")
	mc.addIssue(issue)

	dir := t.TempDir()
	statusDir = dir

	if err := writeIssueFile(issue, dir); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runStatus("owner/repo")

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("status: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "= #42") {
		t.Errorf("expected unchanged marker, got: %s", output)
	}
}

func TestStatusDetectsMilestoneChange(t *testing.T) {
	mc := setupTestClient(t)
	remote := makeIssue(42, "Title", "open", "Body")
	remote.Milestone = "v2.0"
	mc.addIssue(remote)

	dir := t.TempDir()
	statusDir = dir

	local := makeIssue(42, "Title", "open", "Body")
	local.Milestone = "v1.0"
	if err := writeIssueFile(local, dir); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runStatus("owner/repo")

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("status: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "milestone") {
		t.Errorf("expected milestone change detected, got: %s", output)
	}
}

func TestStatusEmptyDir(t *testing.T) {
	setupTestClient(t)

	statusDir = t.TempDir()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runStatus("owner/repo")

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("status: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "No local issue files") {
		t.Errorf("expected no-files message, got: %s", output)
	}
}

// --- stringSliceEqualUnordered tests ---

func TestStringSliceEqualUnordered(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
		want bool
	}{
		{"both nil", nil, nil, true},
		{"both empty", []string{}, []string{}, true},
		{"same order", []string{"a", "b"}, []string{"a", "b"}, true},
		{"different order", []string{"b", "a"}, []string{"a", "b"}, true},
		{"different length", []string{"a"}, []string{"a", "b"}, false},
		{"different values", []string{"a", "b"}, []string{"a", "c"}, false},
		{"one nil one empty", nil, []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringSliceEqualUnordered(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("stringSliceEqualUnordered(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// --- readIssueFile tests ---

func TestReadIssueFileValidation(t *testing.T) {
	t.Run("missing number", func(t *testing.T) {
		dir := t.TempDir()
		filename := filepath.Join(dir, "00001.md")
		// Write a file with number: 0 (invalid)
		if err := os.WriteFile(filename, []byte("---\nnumber: 0\ntitle: Test\nstate: open\ncreated_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\nauthor: bob\n---\n\nBody"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := readIssueFile(filename)
		if err == nil {
			t.Error("expected error for zero issue number")
		}
	})

	t.Run("missing title", func(t *testing.T) {
		dir := t.TempDir()
		filename := filepath.Join(dir, "00001.md")
		if err := os.WriteFile(filename, []byte("---\nnumber: 1\ntitle: \"\"\nstate: open\ncreated_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\nauthor: bob\n---\n\nBody"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := readIssueFile(filename)
		if err == nil {
			t.Error("expected error for empty title")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := readIssueFile("/nonexistent/00001.md")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})
}
