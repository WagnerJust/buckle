// Package skill holds the buckle skill's install-target registry and the
// per-target body wrappers. One Target == one (tool, default path, body)
// tuple. The body is composed at init from the embedded buckle/SKILL.md
// (the single source of truth) plus a target-specific frontmatter or
// preamble.
//
// The Claude target writes SKILL.md verbatim (Claude Code already
// understands the YAML frontmatter). Other targets get a wrapper that
// is appropriate for that tool's convention — e.g., Cursor MDC
// frontmatter, or an HTML-comment preamble for tools that just read
// raw markdown.
package skill

import (
	"fmt"
	"strings"

	buckle "github.com/WagnerJust/buckle"
)

// Target identifies one place buckle can install the skill.
type Target struct {
	// ID is the stable CLI-visible identifier ("claude", "cursor", ...).
	ID string

	// Name is human-readable ("Claude Code", "Cursor", ...).
	Name string

	// Path is the repo-relative install location for project-scope
	// installs.
	Path string

	// GlobalPath is the install location for user-scope (global)
	// installs, relative to the user's home directory. Empty when
	// the tool has no documented global location.
	GlobalPath string

	// DirTail is the path appended to a user-supplied --dir override.
	// It strips the tool-specific prefix from Path so a user installing
	// "into ~/.agents" lands at ~/.agents/skills/buckle/SKILL.md
	// instead of ~/.agents/.claude/skills/buckle/SKILL.md.
	//
	// For dedicated-skill-dir tools (Claude, Cursor) it is the path
	// after the tool's leading directory (e.g., "skills/buckle/SKILL.md").
	// For shared-file tools (AGENTS.md, .windsurfrules) it is just the
	// filename, so --dir relocates the file without further nesting.
	DirTail string

	// SharedFile is true when Path is a file the user may already be
	// using for other purposes (AGENTS.md, .windsurfrules, etc.). The
	// install command surfaces a warning before writing.
	SharedFile bool

	body string
}

// Body returns the full file body that would be written for this target.
func (t Target) Body() string { return t.body }

// SupportsGlobal reports whether this target has a documented global
// install location.
func (t Target) SupportsGlobal() bool { return t.GlobalPath != "" }

var (
	// rawSkillMD is the canonical buckle/SKILL.md, frontmatter and all.
	rawSkillMD = buckle.SkillMD

	// description is extracted from rawSkillMD's frontmatter so we can
	// preserve it in target-specific wrappers (e.g., Cursor MDC).
	description = extractFrontmatterField(rawSkillMD, "description")

	// coreBody is rawSkillMD with its YAML frontmatter stripped — used
	// by targets that supply their own frontmatter or preamble.
	coreBody = stripFrontmatter(rawSkillMD)

	claudeBody   = rawSkillMD
	cursorBody   = buildCursorBody(description, coreBody)
	geminiBody   = buildCommentBody("Gemini Code Assist", coreBody)
	agentsBody   = buildCommentBody("AGENTS.md convention (Codex CLI, OpenCode, etc.)", coreBody)
	windsurfBody = buildCommentBody("Windsurf / Cline single-file rules", coreBody)
	copilotBody  = buildCommentBody("GitHub Copilot project instructions", coreBody)
)

// targets is the ordered list of supported install targets. Order is
// display-friendly: dedicated-skill-directory tools first, then shared-
// file conventions.
var targets = []Target{
	{
		ID:         "claude",
		Name:       "Claude Code",
		Path:       ".claude/skills/buckle/SKILL.md",
		GlobalPath: ".claude/skills/buckle/SKILL.md",
		DirTail:    "skills/buckle/SKILL.md",
		body:       claudeBody,
	},
	{
		ID:         "cursor",
		Name:       "Cursor",
		Path:       ".cursor/rules/buckle.mdc",
		GlobalPath: ".cursor/rules/buckle.mdc",
		DirTail:    "rules/buckle.mdc",
		body:       cursorBody,
	},
	{
		ID:         "gemini",
		Name:       "Gemini Code Assist",
		Path:       "GEMINI.md",
		GlobalPath: ".gemini/GEMINI.md",
		DirTail:    "GEMINI.md",
		SharedFile: true,
		body:       geminiBody,
	},
	{
		ID:         "codex",
		Name:       "OpenAI Codex CLI",
		Path:       "AGENTS.md",
		GlobalPath: ".codex/AGENTS.md",
		DirTail:    "AGENTS.md",
		SharedFile: true,
		body:       agentsBody,
	},
	{
		ID:         "opencode",
		Name:       "OpenCode",
		Path:       "AGENTS.md",
		GlobalPath: ".config/opencode/AGENTS.md",
		DirTail:    "AGENTS.md",
		SharedFile: true,
		body:       agentsBody,
	},
	{
		ID:         "windsurf",
		Name:       "Windsurf",
		Path:       ".windsurfrules",
		DirTail:    ".windsurfrules",
		SharedFile: true,
		body:       windsurfBody,
	},
	{
		ID:         "cline",
		Name:       "Cline",
		Path:       ".clinerules",
		DirTail:    ".clinerules",
		SharedFile: true,
		body:       windsurfBody,
	},
	{
		ID:         "copilot",
		Name:       "GitHub Copilot",
		Path:       ".github/copilot-instructions.md",
		DirTail:    "copilot-instructions.md",
		SharedFile: true,
		body:       copilotBody,
	},
}

// Targets returns all known install targets in display order.
func Targets() []Target {
	out := make([]Target, len(targets))
	copy(out, targets)
	return out
}

// TargetByID looks up a target by its stable CLI ID.
func TargetByID(id string) (Target, bool) {
	for _, t := range targets {
		if t.ID == id {
			return t, true
		}
	}
	return Target{}, false
}

// IDs returns every target ID in display order.
func IDs() []string {
	out := make([]string, len(targets))
	for i, t := range targets {
		out[i] = t.ID
	}
	return out
}

// GlobalCapableIDs returns target IDs that have a documented global
// install location, in display order.
func GlobalCapableIDs() []string {
	var out []string
	for _, t := range targets {
		if t.SupportsGlobal() {
			out = append(out, t.ID)
		}
	}
	return out
}

// extractFrontmatterField pulls the value of a single-line YAML
// frontmatter field from a markdown document. Returns "" if the
// document has no frontmatter or the field is absent. Quotes around
// the value are preserved verbatim so the caller can decide how to
// re-emit them.
func extractFrontmatterField(doc, key string) string {
	rest, ok := strings.CutPrefix(doc, "---\n")
	if !ok {
		return ""
	}
	fm, _, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return ""
	}
	prefix := key + ":"
	for line := range strings.SplitSeq(fm, "\n") {
		if v, ok := strings.CutPrefix(line, prefix); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// stripFrontmatter removes the YAML frontmatter block at the top of a
// markdown document. If there is no frontmatter, returns the document
// unchanged.
func stripFrontmatter(doc string) string {
	rest, ok := strings.CutPrefix(doc, "---\n")
	if !ok {
		return doc
	}
	_, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return doc
	}
	return strings.TrimLeft(body, "\n")
}

// yamlDoubleQuote escapes a string for use as a double-quoted YAML
// scalar. Backslashes and double-quotes are escaped; the rest passes
// through. Sufficient for descriptions that may contain inline
// double-quote characters but not multi-line content.
func yamlDoubleQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

// buildCursorBody wraps the markdown body in Cursor MDC frontmatter.
// Cursor reads the description for rule listings; globs + alwaysApply
// false means agents consult the rule on demand rather than always.
func buildCursorBody(description, body string) string {
	return fmt.Sprintf(`---
description: %s
globs:
  - "**/*"
alwaysApply: false
---

%s`, yamlDoubleQuote(description), body)
}

// buildCommentBody prepends an HTML-comment preamble identifying the
// target tool, then the markdown body. Used for tools that consume
// raw markdown without YAML frontmatter (Gemini, AGENTS.md, Windsurf,
// Copilot, etc.).
func buildCommentBody(toolDesc, body string) string {
	return fmt.Sprintf(`<!-- buckle skill installed for %s. The body below is the
buckle skill — instructions for setting up a Hub-and-Spoke agent
documentation system. Tools that do not natively understand "skills"
should treat this as their general agent-instructions content. -->

%s`, toolDesc, body)
}
