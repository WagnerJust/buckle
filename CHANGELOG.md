# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `README.md` with install, usage, and supported-target reference.
- `README.md` is now bundled into release archives.

## [0.1.0] - 2026-05-04

### Added

- Initial public release of `buckle`.
- `buckle install` — write the buckle Hub-and-Spoke skill into the location
  expected by Claude Code, Cursor, Gemini, Codex CLI, OpenCode, Windsurf,
  and Cline. Defaults to dry-run; `--apply` commits.
- `buckle version` subcommand that prints the version, commit, and build date.
- Version metadata is stamped into the binary at build time via `-ldflags`
  (driven by `make build` / `make install`).
- Release pipeline: `.goreleaser.yaml` and a tag-triggered GitHub Actions
  workflow that builds cross-platform archives and publishes a GitHub Release.
- `docs/sops/release.md` — SOP for cutting a release.
- `docs/sops/add-install-target.md` — SOP for wiring up a new coding-agent
  install target.

[Unreleased]: https://github.com/WagnerJust/buckle/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/WagnerJust/buckle/releases/tag/v0.1.0
