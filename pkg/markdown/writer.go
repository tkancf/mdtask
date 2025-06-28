package markdown

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tkan/mdtask/internal/constants"
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
		Created:     t.Created.Format(constants.DateTimeFormat),
		Updated:     t.Updated.Format(constants.DateTimeFormat),
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

var lastGeneratedTime time.Time
var mu sync.Mutex

func GenerateTaskID() string {
	mu.Lock()
	defer mu.Unlock()
	
	now := time.Now()
	// If generating in the same second, wait a bit
	if now.Format(constants.IDTimeFormat) == lastGeneratedTime.Format(constants.IDTimeFormat) {
		time.Sleep(constants.GenerateIDSleepDuration)
		now = time.Now()
	}
	lastGeneratedTime = now
	
	return fmt.Sprintf("%s%s", constants.TaskIDPrefix, now.Format(constants.IDTimeFormat))
}