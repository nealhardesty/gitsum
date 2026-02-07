package main

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// Summarizer generates a commit message summary from a prompt.
type Summarizer interface {
	Summarize(ctx context.Context, prompt string) (string, error)
}

// GeminiSummarizer uses Google Gemini via Vertex AI to generate summaries.
type GeminiSummarizer struct {
	Project  string
	Location string
	Model    string
}

// Summarize sends the prompt to Gemini and returns the generated summary.
func (g *GeminiSummarizer) Summarize(ctx context.Context, prompt string) (string, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  g.Project,
		Location: g.Location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return "", fmt.Errorf("creating genai client: %w", err)
	}

	temp := float32(0.3)
	result, err := client.Models.GenerateContent(ctx,
		g.Model,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			Temperature: &temp,
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{
					{Text: "You are a concise git commit message writer. Output only the commit message, nothing else."},
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("generating content: %w", err)
	}

	if len(result.Candidates) == 0 || result.Candidates[0].Content == nil || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from model")
	}

	text := result.Candidates[0].Content.Parts[0].Text
	return strings.TrimSpace(text), nil
}
