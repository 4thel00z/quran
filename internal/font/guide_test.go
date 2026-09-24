package font

import (
	"strings"
	"testing"
)

func TestTerminalGuide(t *testing.T) {
	guide := TerminalGuide()
	for _, expected := range []string{"kitty", "wezterm", "alacritty", "ghostty", "Amiri Quran"} {
		if !strings.Contains(strings.ToLower(guide), strings.ToLower(expected)) {
			t.Errorf("expected guide to mention %q", expected)
		}
	}
}
