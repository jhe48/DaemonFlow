package watcher

import (
	"path/filepath"
	"strings"
	"sync"
)

// DefaultIgnorePatterns contains patterns that are always ignored
var DefaultIgnorePatterns = []string{
	".git",
	".git/**",
	"node_modules",
	"node_modules/**",
	"*.swp",
	"*.swo",
	"*~",
	".DS_Store",
	"__pycache__",
	"__pycache__/**",
	"*.pyc",
	".idea",
	".idea/**",
	".vscode",
	".vscode/**",
	"*.log",
	"dist",
	"dist/**",
	"build",
	"build/**",
	"target",
	"target/**",
	"vendor",
	"vendor/**",
}

// IgnoreMatcher determines if file paths should be ignored
type IgnoreMatcher struct {
	patterns []string
	mu       sync.RWMutex
}

// NewIgnoreMatcher creates a new IgnoreMatcher with default patterns
func NewIgnoreMatcher() *IgnoreMatcher {
	return &IgnoreMatcher{
		patterns: append([]string{}, DefaultIgnorePatterns...),
	}
}

// NewIgnoreMatcherWithPatterns creates an IgnoreMatcher with additional patterns
func NewIgnoreMatcherWithPatterns(additionalPatterns []string) *IgnoreMatcher {
	im := NewIgnoreMatcher()
	for _, p := range additionalPatterns {
		im.AddPattern(p)
	}
	return im
}

// AddPattern adds a custom pattern to the matcher
func (im *IgnoreMatcher) AddPattern(pattern string) {
	if pattern == "" {
		return
	}
	im.mu.Lock()
	defer im.mu.Unlock()
	im.patterns = append(im.patterns, pattern)
}

// Match returns true if the path should be ignored
// The path should be relative to the watch root
func (im *IgnoreMatcher) Match(path string) bool {
	if path == "" {
		return false
	}

	im.mu.RLock()
	defer im.mu.RUnlock()

	// Normalize path separators
	path = filepath.ToSlash(path)

	for _, pattern := range im.patterns {
		if im.matchPattern(pattern, path) {
			return true
		}
	}
	return false
}

// matchPattern checks if a path matches a single pattern
func (im *IgnoreMatcher) matchPattern(pattern, path string) bool {
	// Normalize pattern separators
	pattern = filepath.ToSlash(pattern)

	// Handle ** for recursive directory matching
	if strings.Contains(pattern, "**") {
		return im.matchDoubleGlob(pattern, path)
	}

	// Check if pattern matches any segment of the path
	// This allows patterns like ".git" to match ".git" at any level
	pathSegments := strings.Split(path, "/")

	// First, try matching against the full path
	if matched, _ := filepath.Match(pattern, path); matched {
		return true
	}

	// Then try matching against each path segment
	for _, segment := range pathSegments {
		if matched, _ := filepath.Match(pattern, segment); matched {
			return true
		}
	}

	// Also try matching against the base name
	baseName := filepath.Base(path)
	if matched, _ := filepath.Match(pattern, baseName); matched {
		return true
	}

	return false
}

// matchDoubleGlob handles patterns containing **
func (im *IgnoreMatcher) matchDoubleGlob(pattern, path string) bool {
	// Split pattern by **
	parts := strings.Split(pattern, "**")

	if len(parts) == 2 {
		prefix := strings.TrimSuffix(parts[0], "/")
		suffix := strings.TrimPrefix(parts[1], "/")

		// Pattern like "dir/**" matches dir and anything under it
		if suffix == "" {
			// Check if path starts with prefix or is the prefix
			if prefix == "" {
				return true // ** alone matches everything
			}
			if path == prefix {
				return true
			}
			if strings.HasPrefix(path, prefix+"/") {
				return true
			}
		}

		// Pattern like "**/file" matches file in any directory
		if prefix == "" {
			// Check if path ends with suffix or matches suffix
			if matched, _ := filepath.Match(suffix, filepath.Base(path)); matched {
				return true
			}
			if strings.HasSuffix(path, "/"+suffix) {
				return true
			}
			if path == suffix {
				return true
			}
		}

		// Pattern like "dir/**/file" matches file under dir at any depth
		if prefix != "" && suffix != "" {
			if strings.HasPrefix(path, prefix+"/") || path == prefix {
				remaining := strings.TrimPrefix(path, prefix+"/")
				if matched, _ := filepath.Match(suffix, filepath.Base(remaining)); matched {
					return true
				}
				if remaining == suffix {
					return true
				}
				if strings.HasSuffix(remaining, "/"+suffix) {
					return true
				}
			}
		}
	}

	return false
}

// Patterns returns a copy of the current patterns
func (im *IgnoreMatcher) Patterns() []string {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return append([]string{}, im.patterns...)
}
