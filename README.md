# gitsum

Generate concise git commit message summaries using an LLM via [easy-llm-wrapper](https://github.com/nealhardesty/easy-llm-wrapper).

`gitsum` reads diffs from a git repository and sends them to an LLM to produce a detailed, imperative-mood commit message suitable for use with `git commit -m`. By default it uses staged changes only, falling back to all changes when nothing is staged.

Supports **Ollama** (local) and **OpenRouter** (cloud) as backends. Provider and model are selected automatically from environment variables.

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
- One of:
  - **Ollama**: set `OLLAMA_HOST` (e.g. `http://localhost:11434`)
  - **OpenRouter**: set `OPENROUTER_API_KEY`

## Provider Configuration

Provider and model are configured via environment variables. Ollama takes priority over OpenRouter when both are set.

| Variable | Description |
|---|---|
| `OLLAMA_HOST` | Ollama base URL. When set, Ollama is used as the provider. |
| `OPENROUTER_API_KEY` | OpenRouter API key. Used when `OLLAMA_HOST` is not set. |
| `MODEL` | Override the default model for the active provider. |

Default models:
- Ollama: `llama3.2`
- OpenRouter: `anthropic/claude-3-haiku`

## Usage

```bash
# Generate a commit message (staged changes only; falls back to all if nothing staged)
gitsum

# Use directly with git commit
git commit -m "$(gitsum)"

# Include all changes (staged + unstaged + untracked) regardless of staging
gitsum --all

# Specify a different git directory
gitsum -d /path/to/repo

# Verbose output (diff stats, provider, and model to stderr)
gitsum --verbose

# Use a specific model (via env var)
MODEL=llama3.1:70b gitsum
MODEL=anthropic/claude-3-5-sonnet gitsum
```

## Options

| Flag | Default | Description |
|---|---|---|
| `-v` | | Print version and exit |
| `-d <dir>` | `.` | Git repository directory |
| `--all` / `-a` | `false` | Include all changes (staged + unstaged + untracked) |
| `--verbose` | `false` | Print diff stats, provider, and model to stderr |

## Architecture

```
main.go          CLI entry point, flag parsing, orchestration
git.go           Git diff extraction via exec.Command
prompt.go        Prompt construction (system + user split) and diff truncation
summarizer.go    Summarizer interface + LLMSummarizer using easy-llm-wrapper
version.go       Semantic version constant
```

All status and error output goes to stderr. Only the clean commit message goes to stdout, making it safe for command substitution (`$(gitsum)`).

The `Summarizer` interface in `summarizer.go` enables mock-based testing without live LLM calls.

Diffs larger than 100,000 characters are truncated with a warning to stderr.

## Dependencies

- [`github.com/nealhardesty/easy-llm-wrapper`](https://github.com/nealhardesty/easy-llm-wrapper) — unified LLM client for Ollama and OpenRouter

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
