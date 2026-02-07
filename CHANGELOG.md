# Changelog

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
