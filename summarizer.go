package main

import (
	"context"

	llm "github.com/nealhardesty/easy-llm-wrapper"
)

// Summarizer generates a commit message from system and user prompts.
type Summarizer interface {
	Summarize(ctx context.Context, system, user string) (string, error)
}

// LLMSummarizer uses easy-llm-wrapper to generate commit messages.
type LLMSummarizer struct {
	client *llm.Client
}

// Summarize sends the prompts to the configured LLM and returns the generated commit message.
func (s *LLMSummarizer) Summarize(ctx context.Context, system, user string) (string, error) {
	return s.client.Ask(ctx, system, user)
}
