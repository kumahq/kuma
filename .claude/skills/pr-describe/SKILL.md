---
name: pr-describe
description: Generate a PR description with Motivation and Implementation information sections
user-invocable: true
allowed-tools: Bash(git *), Read, Grep, Glob
---

# Generate PR Description

Analyze the current branch changes and generate a PR description.

## Process

1. Determine the base branch (default: `master`)
2. Get the diff and commit history:
   - `git log master...HEAD --oneline`
   - `git diff master...HEAD --stat`
   - `git diff master...HEAD`
3. Read relevant changed files if needed for deeper context
4. Write the PR description

## Output Format

Output ONLY this markdown structure:

```
## Motivation

<!-- Explain WHY this change is needed. What problem does it solve? What issue does it address? -->

[Describe the motivation here]

## Implementation information

<!-- Explain HOW the change was implemented. Key decisions, trade-offs, notable details. -->

[Describe the implementation here]
```

## Rules

- Be concise and specific
- Motivation: focus on the problem/need, not the solution
- Implementation: focus on what was done and key decisions
- Reference file paths when relevant
- If $ARGUMENTS is provided, use it as additional context for the description
