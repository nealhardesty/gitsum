# Changelog

## [Unreleased] - 2026-03-06

### Changed
- Replaced Google Vertex AI / Gemini backend with `easy-llm-wrapper` (supports Ollama and OpenRouter)
- Removed `-p` (GCP project), `-r` (region), and `-m` (model) CLI flags; provider and model are now configured via environment variables (`OLLAMA_HOST`, `OPENROUTER_API_KEY`, `MODEL`)
- Refactored `BuildPrompt` to return `(system, user string, truncated bool)` — proper system/user prompt split for LLM APIs
- Renamed `GeminiSummarizer` to `LLMSummarizer`; updated `Summarizer` interface signature to `Summarize(ctx, system, user string)`
- Removed `gemini.go` / `gemini_test.go`; added `summarizer.go` / `summarizer_test.go`
- Verbose output now shows `Provider` and `Model` instead of GCP project/region
- Updated help text to document environment variable configuration

### Fixed
- Removed stale `"under 500 characters"` assertion from `TestBuildPrompt_ContainsRules` (rule did not exist in prompt template)

## [0.1.2] - 2026-03-05

### Changed
- Default behavior now uses staged changes only; falls back to all changes when nothing is staged
- Replaced `--staged-only` / `-s` flags with `--all` / `-a` to force inclusion of all changes
- Updated `GetDiff` signature: `stagedOnly bool` → `includeAll bool`
- Updated help text to describe new default behavior
- Updated tests: renamed `TestGetDiff_StagedOnly` → `TestGetDiff_DefaultStagedOnly`, added `TestGetDiff_IncludeAll`

## [0.1.1] - 2026-02-14

### Added
- Short alias `-s` for `--staged-only` flag

## [0.1.0] - 2026-02-06

### Added
- Initial implementation of gitsum CLI tool
- Git diff extraction (staged and unstaged) via `git.go`
- Prompt construction with 100K character truncation limit via `prompt.go`
- Gemini integration via Vertex AI using `google.golang.org/genai` SDK via `gemini.go`
- `Summarizer` interface for testability
- CLI flags: `-v`, `-d`, `-p`, `-r`, `-m`, `--staged-only`, `--verbose`
- GCP project resolution: `-p` flag > `GOOGLE_CLOUD_PROJECT` > `CLOUDSDK_CORE_PROJECT` > `gcloud config get-value project`
- Authentication hint in help output (`gcloud auth application-default login`)
- Clean stdout output for use with `git commit -m "$(gitsum)"`
- Unit tests for git, prompt, and gemini layers
- Makefile with build, test, run, clean, lint, fmt, tidy, version, and help targets
- Semantic versioning via `version.go`
