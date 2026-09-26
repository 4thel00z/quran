package font

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallAndStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "11")
		w.Write([]byte("dummy-font-"))
	}))
	defer ts.Close()

	tempDir := t.TempDir()

	testFonts := []Font{
		{
			Name:     "Test Font",
			FileName: "TestFont-Regular.ttf",
			URL:      ts.URL + "/test.ttf",
			Size:     11,
		},
	}

	// Status before install
	status, err := checkStatus(tempDir, testFonts)
	if err != nil {
		t.Fatalf("checkStatus failed: %v", err)
	}
	if len(status) != 1 || status[0].Installed {
		t.Errorf("expected font not to be installed yet")
	}

	// Install
	var progressCalls int
	err = installFonts(context.Background(), tempDir, false, testFonts, func(name string, written int64, total int64) {
		progressCalls++
	})
	if err != nil {
		t.Fatalf("installFonts failed: %v", err)
	}
	if progressCalls == 0 {
		t.Errorf("expected progress callback to be called")
	}

	// Verify file content
	content, err := os.ReadFile(filepath.Join(tempDir, "TestFont-Regular.ttf"))
	if err != nil {
		t.Fatalf("failed to read installed font: %v", err)
	}
	if string(content) != "dummy-font-" {
		t.Errorf("expected 'dummy-font-', got %q", string(content))
	}

	// Status after install
	status, err = checkStatus(tempDir, testFonts)
	if err != nil {
		t.Fatalf("checkStatus after install failed: %v", err)
	}
	if !status[0].Installed {
		t.Errorf("expected font to be installed")
	}

	// Re-installing without force should skip downloading
	progressCalls = 0
	err = installFonts(context.Background(), tempDir, false, testFonts, func(name string, written int64, total int64) {
		progressCalls++
	})
	if err != nil {
		t.Fatalf("installFonts second time failed: %v", err)
	}
	if progressCalls != 0 {
		t.Errorf("expected download to be skipped when already installed, got %d calls", progressCalls)
	}
}
