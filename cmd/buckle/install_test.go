package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runCLI invokes the root cobra command with args and captures stdout/stderr.
// Returns (stdout, stderr, error). Useful for behavioral tests of the CLI
// without spawning a subprocess.
func runCLI(args ...string) (string, string, error) {
	var out, errBuf bytes.Buffer
	err := run(args, &out, &errBuf)
	return out.String(), errBuf.String(), err
}

func TestInstallList(t *testing.T) {
	out, _, err := runCLI("install", "--list")
	if err != nil {
		t.Fatalf("--list returned error: %v", err)
	}
	if !strings.Contains(out, "claude") || !strings.Contains(out, "cursor") {
		t.Errorf("--list output missing expected targets:\n%s", out)
	}
}

func TestInstallDryRunDoesNotWrite(t *testing.T) {
	tmp := t.TempDir()
	out, _, err := runCLI("install", "--target", "claude", tmp)
	if err != nil {
		t.Fatalf("dry-run returned error: %v", err)
	}
	if !strings.Contains(out, "dry-run") {
		t.Errorf("dry-run output should mention 'dry-run':\n%s", out)
	}
	expectedPath := filepath.Join(tmp, ".claude/skills/buckle/SKILL.md")
	if _, err := os.Stat(expectedPath); !os.IsNotExist(err) {
		t.Errorf("dry-run should not have written %s; stat returned %v", expectedPath, err)
	}
}

func TestInstallApplyWritesFile(t *testing.T) {
	tmp := t.TempDir()
	_, _, err := runCLI("install", "--target", "claude", "--apply", tmp)
	if err != nil {
		t.Fatalf("apply returned error: %v", err)
	}
	want := filepath.Join(tmp, ".claude/skills/buckle/SKILL.md")
	data, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("expected file at %s: %v", want, err)
	}
	if !strings.Contains(string(data), "# Buckle") {
		t.Error("written file should contain skill heading")
	}
}

func TestInstallRefusesOverwrite(t *testing.T) {
	tmp := t.TempDir()
	if _, _, err := runCLI("install", "--target", "claude", "--apply", tmp); err != nil {
		t.Fatalf("first apply failed: %v", err)
	}
	_, _, err := runCLI("install", "--target", "claude", "--apply", tmp)
	if err == nil {
		t.Fatal("second apply should have failed (file already exists)")
	}
	if !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Errorf("expected refuse-to-overwrite error, got: %v", err)
	}
}

func TestInstallForceOverwrites(t *testing.T) {
	tmp := t.TempDir()
	if _, _, err := runCLI("install", "--target", "claude", "--apply", tmp); err != nil {
		t.Fatalf("first apply failed: %v", err)
	}
	// Without --force, a second apply still refuses, and the error points at --force.
	_, _, err := runCLI("install", "--target", "claude", "--apply", tmp)
	if err == nil {
		t.Fatal("second apply without --force should still refuse")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("refusal should mention --force for discoverability, got: %v", err)
	}
	// With --force, it overwrites and says so.
	out, _, err := runCLI("install", "--target", "claude", "--apply", "--force", tmp)
	if err != nil {
		t.Fatalf("--force apply should overwrite, got: %v", err)
	}
	if !strings.Contains(out, "Overwrote") {
		t.Errorf("expected overwrite confirmation in output, got:\n%s", out)
	}
	want := filepath.Join(tmp, ".claude/skills/buckle/SKILL.md")
	if _, statErr := os.Stat(want); statErr != nil {
		t.Fatalf("file should still exist after force overwrite: %v", statErr)
	}
}

func TestInstallDirOverridesBaseDir(t *testing.T) {
	tmp := t.TempDir()
	_, _, err := runCLI("install", "--target", "claude", "--apply", "--dir", tmp)
	if err != nil {
		t.Fatalf("--dir apply failed: %v", err)
	}
	// With --dir, claude target writes to <dir>/skills/buckle/SKILL.md, NOT
	// <dir>/.claude/skills/buckle/SKILL.md. This is the user's "install in
	// ~/.agents/skills" use case.
	want := filepath.Join(tmp, "skills/buckle/SKILL.md")
	if _, err := os.Stat(want); err != nil {
		t.Errorf("expected --dir to write to %s, but stat: %v", want, err)
	}
	notWant := filepath.Join(tmp, ".claude/skills/buckle/SKILL.md")
	if _, err := os.Stat(notWant); !os.IsNotExist(err) {
		t.Errorf("--dir should not have written to the tool-prefixed path %s", notWant)
	}
}

func TestInstallPathOverridesEverything(t *testing.T) {
	tmp := t.TempDir()
	want := filepath.Join(tmp, "anywhere/buckle.md")
	_, _, err := runCLI("install", "--target", "cursor", "--apply", "--path", want)
	if err != nil {
		t.Fatalf("--path apply failed: %v", err)
	}
	data, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("expected file at %s: %v", want, err)
	}
	// --path keeps the target's body format. Cursor target = MDC frontmatter.
	if !strings.HasPrefix(string(data), "---\ndescription:") {
		t.Error("--path should keep the cursor body format (MDC frontmatter)")
	}
}

func TestInstallPathConflictsWithDir(t *testing.T) {
	_, _, err := runCLI("install", "--target", "claude", "--path", "/tmp/x", "--dir", "/tmp/y")
	if err == nil {
		t.Fatal("--path with --dir should error")
	}
	if !strings.Contains(err.Error(), "--path is a full destination override") {
		t.Errorf("expected combo-conflict error, got: %v", err)
	}
}

func TestInstallPathConflictsWithGlobal(t *testing.T) {
	_, _, err := runCLI("install", "--target", "claude", "--path", "/tmp/x", "--global")
	if err == nil {
		t.Fatal("--path with --global should error")
	}
}

func TestInstallPathConflictsWithPositional(t *testing.T) {
	_, _, err := runCLI("install", "--target", "claude", "--path", "/tmp/x", "/some/repo")
	if err == nil {
		t.Fatal("--path with positional path should error")
	}
}

func TestInstallGlobalRefusesPositional(t *testing.T) {
	_, _, err := runCLI("install", "--target", "claude", "--global", "/some/repo")
	if err == nil {
		t.Fatal("--global with positional path should error")
	}
}

func TestInstallGlobalUnsupportedTarget(t *testing.T) {
	// copilot has no documented global location.
	_, _, err := runCLI("install", "--target", "copilot", "--global")
	if err == nil {
		t.Fatal("--global on copilot should error")
	}
	if !strings.Contains(err.Error(), "no documented global location") {
		t.Errorf("expected 'no documented global location' error, got: %v", err)
	}
}

func TestInstallUnknownTarget(t *testing.T) {
	_, _, err := runCLI("install", "--target", "nonexistent")
	if err == nil {
		t.Fatal("unknown target should error")
	}
	if !strings.Contains(err.Error(), "unknown install target") {
		t.Errorf("expected 'unknown install target' error, got: %v", err)
	}
}

func TestInstallSharedFileWarning(t *testing.T) {
	tmp := t.TempDir()
	out, errOut, err := runCLI("install", "--target", "codex", tmp)
	if err != nil {
		t.Fatalf("dry-run errored: %v", err)
	}
	// Note about shared file goes to stderr.
	if !strings.Contains(errOut, "which other agent tools may also read") {
		t.Errorf("expected shared-file warning on stderr, got stderr=%q stdout=%q", errOut, out)
	}
}

func TestInstallExpandHomeInPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("could not resolve home dir")
	}
	out, _, err := runCLI("install", "--target", "claude", "--path", "~/some-test-skill.md")
	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}
	want := filepath.Join(home, "some-test-skill.md")
	if !strings.Contains(out, want) {
		t.Errorf("expected expanded path %s in output, got:\n%s", want, out)
	}
}
