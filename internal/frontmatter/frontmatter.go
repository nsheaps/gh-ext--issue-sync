// Package frontmatter provides generic YAML frontmatter serialization for
// markdown files. It reads and writes the "---" delimited YAML header followed
// by markdown body content.
package frontmatter

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

const delimiter = "---"

// Marshal writes YAML frontmatter followed by markdown body content.
func Marshal(w io.Writer, meta interface{}, body string) error {
	yamlBytes, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshaling frontmatter: %w", err)
	}

	if _, err := fmt.Fprintf(w, "%s\n%s%s\n\n%s", delimiter, yamlBytes, delimiter, body); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}

// Unmarshal reads YAML frontmatter and markdown body from a reader.
func Unmarshal(r io.Reader, meta interface{}) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}

	content := bytes.TrimSpace(data)
	if !bytes.HasPrefix(content, []byte(delimiter)) {
		return string(content), nil // no frontmatter
	}

	// Find the closing delimiter
	rest := content[len(delimiter):]
	rest = bytes.TrimLeft(rest, "\n")
	idx := bytes.Index(rest, []byte("\n"+delimiter))
	if idx < 0 {
		return "", fmt.Errorf("unclosed frontmatter delimiter")
	}

	yamlData := rest[:idx]
	body := rest[idx+len("\n"+delimiter):]
	body = bytes.TrimLeft(body, "\n")

	if err := yaml.Unmarshal(yamlData, meta); err != nil {
		return "", fmt.Errorf("parsing frontmatter YAML: %w", err)
	}

	return string(body), nil
}
