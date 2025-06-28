package markdown

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/tkan/mdtask/internal/task"
	"gopkg.in/yaml.v3"
)

func WriteTaskFile(t *task.Task) ([]byte, error) {
	fm := FrontMatter{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Aliases:     t.Aliases,
		Tags:        t.Tags,
		Created:     t.Created,
		Updated:     t.Updated,
	}

	yamlData, err := yaml.Marshal(&fm)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal front matter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlData)
	buf.WriteString("---\n\n")
	
	content := strings.TrimSpace(t.Content)
	if content != "" {
		buf.WriteString(content)
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

func GenerateTaskID() string {
	now := time.Now()
	return fmt.Sprintf("task/%s", now.Format("20060102150405"))
}