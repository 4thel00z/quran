package quran

import (
	"strings"
	"unicode"
)

type Hit struct {
	Ayah
	// InArabic is set when the match is in the Arabic text rather than the
	// translation.
	InArabic bool
}

// Search finds ayahs whose Arabic text (ignoring diacritics and letter
// variants) or translation contains every word of query.
func (b Book) Search(translation [][]string, query string, limit int) []Hit {
	terms := strings.Fields(query)
	if len(terms) == 0 {
		return nil
	}
	arabic := IsArabic(query)
	for i, term := range terms {
		terms[i] = foldTerm(term, arabic)
	}
	hits := []Hit{}
	for _, a := range b.Ayahs {
		if len(hits) >= limit {
			break
		}
		text := foldTerm(a.Clean, true)
		if !arabic {
			text = foldTerm(translation[a.Surah-1][a.Ayah-1], false)
		}
		if !containsAll(text, terms) {
			continue
		}
		hits = append(hits, Hit{Ayah: a, InArabic: arabic})
	}
	return hits
}

func foldTerm(text string, arabic bool) string {
	if arabic {
		return NormalizeArabic(text)
	}
	return strings.ToLower(text)
}

func containsAll(text string, terms []string) bool {
	for _, t := range terms {
		if !strings.Contains(text, t) {
			return false
		}
	}
	return true
}

var arabicVariants = strings.NewReplacer(
	"أ", "ا", "إ", "ا", "آ", "ا", "ٱ", "ا",
	"ى", "ي", "ئ", "ي", "ؤ", "و", "ة", "ه", "ـ", "",
)

// NormalizeArabic removes diacritics and folds letter variants.
func NormalizeArabic(text string) string {
	var b strings.Builder
	for _, r := range text {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return arabicVariants.Replace(b.String())
}

func IsArabic(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Arabic, r) {
			return true
		}
	}
	return false
}
