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
		wantContains  string // checked against user prompt
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
			system, user, truncated := BuildPrompt(tt.diff)

			if truncated != tt.wantTruncated {
				t.Errorf("truncated = %v, want %v", truncated, tt.wantTruncated)
			}

			if tt.wantContains != "" && !strings.Contains(user, tt.wantContains) {
				t.Errorf("user prompt does not contain %q", tt.wantContains)
			}

			// System prompt always contains the core instructions.
			if !strings.Contains(system, "imperative mood") {
				t.Error("system prompt missing imperative mood instruction")
			}

			if tt.wantTruncated {
				// The diff portion of the user prompt should be exactly MaxDiffChars.
				if !strings.HasPrefix(user, "Diff:\n") {
					t.Fatal("user prompt missing 'Diff:' prefix")
				}
				diffPortion := user[len("Diff:\n"):]
				if len(diffPortion) != MaxDiffChars {
					t.Errorf("truncated diff length = %d, want %d", len(diffPortion), MaxDiffChars)
				}
			}
		})
	}
}

func TestBuildPrompt_ContainsRules(t *testing.T) {
	system, _, _ := BuildPrompt("some diff")

	rules := []string{
		"plain text only",
		"imperative mood",
		"under 72 characters",
	}
	for _, rule := range rules {
		if !strings.Contains(system, rule) {
			t.Errorf("system prompt missing rule: %q", rule)
		}
	}
}
