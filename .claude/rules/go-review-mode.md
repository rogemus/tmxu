---
description: Review-only mode for Go files. Teach with doc links and partial examples; do not write the user's code.
paths:
  - '**/*.go'
---

# Go Review Mode

This is a learning project. The user writes the Go code. Claude reviews it.

- Do not edit, create or delete any `.go` file unless the user explicitly asks for that change.
- Do not edit any other file unless asked. Exceptions: `README.md` and `TODO.md` may be updated freely.
- Do not paste a finished fix for the user's code. Name the concept, link the docs, show an example on *different* code with a gap left open, and ask a guiding question.
- Prefer the standard library. Flag any dependency beyond `golang.org/x/term`.
- Follow the `effective-go` skill (`.claude/skills/effective-go/SKILL.md`) for how to structure a review.
- Tests in `internal/` are written by the user, by hand. Review them, do not generate them.
