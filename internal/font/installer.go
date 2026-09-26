package font

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// StatusItem contains the installation status of a single font.
type StatusItem struct {
	Font      Font
	Installed bool
	Path      string
	Size      int64
}

// ProgressFunc reports download progress for a font file.
type ProgressFunc func(name string, written int64, total int64)

// CheckStatus reports whether the recommended fonts are installed in targetDir.
func CheckStatus(targetDir string) ([]StatusItem, error) {
	return checkStatus(targetDir, RecommendedFonts())
}

func checkStatus(targetDir string, fonts []Font) ([]StatusItem, error) {
	items := make([]StatusItem, len(fonts))
	for i, f := range fonts {
		destPath := filepath.Join(targetDir, f.FileName)
		info, err := os.Stat(destPath)
		installed := err == nil && !info.IsDir() && info.Size() > 0
		var sz int64
		if installed {
			sz = info.Size()
		}
		items[i] = StatusItem{
			Font:      f,
			Installed: installed,
			Path:      destPath,
			Size:      sz,
		}
	}
	return items, nil
}

// Install downloads and installs all recommended fonts to targetDir.
// If force is true, existing fonts are overwritten.
func Install(ctx context.Context, targetDir string, force bool, onProgress ProgressFunc) error {
	return installFonts(ctx, targetDir, force, RecommendedFonts(), onProgress)
}

func installFonts(ctx context.Context, targetDir string, force bool, fonts []Font, onProgress ProgressFunc) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create font directory %s: %w", targetDir, err)
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	for _, f := range fonts {
		destPath := filepath.Join(targetDir, f.FileName)
		if !force {
			info, err := os.Stat(destPath)
			if err == nil && info.Size() > 0 {
				// Already installed
				continue
			}
		}

		if err := downloadFont(ctx, client, f, destPath, onProgress); err != nil {
			return fmt.Errorf("failed to download font %s: %w", f.Name, err)
		}
	}

	return nil
}

type progressWriter struct {
	name       string
	total      int64
	written    int64
	onProgress ProgressFunc
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.written += int64(n)
	if pw.onProgress != nil {
		pw.onProgress(pw.name, pw.written, pw.total)
	}
	return n, nil
}

func downloadFont(ctx context.Context, client *http.Client, f Font, destPath string, onProgress ProgressFunc) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "quran-tui/font-installer")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	total := resp.ContentLength
	if total <= 0 {
		total = f.Size
	}

	tmpFile := destPath + ".tmp"
	out, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	pw := &progressWriter{
		name:       f.Name,
		total:      total,
		onProgress: onProgress,
	}

	_, copyErr := io.Copy(out, io.TeeReader(resp.Body, pw))
	closeErr := out.Close()

	if copyErr != nil {
		_ = os.Remove(tmpFile)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpFile)
		return closeErr
	}

	if err := os.Rename(tmpFile, destPath); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}

	return nil
}
