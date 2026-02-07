package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		showVersion bool
		dir         string
		project     string
		region      string
		model       string
		stagedOnly  bool
		verbose     bool
	)

	flag.BoolVar(&showVersion, "v", false, "Print version and exit")
	flag.StringVar(&dir, "d", ".", "Git repository directory")
	flag.StringVar(&project, "p", "", "GCP project ID (default: gcloud config project)")
	flag.StringVar(&region, "r", "us-central1", "GCP region")
	flag.StringVar(&model, "m", "gemini-2.5-flash", "Gemini model name")
	flag.BoolVar(&stagedOnly, "staged-only", false, "Only include staged changes")
	flag.BoolVar(&verbose, "verbose", false, "Print diff stats and model info to stderr")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "gitsum v%s - Generate git commit message summaries using Gemini\n\n", Version)
		fmt.Fprintf(os.Stderr, "Usage: gitsum [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nProject resolution order:\n")
		fmt.Fprintf(os.Stderr, "  1. -p flag\n")
		fmt.Fprintf(os.Stderr, "  2. GOOGLE_CLOUD_PROJECT env var\n")
		fmt.Fprintf(os.Stderr, "  3. CLOUDSDK_CORE_PROJECT env var\n")
		fmt.Fprintf(os.Stderr, "  4. gcloud config get-value project\n\n")
		fmt.Fprintf(os.Stderr, "Authentication:\n")
		fmt.Fprintf(os.Stderr, "  Run 'gcloud auth application-default login' to authenticate.\n\n")
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "  git commit -m \"$(gitsum)\"\n")
	}

	flag.Parse()

	if showVersion {
		fmt.Fprintf(os.Stderr, "gitsum v%s\n", Version)
		return 0
	}

	// Resolve project: flag > env vars > gcloud config.
	if project == "" {
		project = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	if project == "" {
		project = os.Getenv("CLOUDSDK_CORE_PROJECT")
	}
	if project == "" {
		p, err := getDefaultProject()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: no GCP project ID found: %v\n", err)
			return 1
		}
		project = p
	}

	// Get the diff.
	diff, err := GetDiff(dir, stagedOnly)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if diff.IsEmpty() {
		fmt.Fprintln(os.Stderr, "No changes detected.")
		return 0
	}

	combined := diff.Combined()

	if verbose {
		fmt.Fprintf(os.Stderr, "Staged diff:    %d chars\n", len(diff.Staged))
		fmt.Fprintf(os.Stderr, "Unstaged diff:  %d chars\n", len(diff.Unstaged))
		fmt.Fprintf(os.Stderr, "Untracked diff: %d chars\n", len(diff.Untracked))
		fmt.Fprintf(os.Stderr, "Model:          %s\n", model)
		fmt.Fprintf(os.Stderr, "Project:        %s\n", project)
		fmt.Fprintf(os.Stderr, "Region:         %s\n", region)
	}

	// Build the prompt.
	prompt, truncated := BuildPrompt(combined)
	if truncated {
		fmt.Fprintf(os.Stderr, "Warning: diff truncated to %d characters\n", MaxDiffChars)
	}

	// Call Gemini.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	summarizer := &GeminiSummarizer{
		Project:  project,
		Location: region,
		Model:    model,
	}

	summary, err := summarizer.Summarize(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Println(summary)
	return 0
}

// getDefaultProject gets the default GCP project from gcloud config.
func getDefaultProject() (string, error) {
	cmd := exec.Command("gcloud", "config", "get-value", "project")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get default project: %w (ensure gcloud is installed and configured)", err)
	}

	project := strings.TrimSpace(string(output))
	if project == "" {
		return "", fmt.Errorf("no default project set (run: gcloud config set project PROJECT_ID)")
	}

	return project, nil
}
