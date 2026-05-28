package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/WagnerJust/buckle/internal/skill"
)

func newInstallCmd(stdout, stderr io.Writer) *cobra.Command {
	var (
		apply  bool
		target string
		list   bool
		global bool
		dir    string
		path   string
		force  bool
	)

	cmd := &cobra.Command{
		Use:   "install [path]",
		Short: "Install the buckle skill into a coding-agent tool's expected location.",
		Long: `buckle install — write the buckle skill into the location expected
by the chosen coding-agent tool.

Supported targets (--target):

  claude    Claude Code             .claude/skills/buckle/SKILL.md
  cursor    Cursor                  .cursor/rules/buckle.mdc
  gemini    Gemini Code Assist      GEMINI.md                       (shared)
  codex     OpenAI Codex CLI        AGENTS.md                       (shared)
  opencode  OpenCode                AGENTS.md                       (shared)
  windsurf  Windsurf                .windsurfrules                  (shared)
  cline     Cline                   .clinerules                     (shared)
  copilot   GitHub Copilot          .github/copilot-instructions.md (shared)

"Shared" targets write to a file the user might already be using for
other purposes. The install refuses to overwrite, so an existing file
yields a clear error; pass --force to overwrite it anyway.

Scope: project-local by default (the file lands in the current repo or
the [path] argument). Pass --global to install at user scope under
$HOME instead. Pass --dir to override the base directory while keeping
the tool's skills-relative tail (e.g., --dir ~/.agents lands the
Claude target at ~/.agents/skills/buckle/SKILL.md). Pass --path for a
fully-specified destination — the target still controls the body
format.

Default is dry-run; --apply is required to actually write. Use --list
to print just the available targets.

Examples:
  # See what would be written for the default target (claude).
  buckle install

  # Install for Cursor in this repo.
  buckle install --target cursor --apply

  # Install Claude skill globally at ~/.agents/skills/buckle/SKILL.md.
  buckle install --target claude --global --dir ~/.agents --apply

  # Write to a fully-specified path with Claude's body format.
  buckle install --target claude --path ~/Skills/buckle.md --apply

  # Re-install over an existing file (e.g., after updating the skill).
  buckle install --target claude --global --apply --force

Exit codes:
  0  installed (or dry-run / --list completed)
  2  could not run (existing file, path bad, unknown target, bad flag combo)
  3  configuration error`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if list {
				printTargetList(stdout)
				return nil
			}
			if target == "" {
				target = "claude"
			}

			t, ok := skill.TargetByID(target)
			if !ok {
				return errCannotRun(fmt.Errorf("unknown install target %q (available: %s)",
					target, strings.Join(skill.IDs(), ", ")))
			}

			destPath, mode, err := resolveDestination(t, args, global, dir, path)
			if err != nil {
				return errCannotRun(err)
			}

			if t.SharedFile && path == "" {
				fmt.Fprintf(stderr, "Note: %s installs to %s, which other agent tools may also read.\n",
					t.Name, destPath)
			}

			existed := fileExists(destPath)

			if !apply {
				fmt.Fprintf(stdout, "buckle install (dry-run)\n  target: %s (%s)\n  scope:  %s\n  path:   %s\n  bytes:  %d\n",
					t.ID, t.Name, mode, destPath, len(t.Body()))
				if existed {
					if force {
						fmt.Fprintln(stdout, "  exists: yes — will overwrite (--force)")
					} else {
						fmt.Fprintln(stdout, "  exists: yes — re-run with --force to overwrite")
					}
				}
				fmt.Fprintln(stdout, "\nNothing was written. Re-run with --apply to commit.")
				return nil
			}

			if err := writeSkill(destPath, []byte(t.Body()), force); err != nil {
				return errCannotRun(err)
			}
			verb := "Wrote"
			if existed {
				verb = "Overwrote"
			}
			fmt.Fprintf(stdout, "%s %s (%d bytes).\n", verb, destPath, len(t.Body()))
			return nil
		},
	}

	f := cmd.Flags()
	f.BoolVar(&apply, "apply", false, "Actually write the skill (default is dry-run).")
	f.StringVar(&target, "target", "", fmt.Sprintf("Coding-agent tool. One of: %s. Default: claude.", strings.Join(skill.IDs(), ", ")))
	f.BoolVar(&list, "list", false, "List available install targets and exit.")
	f.BoolVar(&global, "global", false, "Install at user scope (under $HOME) instead of the current repo.")
	f.StringVar(&dir, "dir", "", "Override the base directory (replaces $HOME for --global, or the repo root for project scope).")
	f.StringVar(&path, "path", "", "Fully-specified destination path. Target still controls the body format. Conflicts with --global, --dir, and the positional [path] argument.")
	f.BoolVarP(&force, "force", "f", false, "Overwrite the destination if it already exists. Without this, install refuses to clobber an existing file.")
	return cmd
}

// resolveDestination computes the absolute destination file path and a
// human-readable scope description from the user's flag combination.
//
// Conflicts are rejected up front: --path with any other path-affecting
// flag, --global with a positional path, and --global on a target with
// no documented user-scope location.
func resolveDestination(t skill.Target, args []string, global bool, dir, path string) (destPath, scope string, err error) {
	if path != "" {
		if global || dir != "" || len(args) > 0 {
			return "", "", errors.New("--path is a full destination override and cannot be combined with --global, --dir, or a positional path")
		}
		abs, absErr := filepath.Abs(expandHome(path))
		if absErr != nil {
			return "", "", absErr
		}
		return abs, "user-specified path", nil
	}

	if global {
		if len(args) > 0 {
			return "", "", errors.New("--global installs under $HOME (or --dir); do not pass a positional path")
		}
		if !t.SupportsGlobal() {
			return "", "", fmt.Errorf("target %q has no documented global location; install at project scope (drop --global) or use one of: %s",
				t.ID, strings.Join(skill.GlobalCapableIDs(), ", "))
		}
		base, err := globalBase(dir)
		if err != nil {
			return "", "", err
		}
		var rel string
		if dir != "" {
			rel = t.DirTail
		} else {
			rel = t.GlobalPath
		}
		return filepath.Join(base, rel), describeGlobalScope(base, dir), nil
	}

	// Project scope (default).
	repo := "."
	if len(args) == 1 {
		repo = args[0]
	}
	if dir != "" {
		// --dir wins over the positional / repo path.
		base, baseErr := filepath.Abs(expandHome(dir))
		if baseErr != nil {
			return "", "", baseErr
		}
		return filepath.Join(base, t.DirTail), fmt.Sprintf("at %s (--dir)", base), nil
	}
	abs, absErr := filepath.Abs(repo)
	if absErr != nil {
		return "", "", absErr
	}
	return filepath.Join(abs, t.Path), fmt.Sprintf("at %s (repo)", abs), nil
}

func globalBase(dir string) (string, error) {
	if dir != "" {
		abs, err := filepath.Abs(expandHome(dir))
		if err != nil {
			return "", err
		}
		return abs, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve user home dir: %w", err)
	}
	return home, nil
}

func describeGlobalScope(base, dir string) string {
	if dir != "" {
		return fmt.Sprintf("user-scope override at %s", base)
	}
	return fmt.Sprintf("user-scope at %s", base)
}

// expandHome rewrites a leading "~/" in a path to the user's home
// directory. Other "~" patterns (e.g., "~user/...") are left alone.
func expandHome(p string) string {
	if !strings.HasPrefix(p, "~/") && p != "~" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	if p == "~" {
		return home
	}
	return filepath.Join(home, p[2:])
}

// writeSkill writes data to path, creating any missing parent directories.
// By default it refuses to overwrite an existing file — the O_EXCL flag is
// the safety guarantee, so buckle never silently clobbers a user's content.
// With force, it truncates the destination instead, which is an explicit
// opt-in to overwrite.
func writeSkill(path string, data []byte, force bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if force {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	f, err := os.OpenFile(path, flags, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("refusing to overwrite existing file: %s (use --force to overwrite)", path)
		}
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

// fileExists reports whether path already names a file or directory. Used
// only for messaging and dry-run hints; the real overwrite guard is the
// O_EXCL open in writeSkill.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func printTargetList(w io.Writer) {
	fmt.Fprintln(w, "Available install targets:")
	fmt.Fprintln(w)
	for _, t := range skill.Targets() {
		shared := ""
		if t.SharedFile {
			shared = " (shared file)"
		}
		fmt.Fprintf(w, "  %-10s %-20s %s%s\n", t.ID, t.Name, t.Path, shared)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Hint: buckle install --target <id> --apply")
}
