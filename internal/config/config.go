package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Config represents persistent user preferences for quran-tui.
type Config struct {
	Reciter         string   `json:"reciter,omitempty"`
	Translation     string   `json:"translation,omitempty"`
	Arabic          string   `json:"arabic,omitempty"`
	ShowSidebar     *bool    `json:"show_sidebar,omitempty"`
	ShowTranslation *bool    `json:"show_translation,omitempty"`
	Repeat          string   `json:"repeat,omitempty"`
	Volume          float64  `json:"volume,omitempty"`
}

// Path returns the path to the configuration file (~/.config/quran/config.json).
func Path() (string, error) {
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg != "" {
		return filepath.Join(xdg, "quran", "config.json"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user home directory: %w", err)
	}

	return filepath.Join(home, ".config", "quran", "config.json"), nil
}

// Load reads the configuration from the default path.
// If the file does not exist, an empty Config and nil error are returned.
func Load() (Config, error) {
	p, err := Path()
	if err != nil {
		return Config{}, err
	}
	return LoadFrom(p)
}

// LoadFrom reads the configuration from the specified path.
func LoadFrom(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config %s: %w", path, err)
	}

	return cfg, nil
}

// Save writes the configuration to the default path.
func Save(cfg Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	return SaveTo(p, cfg)
}

// SaveTo writes the configuration to the specified path atomically.
func SaveTo(path string, cfg Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary config: %w", err)
	}

	if err := os.Rename(tmpFile, path); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to commit config %s: %w", path, err)
	}

	return nil
}
