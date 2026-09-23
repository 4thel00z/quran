package render

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestNativeLinesAreIsolated(t *testing.T) {
	plain := func(int) lipgloss.Style { return lipgloss.NewStyle() }
	lines := RTLLines([]string{"قُلْ", "هُوَ", "ٱللَّهُ"}, 30, Native, plain, lipgloss.NewStyle())
	if len(lines) != 1 {
		t.Fatalf("got %d lines", len(lines))
	}
	text := strings.TrimLeft(lines[0], " ")
	if !strings.HasPrefix(text, isolationMark) || !strings.HasSuffix(text, isolationMark) {
		t.Fatalf("line %q is not wrapped in LRM", lines[0])
	}
	if got := ansi.StringWidth(lines[0]); got != 30 {
		t.Fatalf("line is %d cells wide, want 30", got)
	}
}

func TestVisualLinesHaveNoMarks(t *testing.T) {
	plain := func(int) lipgloss.Style { return lipgloss.NewStyle() }
	lines := RTLLines([]string{"قُلْ", "هُوَ"}, 20, Visual, plain, lipgloss.NewStyle())
	if strings.Contains(lines[0], isolationMark) {
		t.Fatalf("visual line %q contains LRM", lines[0])
	}
}

func TestVersionAtLeast(t *testing.T) {
	cases := map[string]bool{"3.7.2": true, "3.6.0": true, "3.5.14": false, "4.0": true, "": false, "nightly": false}
	for version, want := range cases {
		if got := versionAtLeast(version, 3, 6); got != want {
			t.Errorf("versionAtLeast(%q) = %v", version, got)
		}
	}
}
