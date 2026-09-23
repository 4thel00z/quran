// Command gen packs Tarteel's quran-assets and per-reciter word timings into
// the gzip JSON files embedded by internal/quran.
package main

import (
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/4thel00z/quran/internal/quran"
)

const (
	pageCount     = 604
	segmentAPIURL = "https://api.quran.com/api/v4/verses/by_page/%d?audio=%d&per_page=50&fields=verse_key"
)

type options struct {
	assets   string
	out      string
	segments bool
}

func main() {
	opts := options{}
	flag.StringVar(&opts.assets, "assets", "", "path to a TarteelAI/quran-assets checkout")
	flag.StringVar(&opts.out, "out", "internal/assets/data", "output directory")
	flag.BoolVar(&opts.segments, "segments", true, "fetch word timings for every reciter")
	flag.Parse()
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}

func run(opts options) error {
	if opts.assets == "" {
		return errors.New("-assets is required")
	}
	if err := os.MkdirAll(filepath.Join(opts.out, "translations"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(opts.out, "segments"), 0o755); err != nil {
		return err
	}
	book, err := buildBook(opts.assets)
	if err != nil {
		return fmt.Errorf("book: %w", err)
	}
	if err := writeGzipJSON(filepath.Join(opts.out, "quran.json.gz"), book); err != nil {
		return err
	}
	if err := writeTranslations(opts); err != nil {
		return fmt.Errorf("translations: %w", err)
	}
	if !opts.segments {
		return nil
	}
	return writeSegments(opts, book)
}

type rawSurahInfo struct {
	NameEn      string `json:"nameEn"`
	NameAr      string `json:"nameAr"`
	NameEnTrans string `json:"nameEnTrans"`
	NumAyahs    int    `json:"numAyahs"`
}

type rawStart struct {
	SurahNum int `json:"surahNum"`
	AyahNum  int `json:"ayahNum"`
}

type rawWord struct {
	Text *string `json:"text"`
}

type rawPage struct {
	Surahs []struct {
		SurahNum int `json:"surahNum"`
		Ayahs    []struct {
			AyahNum int       `json:"ayahNum"`
			Words   []rawWord `json:"words"`
		} `json:"ayahs"`
	} `json:"surahs"`
}

func buildBook(assets string) (quran.Book, error) {
	infos := map[string]rawSurahInfo{}
	if err := readJSON(filepath.Join(assets, "metadata/surah-info.json"), &infos); err != nil {
		return quran.Book{}, err
	}
	juzStarts := map[string]rawStart{}
	if err := readJSON(filepath.Join(assets, "metadata/juz-info.json"), &juzStarts); err != nil {
		return quran.Book{}, err
	}
	hizbStarts := map[string]rawStart{}
	if err := readJSON(filepath.Join(assets, "metadata/hizb-info.json"), &hizbStarts); err != nil {
		return quran.Book{}, err
	}
	clean := map[string]map[string]string{}
	if err := readJSON(filepath.Join(assets, "text/quran-simple-clean.json"), &clean); err != nil {
		return quran.Book{}, err
	}
	revelation, err := readRevelation(filepath.Join(assets, "metadata/revelation-data.csv"))
	if err != nil {
		return quran.Book{}, err
	}
	words, pages, err := readPages(assets)
	if err != nil {
		return quran.Book{}, err
	}

	book := quran.Book{}
	for n := 1; n <= quran.SurahCount; n++ {
		info := infos[strconv.Itoa(n)]
		rev := revelation[n]
		book.Surahs = append(book.Surahs, quran.Surah{
			Number:          n,
			NameArabic:      info.NameAr,
			NameEnglish:     info.NameEn,
			NameMeaning:     info.NameEnTrans,
			Revelation:      rev.place,
			RevelationOrder: rev.order,
			AyahCount:       info.NumAyahs,
		})
	}

	juzAt := startIndex(juzStarts)
	hizbAt := startIndex(hizbStarts)
	juz, hizb := 0, 0
	for _, surah := range book.Surahs {
		for a := 1; a <= surah.AyahCount; a++ {
			key := quran.Key{Surah: surah.Number, Ayah: a}
			if next, ok := juzAt[key]; ok {
				juz = next
			}
			if next, ok := hizbAt[key]; ok {
				hizb = next
			}
			ayahWords := words[key]
			if len(ayahWords) == 0 {
				return quran.Book{}, fmt.Errorf("no words for %s", key)
			}
			book.Ayahs = append(book.Ayahs, quran.Ayah{
				Key:   key,
				Juz:   juz,
				Hizb:  hizb,
				Page:  pages[key],
				Words: ayahWords,
				Clean: clean[strconv.Itoa(surah.Number)][strconv.Itoa(a)],
			})
		}
	}
	if len(book.Ayahs) != quran.AyahCount {
		return quran.Book{}, fmt.Errorf("got %d ayahs, want %d", len(book.Ayahs), quran.AyahCount)
	}
	return book, nil
}

func startIndex(starts map[string]rawStart) map[quran.Key]int {
	index := map[quran.Key]int{}
	for n, start := range starts {
		number, err := strconv.Atoi(n)
		if err != nil {
			continue
		}
		index[quran.Key{Surah: start.SurahNum, Ayah: start.AyahNum}] = number
	}
	return index
}

type revelationInfo struct {
	order int
	place string
}

func readRevelation(path string) (map[int]revelationInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	result := map[int]revelationInfo{}
	for _, row := range rows {
		if len(row) < 4 {
			continue
		}
		order, err := strconv.Atoi(strings.TrimPrefix(row[0], "\ufeff"))
		if err != nil {
			return nil, fmt.Errorf("revelation order %q: %w", row[0], err)
		}
		surah, err := strconv.Atoi(row[2])
		if err != nil {
			return nil, fmt.Errorf("revelation surah %q: %w", row[2], err)
		}
		result[surah] = revelationInfo{order: order, place: row[3]}
	}
	return result, nil
}

func readPages(assets string) (map[quran.Key][]string, map[quran.Key]int, error) {
	words := map[quran.Key][]string{}
	pages := map[quran.Key]int{}
	for p := 1; p <= pageCount; p++ {
		page := rawPage{}
		if err := readJSON(filepath.Join(assets, "pages", strconv.Itoa(p)+".json"), &page); err != nil {
			return nil, nil, err
		}
		for _, surah := range page.Surahs {
			for _, ayah := range surah.Ayahs {
				key := quran.Key{Surah: surah.SurahNum, Ayah: ayah.AyahNum}
				if _, seen := pages[key]; !seen {
					pages[key] = p
				}
				for _, word := range ayah.Words {
					if word.Text == nil {
						continue
					}
					words[key] = append(words[key], *word.Text)
				}
			}
		}
	}
	return words, pages, nil
}

func writeTranslations(opts options) error {
	for _, t := range quran.Translations {
		verses := [][]string{}
		if err := readJSON(filepath.Join(opts.assets, "translations/tanzil", t.ID+".json"), &verses); err != nil {
			return err
		}
		if len(verses) != quran.SurahCount {
			return fmt.Errorf("%s: %d surahs", t.ID, len(verses))
		}
		if err := writeGzipJSON(filepath.Join(opts.out, "translations", t.ID+".json.gz"), verses); err != nil {
			return err
		}
	}
	return nil
}

func writeSegments(opts options, book quran.Book) error {
	for _, r := range quran.Reciters {
		path := filepath.Join(opts.out, "segments", r.Slug+".json.gz")
		if _, err := os.Stat(path); err == nil {
			fmt.Println("skip", r.Slug)
			continue
		}
		segments, err := reciterSegments(opts.assets, book, r)
		if err != nil {
			return fmt.Errorf("%s: %w", r.Slug, err)
		}
		if err := writeGzipJSON(path, segments); err != nil {
			return err
		}
		fmt.Println("wrote", r.Slug)
	}
	return nil
}

func reciterSegments(assets string, book quran.Book, r quran.Reciter) (quran.Segments, error) {
	if r.Slug == "alafasy" {
		return tarteelAlafasy(assets)
	}
	return fetchSegments(book, r.QuranComID)
}

// tarteelAlafasy converts Tarteel's [word, start, end] triples to the
// [firstWord, lastWord, start, end] layout used for every reciter.
func tarteelAlafasy(assets string) (quran.Segments, error) {
	raw := [][][][3]int{}
	if err := readJSON(filepath.Join(assets, "audio/alafasy-ayah-manifest.json"), &raw); err != nil {
		return nil, err
	}
	return mapSurahs(raw, func(t [3]int) quran.Segment {
		return quran.Segment{t[0] - 1, t[0], t[1], t[2]}
	}), nil
}

func mapSurahs(raw [][][][3]int, convert func([3]int) quran.Segment) [][][]quran.Segment {
	return mapSlice(raw, func(surah [][][3]int) [][]quran.Segment {
		return mapSlice(surah, func(ayah [][3]int) []quran.Segment {
			return mapSlice(ayah, convert)
		})
	})
}

func mapSlice[T, U any](items []T, convert func(T) U) []U {
	result := make([]U, 0, len(items))
	for _, item := range items {
		result = append(result, convert(item))
	}
	return result
}

type apiPage struct {
	Verses []struct {
		VerseKey string `json:"verse_key"`
		Audio    struct {
			Segments [][]json.Number `json:"segments"`
		} `json:"audio"`
	} `json:"verses"`
}

func fetchSegments(book quran.Book, recitation int) (quran.Segments, error) {
	segments := make(quran.Segments, quran.SurahCount)
	for i, surah := range book.Surahs {
		segments[i] = make([][]quran.Segment, surah.AyahCount)
	}
	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		firstErr error
		pages    = make(chan int)
	)
	for range 8 {
		wg.Go(func() {
			for p := range pages {
				page, err := fetchPage(p, recitation)
				mu.Lock()
				if err != nil && firstErr == nil {
					firstErr = fmt.Errorf("page %d: %w", p, err)
				}
				for _, verse := range page.Verses {
					key, err := quran.ParseKey(verse.VerseKey)
					if err != nil {
						continue
					}
					segments[key.Surah-1][key.Ayah-1] = toSegments(verse.Audio.Segments)
				}
				mu.Unlock()
			}
		})
	}
	for p := 1; p <= pageCount; p++ {
		pages <- p
	}
	close(pages)
	wg.Wait()
	return segments, firstErr
}

func toSegments(raw [][]json.Number) []quran.Segment {
	result := []quran.Segment{}
	for _, s := range raw {
		if len(s) != 4 {
			continue
		}
		segment := quran.Segment{}
		valid := true
		for i, n := range s {
			value, err := n.Float64()
			valid = valid && err == nil
			segment[i] = int(value)
		}
		if !valid {
			continue
		}
		result = append(result, segment)
	}
	return result
}

func fetchPage(page int, recitation int) (apiPage, error) {
	url := fmt.Sprintf(segmentAPIURL, page, recitation)
	var lastErr error
	for attempt := range 5 {
		result, err := getPage(url)
		if err == nil {
			return result, nil
		}
		lastErr = err
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}
	return apiPage{}, lastErr
}

func getPage(url string) (apiPage, error) {
	resp, err := http.Get(url)
	if err != nil {
		return apiPage{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return apiPage{}, fmt.Errorf("%s: %s", url, resp.Status)
	}
	page := apiPage{}
	return page, json.NewDecoder(resp.Body).Decode(&page)
}

func readJSON(path string, target any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(target); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

func writeGzipJSON(path string, value any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	zw, err := gzip.NewWriterLevel(f, gzip.BestCompression)
	if err != nil {
		return errors.Join(err, f.Close())
	}
	encodeErr := json.NewEncoder(zw).Encode(value)
	return errors.Join(encodeErr, zw.Close(), f.Close())
}
