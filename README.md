# buckle

Install a **Hub-and-Spoke** agent documentation skill into the location your
coding agent expects.

The hub is `AGENTS.md` at the repo root — a one-screen table of contents that
routes agents to spoke files (docs, runbooks, conventions, ADRs). Tool-specific
files (`CLAUDE.md`, `.cursorrules`, `GEMINI.md`, …) become one-line pointers
back to the hub, so rules don't drift across tools.

`buckle` is a single Go binary that writes the skill content into the right
file for each tool. The skill itself lives at
[`buckle/SKILL.md`](buckle/SKILL.md) and is embedded into the binary at build
time.

## Install

```sh
# From source (latest):
go install github.com/WagnerJust/buckle/cmd/buckle@latest
```

Or grab a pre-built archive from the
[releases page](https://github.com/WagnerJust/buckle/releases) — Linux, macOS,
and Windows on amd64 / arm64.

## Usage

```sh
# Dry-run: show where buckle would install the skill for Claude Code (default).
buckle install

# Install for Claude Code in the current repo.
buckle install --apply

# Install for Cursor.
buckle install --target cursor --apply

# Install Claude's skill globally under a custom dir.
buckle install --target claude --global --dir ~/.agents --apply

# List supported targets.
buckle install --list

# Print version / commit / build date.
buckle version
```

Default is **dry-run** — `--apply` is required to actually write.

## Supported targets

| ID         | Tool               | Path                                    |
| ---------- | ------------------ | --------------------------------------- |
| `claude`   | Claude Code        | `.claude/skills/buckle/SKILL.md`        |
| `cursor`   | Cursor             | `.cursor/rules/buckle.mdc`              |
| `gemini`   | Gemini Code Assist | `GEMINI.md` *(shared)*                  |
| `codex`    | OpenAI Codex CLI   | `AGENTS.md` *(shared)*                  |
| `opencode` | OpenCode           | `AGENTS.md` *(shared)*                  |
| `windsurf` | Windsurf           | `.windsurfrules` *(shared)*             |
| `cline`    | Cline              | `.clinerules` *(shared)*                |
| `copilot`  | GitHub Copilot     | `.github/copilot-instructions.md` *(shared)* |

*Shared* targets write to a file you might already use for other things.
buckle refuses to overwrite, so an existing file yields a clear error rather
than clobbering your work.

## Build from source

```sh
make build       # → bin/buckle (with version stamped via -ldflags)
make test        # go test ./...
make test-race   # go test -race ./...
```

The only direct dependency is [`spf13/cobra`](https://github.com/spf13/cobra).
Frontmatter is parsed by hand to keep the binary tiny and dependency-free.

## Repository layout

- `buckle/SKILL.md` — canonical skill content (single source of truth)
- `cmd/buckle/` — CLI entry point and subcommands
- `internal/skill/` — target registry + body wrappers per tool
- `docs/agents.md` — agent rules for working in this repo
- `docs/sops/` — SOPs (adding install targets, cutting releases)
- `AGENTS.md` — hub for this repo (yes, buckle uses itself)

## Contributing

See [`AGENTS.md`](AGENTS.md) for the routes humans and agents should follow
when changing this repo. New install target? Follow
[`docs/sops/add-install-target.md`](docs/sops/add-install-target.md). Cutting
a release? Follow [`docs/sops/release.md`](docs/sops/release.md).
