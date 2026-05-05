# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `buckle version` subcommand that prints the version, commit, and build date.
- Version metadata is now stamped into the binary at build time via `-ldflags`
  (driven by `make build` / `make install`).
- Release scaffolding: `.goreleaser.yaml` and a tag-triggered GitHub Actions
  workflow that builds cross-platform archives and publishes a GitHub Release.
- `docs/sops/release.md` — SOP for cutting a release.

[Unreleased]: https://github.com/WagnerJust/buckle/commits/main
