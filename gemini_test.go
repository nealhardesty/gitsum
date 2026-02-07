package main

import (
	"context"
	"testing"
)

// mockSummarizer is a test double for Summarizer.
type mockSummarizer struct {
	response string
	err      error
}

func (m *mockSummarizer) Summarize(_ context.Context, _ string) (string, error) {
	return m.response, m.err
}

func TestMockSummarizer_Success(t *testing.T) {
	mock := &mockSummarizer{
		response: "Add new feature for user auth",
	}

	result, err := mock.Summarize(context.Background(), "some prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Add new feature for user auth" {
		t.Errorf("got %q, want %q", result, "Add new feature for user auth")
	}
}

func TestMockSummarizer_Error(t *testing.T) {
	mock := &mockSummarizer{
		err: context.DeadlineExceeded,
	}

	_, err := mock.Summarize(context.Background(), "some prompt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGeminiSummarizer_ImplementsInterface(t *testing.T) {
	// Compile-time check that GeminiSummarizer implements Summarizer.
	var _ Summarizer = (*GeminiSummarizer)(nil)
}
