# gitsum

Generate concise git commit message summaries using Google Gemini via Vertex AI.

`gitsum` reads the staged and unstaged diffs from a git repository and sends them to Gemini to produce a short, imperative-mood commit message suitable for use with `git commit -m`.

## Installation

```bash
go install github.com/nealhardesty/gitsum@latest
```

Or build from source:

```bash
git clone https://github.com/nealhardesty/gitsum.git
cd gitsum
make build
```

## Prerequisites

- Go 1.24+
- A Google Cloud project with Vertex AI API enabled
- Application Default Credentials configured:

```bash
gcloud auth application-default login
```

## Usage

```bash
# Generate a commit message for current changes (uses gcloud default project)
gitsum

# Use directly with git commit
git commit -m "$(gitsum)"

# Only summarize staged changes
gitsum --staged-only

# Use a different model or region
gitsum -m gemini-2.5-pro -r us-east1

# Override GCP project explicitly
gitsum -p my-gcp-project

# Verbose output (diff stats and model info to stderr)
gitsum --verbose
```

## Options

| Flag | Default | Description |
|---|---|---|
| `-v` | | Print version and exit |
| `-d <dir>` | `.` | Git repository directory |
| `-p <project>` | gcloud default | GCP project ID |
| `-r <region>` | `us-central1` | GCP region |
| `-m <model>` | `gemini-2.5-flash` | Gemini model name |
| `--staged-only` | `false` | Only include staged changes |
| `--verbose` | `false` | Print diff stats and model info to stderr |

## Project Resolution

The GCP project ID is resolved in this order:

1. `-p` flag
2. `GOOGLE_CLOUD_PROJECT` environment variable
3. `CLOUDSDK_CORE_PROJECT` environment variable
4. `gcloud config get-value project` (your gcloud default)

## Architecture

```
main.go        CLI entry point, flag parsing, orchestration
git.go         Git diff extraction via exec.Command
prompt.go      Prompt template construction and diff truncation
gemini.go      Vertex AI Gemini client (Summarizer interface)
version.go     Semantic version constant
```

All status and error output goes to stderr. Only the clean commit message summary goes to stdout, making it safe for command substitution (`$(gitsum)`).

The `Summarizer` interface in `gemini.go` enables mock-based testing without API calls.

Diffs larger than 100,000 characters are truncated with a warning to stderr.

## Dependencies

- [`google.golang.org/genai`](https://pkg.go.dev/google.golang.org/genai) - Google Gen AI unified SDK for Vertex AI

## Development

```bash
make build     # Compile the project
make test      # Run tests with race detection
make lint      # Run go vet
make fmt       # Format code
make tidy      # Run go mod tidy
make version   # Display current version
make help      # Show all targets
```

## License

See [LICENSE](LICENSE) file.
