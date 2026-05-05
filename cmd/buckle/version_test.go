package main

import (
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	out, _, err := runCLI("version")
	if err != nil {
		t.Fatalf("version returned error: %v", err)
	}
	if !strings.HasPrefix(out, "buckle ") {
		t.Errorf("expected output to start with %q, got:\n%s", "buckle ", out)
	}
	for _, want := range []string{"commit:", "built:"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}
