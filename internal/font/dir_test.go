package font

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestTargetDirOverride(t *testing.T) {
	custom := "/tmp/custom-fonts"
	dir, err := TargetDir(custom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir != custom {
		t.Errorf("expected %s, got %s", custom, dir)
	}
}

func TestTargetDirDefault(t *testing.T) {
	dir, err := TargetDir("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir == "" {
		t.Fatal("expected non-empty target directory")
	}
	switch runtime.GOOS {
	case "linux":
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("cannot get user home dir")
		}
		expected := filepath.Join(home, ".local", "share", "fonts", "quran")
		if dir != expected {
			t.Errorf("expected %s, got %s", expected, dir)
		}
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("cannot get user home dir")
		}
		expected := filepath.Join(home, "Library", "Fonts")
		if dir != expected {
			t.Errorf("expected %s, got %s", expected, dir)
		}
	}
}
