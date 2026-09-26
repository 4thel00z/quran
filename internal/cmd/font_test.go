package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestFontCommandHelp(t *testing.T) {
	root := newRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"font", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	for _, sub := range []string{"install", "status"} {
		if !strings.Contains(output, sub) {
			t.Errorf("expected help output to mention %q, got:\n%s", sub, output)
		}
	}
}

func TestFontStatusCommand(t *testing.T) {
	root := newRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"font", "status", "--dir", t.TempDir()})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Amiri Quran") {
		t.Errorf("expected output to mention Amiri Quran, got:\n%s", output)
	}
	if !strings.Contains(output, "Not installed") {
		t.Errorf("expected output to indicate Not installed, got:\n%s", output)
	}
}
