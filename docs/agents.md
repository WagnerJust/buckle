# Agent rules — buckle repo

Apply to all code in this repo.

## Source of truth

- `buckle/SKILL.md` is the canonical skill content. The binary embeds it via `//go:embed buckle/SKILL.md` in `embed.go`. NEVER edit `.claude/skills/buckle/SKILL.md` directly — that path is an install destination and gets overwritten on the next `buckle install --apply`.
- The Go module path is `github.com/WagnerJust/buckle`. Required Go version: `1.24` (see `go.mod`).

## Testing

- Run `go test ./...` before claiming a change is done. Use `make test test-race vet` for the full local check.

## Build

- `make build` produces `bin/buckle`. `make install` puts it on `$GOPATH/bin`.
- `bin/` is gitignored.

## Dependencies

- The only direct dependency is `github.com/spf13/cobra`. Do NOT add a YAML parser — `internal/skill/skill.go` parses frontmatter by hand on purpose so the binary stays small and dependency-free.

## Eval workspace

- `buckle-workspace/` is gitignored eval scratch. Do not commit anything from it. If you need to reference a past eval result, copy the relevant snippet into a tracked doc instead.
