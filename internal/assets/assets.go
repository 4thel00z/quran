// Package assets loads the Quran text, translations and word timings
// embedded at build time from Tarteel's quran-assets.
package assets

import (
	"compress/gzip"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sync"

	"github.com/4thel00z/quran/internal/quran"
)

//go:embed data
var files embed.FS

var loadBook = sync.OnceValues(func() (quran.Book, error) {
	book := quran.Book{}
	return book, decode("data/quran.json.gz", &book)
})

func Book() (quran.Book, error) {
	return loadBook()
}

// Translation returns the verses of one translation per surah, per ayah.
func Translation(id string) ([][]string, error) {
	if _, err := quran.FindTranslation(id); err != nil {
		return nil, err
	}
	verses := [][]string{}
	return verses, decode(path.Join("data/translations", id+".json.gz"), &verses)
}

func Segments(slug string) (quran.Segments, error) {
	if _, err := quran.FindReciter(slug); err != nil {
		return nil, err
	}
	segments := quran.Segments{}
	return segments, decode(path.Join("data/segments", slug+".json.gz"), &segments)
}

func decode(name string, target any) error {
	f, err := files.Open(name)
	if err != nil {
		return err
	}
	zr, err := gzip.NewReader(f)
	if err != nil {
		return errors.Join(fmt.Errorf("%s: %w", name, err), f.Close())
	}
	decodeErr := json.NewDecoder(zr).Decode(target)
	if decodeErr != nil {
		decodeErr = fmt.Errorf("%s: %w", name, decodeErr)
	}
	return errors.Join(decodeErr, zr.Close(), f.Close())
}
