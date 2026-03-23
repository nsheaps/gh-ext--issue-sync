package sync

import "time"

// Issue represents a GitHub issue with all syncable metadata.
type Issue struct {
	Number    int       `yaml:"number"`
	Title     string    `yaml:"title"`
	State     string    `yaml:"state"`
	Labels    []string  `yaml:"labels,omitempty"`
	Assignees []string  `yaml:"assignees,omitempty"`
	Milestone string    `yaml:"milestone,omitempty"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
	Author    string    `yaml:"author"`
	Body      string    `yaml:"-"` // stored as markdown content, not in frontmatter
}

// IssueFrontmatter is the YAML frontmatter written to issue files.
// Separate from Issue to control serialization independently.
type IssueFrontmatter struct {
	Number    int       `yaml:"number"`
	Title     string    `yaml:"title"`
	State     string    `yaml:"state"`
	Labels    []string  `yaml:"labels,omitempty"`
	Assignees []string  `yaml:"assignees,omitempty"`
	Milestone string    `yaml:"milestone,omitempty"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
	Author    string    `yaml:"author"`
}
