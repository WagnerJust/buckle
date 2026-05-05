# Adding a new install target

Adding a coding-agent tool to buckle's registry touches 4 places. Follow this SOP to avoid the obvious miss.

## Steps

1. **Edit `internal/skill/skill.go`** — add an entry to the `targets` slice:
   - `ID` (lowercase identifier, e.g., `"newtool"`)
   - `Name` (human-readable)
   - `Path` (project-scope install path, repo-relative)
   - `GlobalPath` (user-scope install path, $HOME-relative; leave empty if the tool has no documented global location)
   - `DirTail` (path the user lands at when they pass `--dir`; strip the tool-specific leading directory so `--dir ~/.agents` lands at `~/.agents/<DirTail>`, not `~/.agents/<Path>`)
   - `SharedFile` (true if `Path` is a file the user may already use for other purposes, e.g., `AGENTS.md`)
   - `body` (per-target body — see step 2)

2. **Pick the body wrapper.**
   - For tools that read YAML frontmatter (e.g., Cursor MDC): write a `build<Tool>Body` helper that prepends the right frontmatter to `coreBody` and assign the result to a package-level `<tool>Body` var.
   - For tools that read raw markdown: use `buildCommentBody("<tool description>", coreBody)` — it prepends an HTML-comment preamble identifying the tool.
   - The Claude target uses `rawSkillMD` verbatim; Claude Code already understands the skill's YAML frontmatter.

3. **Update help text in `cmd/buckle/install.go`** — add the new target to the table in the `Long:` field of `newInstallCmd`. The tabular layout is load-bearing for `--help` output.

4. **Add tests in `internal/skill/skill_test.go`** — the data-driven `TestTargetsAllPopulated` already covers required-field non-emptiness for new targets. If the new target uses a custom body wrapper, add a dedicated test asserting the wrapper produced the expected frontmatter (see `TestCursorBodyHasCursorFrontmatter` for shape).

5. **Run `make test test-race vet`.** All green before commit.

## NEVER

- NEVER skip the `DirTail` field — it's required for `--dir` to behave correctly. `TestTargetsAllPopulated` enforces this.
- NEVER add a target whose `Path` reaches outside the project root or `$HOME`. Buckle refuses to clobber, but it shouldn't be writing to arbitrary system paths to begin with.
