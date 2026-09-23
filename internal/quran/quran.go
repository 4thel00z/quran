// Package quran holds the Quran data model, reference parsing and search.
package quran

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	SurahCount = 114
	AyahCount  = 6236
	JuzCount   = 30
	HizbCount  = 60
	PageCount  = 604
)

type Key struct {
	Surah int `json:"s"`
	Ayah  int `json:"a"`
}

func (k Key) String() string {
	return fmt.Sprintf("%d:%d", k.Surah, k.Ayah)
}

func ParseKey(text string) (Key, error) {
	surah, ayah, found := strings.Cut(strings.TrimSpace(text), ":")
	if !found {
		return Key{}, fmt.Errorf("%q is not surah:ayah", text)
	}
	s, err := strconv.Atoi(strings.TrimSpace(surah))
	if err != nil {
		return Key{}, fmt.Errorf("surah %q: %w", surah, err)
	}
	a, err := strconv.Atoi(strings.TrimSpace(ayah))
	if err != nil {
		return Key{}, fmt.Errorf("ayah %q: %w", ayah, err)
	}
	return Key{Surah: s, Ayah: a}, nil
}

type Surah struct {
	Number          int    `json:"number"`
	NameArabic      string `json:"nameArabic"`
	NameEnglish     string `json:"nameEnglish"`
	NameMeaning     string `json:"nameMeaning"`
	Revelation      string `json:"revelation"`
	RevelationOrder int    `json:"revelationOrder"`
	AyahCount       int    `json:"ayahCount"`
}

type Ayah struct {
	Key
	Juz   int      `json:"juz"`
	Hizb  int      `json:"hizb"`
	Page  int      `json:"page"`
	Words []string `json:"words"`
	Clean string   `json:"clean"`
}

func (a Ayah) Text() string {
	return strings.Join(a.Words, " ")
}

type Book struct {
	Surahs []Surah `json:"surahs"`
	Ayahs  []Ayah  `json:"ayahs"`
}

var ErrNotFound = errors.New("not found")

// Index returns the position of key in b.Ayahs.
func (b Book) Index(key Key) (int, error) {
	if key.Surah < 1 || key.Surah > len(b.Surahs) {
		return 0, fmt.Errorf("surah %d: %w", key.Surah, ErrNotFound)
	}
	if key.Ayah < 1 || key.Ayah > b.Surahs[key.Surah-1].AyahCount {
		return 0, fmt.Errorf("ayah %s: %w", key, ErrNotFound)
	}
	index := key.Ayah - 1
	for _, s := range b.Surahs[:key.Surah-1] {
		index += s.AyahCount
	}
	return index, nil
}

func (b Book) Surah(number int) Surah {
	return b.Surahs[number-1]
}

// SurahAyahs returns the ayahs of one surah.
func (b Book) SurahAyahs(number int) []Ayah {
	first, err := b.Index(Key{Surah: number, Ayah: 1})
	if err != nil {
		return nil
	}
	return b.Ayahs[first : first+b.Surahs[number-1].AyahCount]
}

// Segment is [firstWord, lastWord, startMs, endMs]; words are the half-open
// range (firstWord, lastWord] counted from one.
type Segment [4]int

// Segments holds word timings per surah, per ayah.
type Segments [][][]Segment

func (s Segments) For(key Key) []Segment {
	if key.Surah < 1 || key.Surah > len(s) {
		return nil
	}
	surah := s[key.Surah-1]
	if key.Ayah < 1 || key.Ayah > len(surah) {
		return nil
	}
	return surah[key.Ayah-1]
}

// WordAt returns the zero-based word index recited at ms, or -1.
func WordAt(segments []Segment, ms int, wordCount int) int {
	for _, s := range segments {
		if ms < s[2] || ms >= s[3] {
			continue
		}
		return min(s[0], wordCount-1)
	}
	return -1
}
