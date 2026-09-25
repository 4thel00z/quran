package config

import (
	"path/filepath"
	"testing"
)

func TestPath(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", "")

	p, err := Path()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(tempHome, ".config", "quran", "config.json")
	if p != expected {
		t.Errorf("expected %s, got %s", expected, p)
	}

	// Test XDG_CONFIG_HOME override
	customXDG := filepath.Join(tempHome, "custom-config")
	t.Setenv("XDG_CONFIG_HOME", customXDG)

	p, err = Path()
	if err != nil {
		t.Fatalf("unexpected error with XDG: %v", err)
	}
	expected = filepath.Join(customXDG, "quran", "config.json")
	if p != expected {
		t.Errorf("expected %s, got %s", expected, p)
	}
}

func TestLoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentFile := filepath.Join(tempDir, "does-not-exist.json")

	cfg, err := LoadFrom(nonExistentFile)
	if err != nil {
		t.Fatalf("loading non-existent file should not error: %v", err)
	}
	if cfg.Reciter != "" || cfg.Translation != "" {
		t.Errorf("expected empty config, got %+v", cfg)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "sub", "config.json")

	showSidebar := true
	showTranslation := false
	original := Config{
		Reciter:         "husary",
		Translation:     "en-sahih",
		Arabic:          "native",
		ShowSidebar:     &showSidebar,
		ShowTranslation: &showTranslation,
		Repeat:          "ayah",
		Volume:          0.75,
	}

	if err := SaveTo(configFile, original); err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	loaded, err := LoadFrom(configFile)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if loaded.Reciter != original.Reciter {
		t.Errorf("expected Reciter %q, got %q", original.Reciter, loaded.Reciter)
	}
	if loaded.Translation != original.Translation {
		t.Errorf("expected Translation %q, got %q", original.Translation, loaded.Translation)
	}
	if loaded.Arabic != original.Arabic {
		t.Errorf("expected Arabic %q, got %q", original.Arabic, loaded.Arabic)
	}
	if loaded.ShowSidebar == nil || *loaded.ShowSidebar != true {
		t.Errorf("expected ShowSidebar true, got %v", loaded.ShowSidebar)
	}
	if loaded.ShowTranslation == nil || *loaded.ShowTranslation != false {
		t.Errorf("expected ShowTranslation false, got %v", loaded.ShowTranslation)
	}
	if loaded.Repeat != "ayah" {
		t.Errorf("expected Repeat ayah, got %q", loaded.Repeat)
	}
	if loaded.Volume != 0.75 {
		t.Errorf("expected Volume 0.75, got %v", loaded.Volume)
	}
}
