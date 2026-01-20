package task

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// Task represents a single task item from a markdown file
type Task struct {
	// Line is the line number in the file (1-indexed)
	Line int

	// Text is the task text without the checkbox prefix
	Text string

	// Completed indicates whether the checkbox is checked
	Completed bool
}

// checkboxRegex matches markdown checkbox patterns
// Supports: - [ ] unchecked, - [x] or - [X] checked
// Group 1: space = unchecked, x/X = checked
// Group 2: task text
var checkboxRegex = regexp.MustCompile(`^\s*-\s*\[([ xX])\]\s*(.+)$`)

// ParseFile reads a file and extracts all task items
func ParseFile(path string) ([]Task, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tasks []Task
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if task, ok := parseLine(line, lineNum); ok {
			tasks = append(tasks, task)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// ParseContent extracts task items from a string content
// Useful for testing and comparing old vs new state
func ParseContent(content string) []Task {
	var tasks []Task
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		lineNum := i + 1 // 1-indexed
		if task, ok := parseLine(line, lineNum); ok {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// parseLine attempts to parse a single line as a task
// Returns the task and true if successful, or zero value and false if not a task line
func parseLine(line string, lineNum int) (Task, bool) {
	matches := checkboxRegex.FindStringSubmatch(line)
	if matches == nil {
		return Task{}, false
	}

	// matches[0] = full match
	// matches[1] = checkbox state (space, x, or X)
	// matches[2] = task text

	checkState := matches[1]
	text := strings.TrimSpace(matches[2])

	// Skip empty task text
	if text == "" {
		return Task{}, false
	}

	completed := checkState == "x" || checkState == "X"

	return Task{
		Line:      lineNum,
		Text:      text,
		Completed: completed,
	}, true
}
