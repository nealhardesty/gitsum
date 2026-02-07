package main

import (
	"strings"
	"testing"
)

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name          string
		diff          string
		wantTruncated bool
		wantContains  string
	}{
		{
			name:          "normal diff",
			diff:          "diff --git a/foo.go b/foo.go\n+added line",
			wantTruncated: false,
			wantContains:  "diff --git a/foo.go b/foo.go",
		},
		{
			name:          "empty diff",
			diff:          "",
			wantTruncated: false,
			wantContains:  "Diff:\n",
		},
		{
			name:          "truncated diff",
			diff:          strings.Repeat("x", MaxDiffChars+100),
			wantTruncated: true,
		},
		{
			name:          "exactly at limit",
			diff:          strings.Repeat("y", MaxDiffChars),
			wantTruncated: false,
			wantContains:  strings.Repeat("y", 100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, truncated := BuildPrompt(tt.diff)

			if truncated != tt.wantTruncated {
				t.Errorf("truncated = %v, want %v", truncated, tt.wantTruncated)
			}

			if tt.wantContains != "" && !strings.Contains(prompt, tt.wantContains) {
				t.Errorf("prompt does not contain %q", tt.wantContains)
			}

			// Verify the prompt always contains the template instructions.
			if !strings.Contains(prompt, "imperative mood") {
				t.Error("prompt missing imperative mood instruction")
			}

			if tt.wantTruncated {
				// The diff portion should be exactly MaxDiffChars.
				idx := strings.Index(prompt, "Diff:\n")
				if idx == -1 {
					t.Fatal("prompt missing 'Diff:' marker")
				}
				diffPortion := prompt[idx+len("Diff:\n"):]
				if len(diffPortion) != MaxDiffChars {
					t.Errorf("truncated diff length = %d, want %d", len(diffPortion), MaxDiffChars)
				}
			}
		})
	}
}

func TestBuildPrompt_ContainsRules(t *testing.T) {
	prompt, _ := BuildPrompt("some diff")

	rules := []string{
		"plain text only",
		"imperative mood",
		"under 500 characters",
		"under 72 characters",
	}
	for _, rule := range rules {
		if !strings.Contains(prompt, rule) {
			t.Errorf("prompt missing rule: %q", rule)
		}
	}
}
