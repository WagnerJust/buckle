// Package buckle exposes the embedded buckle skill content so the
// install command (cmd/buckle) and the target registry (internal/skill)
// can read it without keeping a duplicate copy in source.
//
// The single source of truth is buckle/SKILL.md — Claude Code loads
// the skill from that exact path, and the binary writes it (or a
// per-tool wrapping) wherever the user installs it. Keeping the
// embed at the module root is the only way Go's //go:embed will
// reach into the buckle/ subdirectory; embed patterns cannot walk
// upward with "..".
package buckle

import _ "embed"

//go:embed buckle/SKILL.md
var SkillMD string
