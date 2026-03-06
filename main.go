package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	llm "github.com/nealhardesty/easy-llm-wrapper"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		showVersion bool
		dir         string
		model       string
		includeAll  bool
		verbose     bool
		debug       bool
		extraDebug  bool
		noStream    bool
	)

	flag.BoolVar(&showVersion, "v", false, "Print version and exit")
	flag.StringVar(&dir, "dir", ".", "Git repository directory")
	flag.StringVar(&model, "m", "gpt-oss:20b", "Model override (also honoured via MODEL env var)")
	flag.BoolVar(&includeAll, "all", false, "Include all changes (staged + unstaged + untracked)")
	flag.BoolVar(&includeAll, "a", false, "Include all changes (alias for --all)")
	flag.BoolVar(&verbose, "verbose", false, "Print diff stats and model info to stderr")
	flag.BoolVar(&debug, "d", false, "Debug mode: print env, config, diff stats, and prompt sizes to stderr")
	flag.BoolVar(&extraDebug, "D", false, "Extra debug: everything -d does plus print prompts to stderr and write system-prompt.txt / user-prompt.txt")
	flag.BoolVar(&noStream, "S", false, "Disable streaming (collect full response before printing)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "gitsum v%s - Generate git commit messages using an LLM\n\n", Version)
		fmt.Fprintf(os.Stderr, "Usage: gitsum [options]\n\n")
		fmt.Fprintf(os.Stderr, "Default behavior: uses staged changes only; falls back to all changes\n")
		fmt.Fprintf(os.Stderr, "if nothing is staged. Responses are streamed by default.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nProvider configuration (environment variables):\n")
		fmt.Fprintf(os.Stderr, "  OLLAMA_HOST        Ollama base URL (e.g. http://localhost:11434)\n")
		fmt.Fprintf(os.Stderr, "                     When set, Ollama is used as the provider.\n")
		fmt.Fprintf(os.Stderr, "  OPENROUTER_API_KEY OpenRouter API key.\n")
		fmt.Fprintf(os.Stderr, "                     Used when OLLAMA_HOST is not set.\n")
		fmt.Fprintf(os.Stderr, "  MODEL              Override the model (takes priority over -m flag).\n\n")
		fmt.Fprintf(os.Stderr, "Provider selection: Ollama takes priority over OpenRouter.\n\n")
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "  git commit -m \"$(gitsum)\"\n")
	}

	flag.Parse()

	if showVersion {
		fmt.Fprintf(os.Stderr, "gitsum v%s\n", Version)
		return 0
	}

	// -D implies -d.
	if extraDebug {
		debug = true
	}

	if debug {
		debugf("=== gitsum v%s ===", Version)
		debugf("OLLAMA_HOST        = %q", os.Getenv("OLLAMA_HOST"))
		debugf("OPENROUTER_API_KEY = %q", maskSecret(os.Getenv("OPENROUTER_API_KEY")))
		debugf("MODEL (env)        = %q", os.Getenv("MODEL"))
		debugf("-m flag            = %q", model)
		debugf("--dir              = %q", dir)
		debugf("--all              = %v", includeAll)
		debugf("streaming          = %v", !noStream)
	}

	// Apply model default: env var takes priority, then -m flag.
	if os.Getenv("MODEL") == "" {
		os.Setenv("MODEL", model)
	}

	// Create LLM client early for fail-fast behavior.
	client, err := llm.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if debug {
		debugf("provider           = %s", client.Provider())
		debugf("model              = %s", client.Model())
	}

	// Get the diff.
	diff, err := GetDiff(dir, includeAll)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if diff.IsEmpty() {
		fmt.Fprintln(os.Stderr, "No changes detected.")
		return 0
	}

	combined := diff.Combined()

	if debug || verbose {
		fmt.Fprintf(os.Stderr, "Staged diff:    %d chars\n", len(diff.Staged))
		fmt.Fprintf(os.Stderr, "Unstaged diff:  %d chars\n", len(diff.Unstaged))
		fmt.Fprintf(os.Stderr, "Untracked diff: %d chars\n", len(diff.Untracked))
		fmt.Fprintf(os.Stderr, "Provider:       %s\n", client.Provider())
		fmt.Fprintf(os.Stderr, "Model:          %s\n", client.Model())
	}

	// Build the prompt.
	system, user, truncated := BuildPrompt(combined)
	if truncated {
		fmt.Fprintf(os.Stderr, "Warning: diff truncated to %d characters\n", MaxDiffChars)
	}

	if debug {
		debugf("system prompt      = %d chars", len(system))
		debugf("user prompt        = %d chars", len(user))
		debugf("truncated          = %v", truncated)
	}

	if extraDebug {
		if werr := os.WriteFile("system-prompt.txt", []byte(system), 0644); werr != nil {
			debugf("warning: could not write system-prompt.txt: %v", werr)
		} else {
			debugf("system-prompt.txt written")
		}
		if werr := os.WriteFile("user-prompt.txt", []byte(user), 0644); werr != nil {
			debugf("warning: could not write user-prompt.txt: %v", werr)
		} else {
			debugf("user-prompt.txt written")
		}
		debugf("--- system prompt (%d chars) ---", len(system))
		fmt.Fprintln(os.Stderr, system)
		debugf("--- user prompt (%d chars) ---", len(user))
		fmt.Fprintln(os.Stderr, user)
		debugf("--- end prompts ---")
	}

	if debug {
		debugf("--- sending to LLM ---")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	start := time.Now()

	if !noStream {
		stream, serr := client.Stream(ctx, llm.Request{
			System: system,
			Messages: []llm.Message{
				{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart(user)}},
			},
		})
		if serr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", serr)
			return 1
		}
		defer stream.Close()

		var chunks, totalChars int
		for stream.Next() {
			chunk := stream.Chunk()
			fmt.Print(chunk)
			chunks++
			totalChars += len(chunk)
		}
		fmt.Println()

		if serr := stream.Err(); serr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", serr)
			return 1
		}

		if debug {
			debugf("chunks             = %d", chunks)
			debugf("response           = %d chars", totalChars)
			debugf("elapsed            = %s", time.Since(start).Round(time.Millisecond))
			debugf("--- done ---")
		}
	} else {
		summarizer := &LLMSummarizer{client: client}
		summary, serr := summarizer.Summarize(ctx, system, user)
		if serr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", serr)
			return 1
		}

		if debug {
			debugf("response           = %d chars", len(summary))
			debugf("elapsed            = %s", time.Since(start).Round(time.Millisecond))
			debugf("--- done ---")
		}

		fmt.Println(summary)
	}

	return 0
}

// debugf prints a labelled debug line to stderr.
func debugf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", args...)
}

// maskSecret redacts all but the first 4 characters of a secret.
func maskSecret(s string) string {
	if len(s) <= 4 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-4)
}
