package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initTestRepo creates a temporary git repo with an initial commit.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("setup %v: %v\n%s", args, err, out)
		}
	}

	// Create and commit an initial file.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "README.md"},
		{"git", "commit", "-m", "initial"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("setup %v: %v\n%s", args, err, out)
		}
	}

	return dir
}

func TestGetDiff_NoChanges(t *testing.T) {
	dir := initTestRepo(t)

	diff, err := GetDiff(dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !diff.IsEmpty() {
		t.Errorf("expected empty diff, got staged=%q unstaged=%q", diff.Staged, diff.Unstaged)
	}
}

func TestGetDiff_UnstagedChanges(t *testing.T) {
	dir := initTestRepo(t)

	// Modify a tracked file without staging.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# updated\n"), 0644); err != nil {
		t.Fatal(err)
	}

	diff, err := GetDiff(dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff.Unstaged == "" {
		t.Error("expected unstaged diff, got empty")
	}
	if diff.Staged != "" {
		t.Errorf("expected empty staged diff, got %q", diff.Staged)
	}
	if diff.IsEmpty() {
		t.Error("expected non-empty diff")
	}
}

func TestGetDiff_StagedChanges(t *testing.T) {
	dir := initTestRepo(t)

	// Stage a modification.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# staged\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "README.md")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

	diff, err := GetDiff(dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff.Staged == "" {
		t.Error("expected staged diff, got empty")
	}
}

func TestGetDiff_StagedOnly(t *testing.T) {
	dir := initTestRepo(t)

	// Stage a change and also have an unstaged change.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# staged\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "README.md")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# staged then modified\n"), 0644); err != nil {
		t.Fatal(err)
	}

	diff, err := GetDiff(dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff.Staged == "" {
		t.Error("expected staged diff, got empty")
	}
	if diff.Unstaged != "" {
		t.Errorf("expected empty unstaged diff with staged-only, got %q", diff.Unstaged)
	}
}

func TestGetDiff_Combined(t *testing.T) {
	dir := initTestRepo(t)

	// Create both staged and unstaged changes.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# staged\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "README.md")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new file\n"), 0644); err != nil {
		t.Fatal(err)
	}

	diff, err := GetDiff(dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	combined := diff.Combined()
	if combined == "" {
		t.Error("expected non-empty combined diff")
	}
	if diff.Staged == "" {
		t.Error("expected staged diff")
	}
	// new.txt is untracked, not modified, so unstaged may be empty.
	// That's fine - this test verifies Combined() works.
}

func TestGetDiff_InvalidDir(t *testing.T) {
	_, err := GetDiff("/nonexistent/path", false)
	if err == nil {
		t.Error("expected error for invalid directory")
	}
}

func TestDiffResult_Combined_Empty(t *testing.T) {
	d := DiffResult{}
	if d.Combined() != "" {
		t.Errorf("expected empty combined, got %q", d.Combined())
	}
}

func TestDiffResult_Combined_StagedOnly(t *testing.T) {
	d := DiffResult{Staged: "staged content"}
	if d.Combined() != "staged content" {
		t.Errorf("expected 'staged content', got %q", d.Combined())
	}
}

func TestDiffResult_Combined_Both(t *testing.T) {
	d := DiffResult{Staged: "staged", Unstaged: "unstaged"}
	combined := d.Combined()
	if combined != "staged\nunstaged" {
		t.Errorf("expected 'staged\\nunstaged', got %q", combined)
	}
}
