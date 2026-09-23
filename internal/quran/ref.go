package quran

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type TargetKind int

const (
	TargetSurah TargetKind = iota
	TargetAyah
	TargetRange
	TargetJuz
	TargetHizb
	TargetPage
)

// Target is a resolved reference: the ayahs from Start to End inclusive.
type Target struct {
	Kind   TargetKind
	Number int
	Start  Key
	End    Key
}

func (t Target) String() string {
	switch t.Kind {
	case TargetJuz:
		return fmt.Sprintf("Juz %d", t.Number)
	case TargetHizb:
		return fmt.Sprintf("Hizb %d", t.Number)
	case TargetPage:
		return fmt.Sprintf("Page %d", t.Number)
	case TargetAyah:
		return t.Start.String()
	case TargetRange:
		return fmt.Sprintf("%s-%d", t.Start, t.End.Ayah)
	}
	return fmt.Sprintf("Surah %d", t.Number)
}

var (
	ayahPattern    = regexp.MustCompile(`^(\d{1,3})\s*[:.]\s*(\d{1,3})(?:\s*-\s*(\d{1,3}))?$`)
	sectionPattern = regexp.MustCompile(`^(juz|j|para|hizb|h|page|p)\s*[:#]?\s*(\d{1,3})$`)
)

// Resolve parses "18", "2:255", "2:255-257", "juz 30", "hizb 5", "page 50"
// or a surah name such as "kahf" or "الكهف".
func (b Book) Resolve(query string) (Target, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return Target{}, fmt.Errorf("empty reference: %w", ErrNotFound)
	}
	if m := ayahPattern.FindStringSubmatch(q); m != nil {
		return b.resolveAyahs(m[1], m[2], m[3])
	}
	if m := sectionPattern.FindStringSubmatch(q); m != nil {
		number, _ := strconv.Atoi(m[2])
		return b.resolveSection(sectionKind(m[1]), number)
	}
	if number, err := strconv.Atoi(q); err == nil {
		return b.surahTarget(number)
	}
	matches := b.MatchSurahs(q)
	if len(matches) == 0 {
		return Target{}, fmt.Errorf("%q: %w", query, ErrNotFound)
	}
	return b.surahTarget(matches[0].Number)
}

func sectionKind(word string) TargetKind {
	switch word {
	case "hizb", "h":
		return TargetHizb
	case "page", "p":
		return TargetPage
	}
	return TargetJuz
}

func (b Book) resolveAyahs(surah string, first string, last string) (Target, error) {
	s, _ := strconv.Atoi(surah)
	a, _ := strconv.Atoi(first)
	start := Key{Surah: s, Ayah: a}
	if _, err := b.Index(start); err != nil {
		return Target{}, err
	}
	if last == "" {
		return Target{Kind: TargetAyah, Number: s, Start: start, End: start}, nil
	}
	e, _ := strconv.Atoi(last)
	end := Key{Surah: s, Ayah: e}
	if _, err := b.Index(end); err != nil {
		return Target{}, err
	}
	if e < a {
		return Target{}, fmt.Errorf("range %s-%d runs backwards", start, e)
	}
	return Target{Kind: TargetRange, Number: s, Start: start, End: end}, nil
}

func (b Book) surahTarget(number int) (Target, error) {
	if number < 1 || number > len(b.Surahs) {
		return Target{}, fmt.Errorf("surah %d: %w", number, ErrNotFound)
	}
	return Target{
		Kind:   TargetSurah,
		Number: number,
		Start:  Key{Surah: number, Ayah: 1},
		End:    Key{Surah: number, Ayah: b.Surahs[number-1].AyahCount},
	}, nil
}

func (b Book) resolveSection(kind TargetKind, number int) (Target, error) {
	section := func(a Ayah) int {
		switch kind {
		case TargetHizb:
			return a.Hizb
		case TargetPage:
			return a.Page
		}
		return a.Juz
	}
	target := Target{Kind: kind, Number: number}
	found := false
	for _, a := range b.Ayahs {
		if section(a) != number {
			continue
		}
		if !found {
			target.Start = a.Key
			found = true
		}
		target.End = a.Key
	}
	if !found {
		return Target{}, fmt.Errorf("%s: %w", target, ErrNotFound)
	}
	return target, nil
}

// Ayahs returns the ayahs covered by t.
func (b Book) TargetAyahs(t Target) []Ayah {
	start, err := b.Index(t.Start)
	if err != nil {
		return nil
	}
	end, err := b.Index(t.End)
	if err != nil {
		return nil
	}
	return b.Ayahs[start : end+1]
}

// MatchSurahs returns surahs whose number or name contains query, best
// matches first.
func (b Book) MatchSurahs(query string) []Surah {
	q := foldName(query)
	if q == "" {
		return b.Surahs
	}
	prefix, contains := []Surah{}, []Surah{}
	for _, s := range b.Surahs {
		names := []string{foldName(s.NameEnglish), foldName(s.NameMeaning), NormalizeArabic(s.NameArabic), strconv.Itoa(s.Number)}
		switch {
		case anyPrefix(names, q):
			prefix = append(prefix, s)
		case anyContains(names, q):
			contains = append(contains, s)
		}
	}
	return append(prefix, contains...)
}

func anyPrefix(names []string, q string) bool {
	for _, n := range names {
		if strings.HasPrefix(n, q) {
			return true
		}
	}
	return false
}

func anyContains(names []string, q string) bool {
	for _, n := range names {
		if strings.Contains(n, q) {
			return true
		}
	}
	return false
}

// foldName lowercases, drops the article and keeps letters and digits so
// "Al-Kahf", "al kahf" and "kahf" compare equal.
func foldName(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	for _, article := range []string{"al-", "an-", "ar-", "as-", "ash-", "at-", "az-", "ad-", "adh-", "al ", "the "} {
		n = strings.TrimPrefix(n, article)
	}
	if IsArabic(n) {
		return strings.TrimPrefix(NormalizeArabic(n), "ال")
	}
	var b strings.Builder
	for _, r := range n {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
