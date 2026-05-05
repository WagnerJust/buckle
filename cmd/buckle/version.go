package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newVersionCmd(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print buckle version, commit, and build date.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(stdout, "buckle %s\ncommit: %s\nbuilt:  %s\n", version, commit, date)
			return nil
		},
	}
}
