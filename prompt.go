package main

const (
	// MaxDiffChars is the maximum number of characters allowed in a diff
	// before truncation occurs.
	MaxDiffChars = 100_000

	systemInstruction = `You are a senior software engineer writing a comprehensive git commit message.

Critical rules:
- Use imperative mood (e.g., "Add feature" not "Added feature")
- First line: specific subject under 72 characters describing the primary change
- Always add a blank line after the subject
- Add bullet points proportional to the scope of the change:
  - Trivial changes (1-2 files, minor edits): 1-2 bullets max
  - Moderate changes (a few files or features): 2-4 bullets
  - Large changes (many files, new systems): 4-8 bullets
- Be SPECIFIC: mention actual file types, functions, APIs, features by name
- Focus on WHAT changed and WHY (business/technical reason), not HOW
- For large changes, group related items together
- Use plain text only, no markdown formatting, code blocks, or quotes
- Output only the commit message, nothing else

Example (small change):
Fix typo in Makefile release target

- Correct misspelled variable name in release recipe

Example (large change):
Implement user authentication system

- Add JWT token generation and validation in auth module
- Create user login and registration endpoints
- Implement password hashing with bcrypt
- Add authentication middleware for protected routes
- Add comprehensive test suite for auth flows`
)

// BuildPrompt constructs the system and user prompts for the LLM from the diff text.
// If the diff exceeds MaxDiffChars, it is truncated and truncated is true.
func BuildPrompt(diff string) (system, user string, truncated bool) {
	if len(diff) > MaxDiffChars {
		diff = diff[:MaxDiffChars]
		truncated = true
	}
	return systemInstruction, "Diff:\n" + diff, truncated
}
