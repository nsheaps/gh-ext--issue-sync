package frontmatter

import (
	"bytes"
	"strings"
	"testing"
)

type testMeta struct {
	Title  string   `yaml:"title"`
	Number int      `yaml:"number"`
	Labels []string `yaml:"labels,omitempty"`
}

func TestMarshalUnmarshal(t *testing.T) {
	meta := testMeta{
		Title:  "Test Issue",
		Number: 42,
		Labels: []string{"bug", "priority/high"},
	}
	body := "This is the issue body.\n\nWith multiple paragraphs."

	var buf bytes.Buffer
	if err := Marshal(&buf, &meta, body); err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	output := buf.String()

	// Verify structure
	if !strings.HasPrefix(output, "---\n") {
		t.Error("output should start with ---")
	}
	if !strings.Contains(output, "title: Test Issue") {
		t.Error("output should contain title")
	}
	if !strings.Contains(output, "number: 42") {
		t.Error("output should contain number")
	}
	if !strings.Contains(output, "This is the issue body.") {
		t.Error("output should contain body")
	}

	// Round-trip
	var parsed testMeta
	parsedBody, err := Unmarshal(strings.NewReader(output), &parsed)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if parsed.Title != meta.Title {
		t.Errorf("title: got %q, want %q", parsed.Title, meta.Title)
	}
	if parsed.Number != meta.Number {
		t.Errorf("number: got %d, want %d", parsed.Number, meta.Number)
	}
	if len(parsed.Labels) != len(meta.Labels) {
		t.Errorf("labels: got %v, want %v", parsed.Labels, meta.Labels)
	}
	if strings.TrimSpace(parsedBody) != body {
		t.Errorf("body: got %q, want %q", strings.TrimSpace(parsedBody), body)
	}
}

func TestUnmarshalNoFrontmatter(t *testing.T) {
	input := "Just plain markdown content."
	var meta testMeta
	body, err := Unmarshal(strings.NewReader(input), &meta)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if body != input {
		t.Errorf("body: got %q, want %q", body, input)
	}
}

func TestUnmarshalUnclosedDelimiter(t *testing.T) {
	input := "---\ntitle: test\nno closing delimiter"
	var meta testMeta
	_, err := Unmarshal(strings.NewReader(input), &meta)
	if err == nil {
		t.Error("expected error for unclosed delimiter")
	}
}
