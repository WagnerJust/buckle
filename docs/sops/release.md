# Cutting a release

Buckle ships as a single Go binary. Releases are tag-driven: pushing a `vX.Y.Z`
tag triggers `.github/workflows/release.yml`, which runs GoReleaser and
publishes a GitHub Release with cross-platform archives + checksums.

## Steps

1. **Pick the version.** Follow [SemVer](https://semver.org). Exported behavior
   change in `cmd/buckle` or `internal/skill` → minor bump. Bug fix only →
   patch. The skill body itself (`buckle/SKILL.md`) is part of the public
   surface; meaningful skill edits warrant at least a minor bump.

2. **Update `CHANGELOG.md`.**
   - Rename the `## [Unreleased]` heading to `## [X.Y.Z] - YYYY-MM-DD`.
   - Add a fresh empty `## [Unreleased]` section above it.
   - Update the link references at the bottom of the file.
   - Group entries under `Added`, `Changed`, `Fixed`, `Removed`, `Deprecated`,
     `Security` (Keep a Changelog).

3. **Verify locally.**
   ```sh
   make test test-race vet
   make build
   ./bin/buckle version    # confirms ldflags wired up
   goreleaser check        # validates .goreleaser.yaml
   goreleaser release --snapshot --clean  # optional: dry-run cross-build
   ```

4. **Commit and tag.**
   ```sh
   git commit -am "release: vX.Y.Z"
   git tag -a vX.Y.Z -m "vX.Y.Z"
   git push origin main --follow-tags
   ```

5. **Watch the workflow.** GitHub Actions runs `release.yml` on the tag push.
   It publishes the release; review the generated notes and edit if needed.

## NEVER

- NEVER tag without bumping `CHANGELOG.md` first. The release notes link back
  to the changelog entry.
- NEVER move or delete a published tag. If you cut a bad release, ship a new
  patch instead.
- NEVER hand-edit `version` strings in source — they are stamped at build time
  via `-ldflags` (see `Makefile` and `.goreleaser.yaml`).
