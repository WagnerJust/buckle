package skill

import (
	"strings"
	"testing"
)

func TestTargetsAllPopulated(t *testing.T) {
	for _, tg := range Targets() {
		if tg.ID == "" {
			t.Errorf("target with empty ID: %+v", tg)
		}
		if tg.Name == "" {
			t.Errorf("target %q has empty Name", tg.ID)
		}
		if tg.Path == "" {
			t.Errorf("target %q has empty Path", tg.ID)
		}
		if tg.DirTail == "" {
			t.Errorf("target %q has empty DirTail; --dir would not work", tg.ID)
		}
		if tg.Body() == "" {
			t.Errorf("target %q has empty Body", tg.ID)
		}
	}
}

func TestTargetByIDFound(t *testing.T) {
	tg, ok := TargetByID("claude")
	if !ok {
		t.Fatal("expected claude target to be found")
	}
	if tg.ID != "claude" {
		t.Errorf("got ID=%q, want claude", tg.ID)
	}
}

func TestTargetByIDNotFound(t *testing.T) {
	if _, ok := TargetByID("nonexistent"); ok {
		t.Error("expected nonexistent target to return ok=false")
	}
}

func TestIDsOrderMatchesTargets(t *testing.T) {
	ids := IDs()
	got := Targets()
	if len(ids) != len(got) {
		t.Fatalf("IDs len %d != Targets len %d", len(ids), len(got))
	}
	for i, id := range ids {
		if got[i].ID != id {
			t.Errorf("IDs[%d]=%q does not match Targets[%d].ID=%q", i, id, i, got[i].ID)
		}
	}
}

func TestGlobalCapableIDsExcludesNonGlobal(t *testing.T) {
	got := GlobalCapableIDs()
	for _, id := range got {
		tg, _ := TargetByID(id)
		if !tg.SupportsGlobal() {
			t.Errorf("GlobalCapableIDs includes %q, but it has no GlobalPath", id)
		}
	}
	// Sanity: copilot is project-only in the registry.
	for _, id := range got {
		if id == "copilot" {
			t.Error("copilot listed as global-capable but has no documented global path")
		}
	}
}

func TestClaudeBodyIsVerbatimSkill(t *testing.T) {
	tg, _ := TargetByID("claude")
	if tg.Body() != rawSkillMD {
		t.Error("claude body should be the embedded SKILL.md verbatim")
	}
}

func TestCursorBodyHasCursorFrontmatter(t *testing.T) {
	tg, _ := TargetByID("cursor")
	body := tg.Body()
	if !strings.HasPrefix(body, "---\ndescription: ") {
		t.Errorf("cursor body should start with cursor MDC frontmatter; got first 60 chars: %q", body[:min(60, len(body))])
	}
	if !strings.Contains(body, "alwaysApply: false") {
		t.Error("cursor body should set alwaysApply: false so the rule is consulted on demand")
	}
	if !strings.Contains(body, "globs:") {
		t.Error("cursor body should declare globs in frontmatter")
	}
}

func TestSharedFileBodiesHaveCommentPreamble(t *testing.T) {
	for _, tg := range Targets() {
		if !tg.SharedFile {
			continue
		}
		if !strings.HasPrefix(tg.Body(), "<!--") {
			t.Errorf("shared-file target %q should start with HTML comment preamble; got first 60 chars: %q",
				tg.ID, tg.Body()[:min(60, len(tg.Body()))])
		}
	}
}

func TestExtractFrontmatterFieldFound(t *testing.T) {
	doc := "---\nname: foo\ndescription: hello world\n---\n\nbody"
	if got := extractFrontmatterField(doc, "description"); got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestExtractFrontmatterFieldMissing(t *testing.T) {
	doc := "---\nname: foo\n---\n\nbody"
	if got := extractFrontmatterField(doc, "description"); got != "" {
		t.Errorf("missing field should return empty, got %q", got)
	}
}

func TestExtractFrontmatterFieldNoFrontmatter(t *testing.T) {
	doc := "no frontmatter\nhere"
	if got := extractFrontmatterField(doc, "description"); got != "" {
		t.Errorf("no frontmatter should return empty, got %q", got)
	}
}

func TestStripFrontmatterRemovesYAMLBlock(t *testing.T) {
	doc := "---\nname: foo\n---\n\nthe body"
	if got := stripFrontmatter(doc); got != "the body" {
		t.Errorf("got %q, want %q", got, "the body")
	}
}

func TestStripFrontmatterPassesThroughWhenNone(t *testing.T) {
	doc := "no frontmatter here"
	if got := stripFrontmatter(doc); got != doc {
		t.Errorf("doc without frontmatter should pass through unchanged")
	}
}

func TestYAMLDoubleQuoteEscapesQuotes(t *testing.T) {
	got := yamlDoubleQuote(`a "quoted" word`)
	want := `"a \"quoted\" word"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestYAMLDoubleQuoteEscapesBackslashes(t *testing.T) {
	got := yamlDoubleQuote(`a \backslash`)
	want := `"a \\backslash"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEveryBodyContainsSkillCore(t *testing.T) {
	// The "Buckle" heading is in the body; every target's wrapper
	// must include it so we know the body wasn't accidentally dropped.
	for _, tg := range Targets() {
		if !strings.Contains(tg.Body(), "# Buckle") {
			t.Errorf("target %q body is missing the skill heading; the wrapper may have dropped the body", tg.ID)
		}
	}
}
