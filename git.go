package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// DiffResult holds the combined diff output from a git repository.
type DiffResult struct {
	Staged     string
	Unstaged   string
	Untracked  string
}

// Combined returns the full diff text (staged + unstaged + untracked).
func (d DiffResult) Combined() string {
	parts := make([]string, 0, 3)
	if d.Staged != "" {
		parts = append(parts, d.Staged)
	}
	if d.Unstaged != "" {
		parts = append(parts, d.Unstaged)
	}
	if d.Untracked != "" {
		parts = append(parts, d.Untracked)
	}
	return strings.Join(parts, "\n")
}

// IsEmpty returns true if there are no changes.
func (d DiffResult) IsEmpty() bool {
	return d.Staged == "" && d.Unstaged == "" && d.Untracked == ""
}

// GetDiff extracts git diffs from the given directory.
// If includeAll is false (default), only staged changes are returned; if there
// are no staged changes, it falls back to all changes (unstaged + untracked).
// If includeAll is true, all changes are always returned regardless of staging.
func GetDiff(dir string, includeAll bool) (DiffResult, error) {
	var result DiffResult

	staged, err := runGit(dir, "diff", "--cached")
	if err != nil {
		return result, fmt.Errorf("getting staged diff: %w", err)
	}
	result.Staged = strings.TrimSpace(staged)

	// Include unstaged/untracked when explicitly requested OR when there are
	// no staged changes (fallback so the tool is always useful).
	if includeAll || result.Staged == "" {
		unstaged, err := runGit(dir, "diff")
		if err != nil {
			return result, fmt.Errorf("getting unstaged diff: %w", err)
		}
		result.Unstaged = strings.TrimSpace(unstaged)

		untracked, err := getUntrackedDiff(dir)
		if err != nil {
			return result, fmt.Errorf("getting untracked files: %w", err)
		}
		result.Untracked = strings.TrimSpace(untracked)
	}

	return result, nil
}

// getUntrackedDiff returns a diff-like representation of untracked files.
func getUntrackedDiff(dir string) (string, error) {
	// Get list of untracked files
	files, err := runGit(dir, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return "", err
	}

	files = strings.TrimSpace(files)
	if files == "" {
		return "", nil
	}

	var diffs []string
	for _, file := range strings.Split(files, "\n") {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}

		// Use git diff --no-index with null device to show as new file
		cmd := exec.Command("git", "diff", "--no-index", "/dev/null", file)
		cmd.Dir = dir
		out, _ := cmd.CombinedOutput()

		if len(out) > 0 {
			diffs = append(diffs, string(out))
		}
	}

	return strings.Join(diffs, "\n"), nil
}

// runGit executes a git command in the given directory and returns stdout.
func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), string(exitErr.Stderr))
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}
