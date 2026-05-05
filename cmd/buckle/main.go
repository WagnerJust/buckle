// Command buckle installs the buckle skill — a Hub-and-Spoke agent
// documentation guide — into the location expected by various coding-
// agent tools (Claude Code, Cursor, Codex CLI, etc.).
//
// The single source of truth is buckle/SKILL.md at the repo root,
// which is embedded into the binary at build time. The install
// command writes that content (with a per-tool frontmatter or
// preamble) wherever the target tool expects to find it. Default
// is dry-run; --apply commits.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// Exit codes are part of the CLI contract.
const (
	exitCannotRun = 2
	exitInternal  = 4
)

// exitError lets a subcommand return an error annotated with the exit
// code that should be reported to the shell.
type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string { return e.err.Error() }
func (e *exitError) Unwrap() error { return e.err }

func errCannotRun(err error) error { return &exitError{exitCannotRun, err} }

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "buckle:", err)
		var ee *exitError
		if errors.As(err, &ee) {
			os.Exit(ee.code)
		}
		os.Exit(exitInternal)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	root := &cobra.Command{
		Use:           "buckle",
		Short:         "Install the buckle skill into your coding agent's expected location.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.AddCommand(newInstallCmd(stdout, stderr))
	root.SetArgs(args)
	return root.Execute()
}
