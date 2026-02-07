package main

import "fmt"

const (
	// MaxDiffChars is the maximum number of characters allowed in a diff
	// before truncation occurs.
	MaxDiffChars = 100_000

	promptTemplate = `You are a senior software engineer writing a comprehensive git commit message.

Analyze the following git diff and produce a detailed commit message that thoroughly describes the changes.

Critical rules:
- Use imperative mood (e.g., "Add feature" not "Added feature")
- First line: specific subject under 72 characters describing the primary change
- Always add a blank line after the subject
- Add 3-8 detailed bullet points describing specific changes:
  - New files/modules added and their purpose
  - Modified functionality and what changed
  - Removed code and why
  - Configuration or infrastructure changes
  - Tests added or updated
  - Documentation updates
- Be SPECIFIC: mention actual file types, functions, APIs, features by name
- Focus on WHAT changed and WHY (business/technical reason), not HOW
- For large changes, group related items together
- Use plain text only, no markdown formatting, code blocks, or quotes

Example structure:
Implement user authentication system

- Add JWT token generation and validation in auth module
- Create user login and registration endpoints
- Implement password hashing with bcrypt
- Add authentication middleware for protected routes
- Create user session management with Redis
- Add comprehensive test suite for auth flows
- Update API documentation with authentication requirements

Diff:
%s`
)

// BuildPrompt constructs the prompt for the Gemini model from the diff text.
// If the diff exceeds MaxDiffChars, it is truncated and a warning is returned.
func BuildPrompt(diff string) (prompt string, truncated bool) {
	if len(diff) > MaxDiffChars {
		diff = diff[:MaxDiffChars]
		truncated = true
	}
	return fmt.Sprintf(promptTemplate, diff), truncated
}
