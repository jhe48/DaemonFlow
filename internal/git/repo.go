package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Repo represents a Git repository
type Repo struct {
	path string // Root directory of the repository
}

// NewRepo creates a new Repo instance for the given path
func NewRepo(path string) *Repo {
	return &Repo{path: path}
}

// Path returns the root directory of the repository
func (r *Repo) Path() string {
	return r.path
}

// IsGitRepo checks if the path is a Git repository
func (r *Repo) IsGitRepo() bool {
	gitDir := filepath.Join(r.path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	// .git can be a directory (normal) or a file (worktree/submodule)
	return info.IsDir() || info.Mode().IsRegular()
}

// GetHEAD returns the current HEAD commit hash
func (r *Repo) GetHEAD() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetBranch returns the current branch name
// Returns empty string if in detached HEAD state
func (r *Repo) GetBranch() (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		// Check if we're in detached HEAD state
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 128 {
			// Detached HEAD - not an error, just return empty
			return "", nil
		}
		return "", fmt.Errorf("failed to get branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetLastCommitHash returns the hash of the last commit
// This is the same as GetHEAD but named differently for clarity
func (r *Repo) GetLastCommitHash() (string, error) {
	return r.GetHEAD()
}

// GetLastCommitTime returns the timestamp of the last commit
func (r *Repo) GetLastCommitTime() (time.Time, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%ct")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last commit time: %w", err)
	}

	timestamp, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse commit timestamp: %w", err)
	}

	return time.Unix(timestamp, 0), nil
}

// GetLastCommitMessage returns the subject line of the last commit message
func (r *Repo) GetLastCommitMessage() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get last commit message: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// FindGitRoot walks up from the given path to find the Git repository root
// Returns the repository root path or an error if not in a Git repository
func FindGitRoot(path string) (string, error) {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Walk up the directory tree
	current := absPath
	for {
		gitDir := filepath.Join(current, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return current, nil
		}

		// Move to parent directory
		parent := filepath.Dir(current)
		if parent == current {
			// Reached root, no .git found
			return "", fmt.Errorf("not a git repository (or any parent up to mount point %s)", absPath)
		}
		current = parent
	}
}
