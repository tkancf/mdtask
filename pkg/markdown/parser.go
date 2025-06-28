package markdown

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/tkan/mdtask/internal/task"
	"gopkg.in/yaml.v3"
)

type FrontMatter struct {
	ID          string   `yaml:"id"`
	Aliases     []string `yaml:"aliases"`
	Tags        []string `yaml:"tags"`
	Created     string   `yaml:"created"`
	Description string   `yaml:"description"`
	Title       string   `yaml:"title"`
	Updated     string   `yaml:"updated"`
}

func ParseTaskFile(content []byte) (*task.Task, error) {
	frontMatter, body, err := extractFrontMatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract front matter: %w", err)
	}

	var fm FrontMatter
	if err := yaml.Unmarshal([]byte(frontMatter), &fm); err != nil {
		return nil, fmt.Errorf("failed to parse YAML front matter: %w", err)
	}

	// Parse time strings
	created, err := time.Parse("2006-01-02 15:04", fm.Created)
	if err != nil {
		// Try without time
		created, err = time.Parse("2006-01-02", fm.Created)
		if err != nil {
			created = time.Now()
		}
	}
	
	updated, err := time.Parse("2006-01-02 15:04", fm.Updated)
	if err != nil {
		// Try without time
		updated, err = time.Parse("2006-01-02", fm.Updated)
		if err != nil {
			updated = time.Now()
		}
	}

	t := &task.Task{
		ID:          fm.ID,
		Title:       fm.Title,
		Description: fm.Description,
		Aliases:     fm.Aliases,
		Tags:        fm.Tags,
		Created:     created,
		Updated:     updated,
		Content:     body,
	}

	return t, nil
}

func extractFrontMatter(content []byte) (string, string, error) {
	lines := bytes.Split(content, []byte("\n"))
	
	if len(lines) < 3 || !bytes.Equal(lines[0], []byte("---")) {
		return "", "", fmt.Errorf("no front matter found")
	}

	endIndex := -1
	for i := 1; i < len(lines); i++ {
		if bytes.Equal(lines[i], []byte("---")) {
			endIndex = i
			break
		}
	}

	if endIndex == -1 {
		return "", "", fmt.Errorf("front matter not properly closed")
	}

	frontMatterLines := lines[1:endIndex]
	frontMatter := strings.Join(bytesToStrings(frontMatterLines), "\n")

	bodyLines := lines[endIndex+1:]
	body := strings.Join(bytesToStrings(bodyLines), "\n")
	body = strings.TrimLeft(body, "\n")

	return frontMatter, body, nil
}

func bytesToStrings(byteSlices [][]byte) []string {
	result := make([]string, len(byteSlices))
	for i, b := range byteSlices {
		result[i] = string(b)
	}
	return result
}