// Package audio fetches ayah recordings and plays them.
package audio

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/4thel00z/quran/internal/quran"
)

// Fetcher downloads ayah MP3s from BaseURL and keeps them in CacheDir.
type Fetcher struct {
	BaseURL  string
	CacheDir string
	Client   *http.Client

	mu     sync.Mutex
	recent map[string][]byte
	order  []string
}

const recentLimit = 16

func NewFetcher(baseURL string, cacheDir string) *Fetcher {
	return &Fetcher{
		BaseURL:  baseURL,
		CacheDir: cacheDir,
		Client:   &http.Client{Timeout: 60 * time.Second},
		recent:   map[string][]byte{},
	}
}

func (f *Fetcher) Fetch(ctx context.Context, r quran.Reciter, key quran.Key) ([]byte, error) {
	id := r.Slug + "/" + quran.AudioFile(key)
	if data, ok := f.remembered(id); ok {
		return data, nil
	}
	path := f.cachePath(id)
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		f.remember(id, data)
		return data, nil
	}
	data, err := f.download(ctx, r.AudioURL(f.BaseURL, key))
	if err != nil {
		return nil, err
	}
	f.remember(id, data)
	if path != "" {
		f.store(path, data)
	}
	return data, nil
}

func (f *Fetcher) download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func (f *Fetcher) cachePath(id string) string {
	if f.CacheDir == "" {
		return ""
	}
	return filepath.Join(f.CacheDir, filepath.FromSlash(id))
}

// store writes the cache entry; a failure only costs a later re-download.
func (f *Fetcher) store(path string, data []byte) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	partial := path + ".part"
	if err := os.WriteFile(partial, data, 0o644); err != nil {
		return
	}
	_ = os.Rename(partial, path)
}

func (f *Fetcher) remembered(id string) ([]byte, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	data, ok := f.recent[id]
	return data, ok
}

func (f *Fetcher) remember(id string, data []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.recent[id]; ok {
		return
	}
	f.recent[id] = data
	f.order = append(f.order, id)
	if len(f.order) <= recentLimit {
		return
	}
	delete(f.recent, f.order[0])
	f.order = f.order[1:]
}
