package tui

import (
	"fmt"
	"strings"

	"github.com/4thel00z/quran/internal/quran"
	"github.com/4thel00z/quran/internal/render"
)

const searchLimit = 200

func (m *Model) searchItems(query string) []pickerItem {
	q := strings.TrimSpace(query)
	if q == "" {
		return m.surahItems(m.book.Surahs)
	}
	items := []pickerItem{}
	if target, err := m.book.Resolve(q); err == nil {
		items = append(items, m.targetItem(target))
	}
	for _, surah := range m.book.MatchSurahs(q) {
		if len(items) >= 8 {
			break
		}
		if len(items) > 0 && items[0].value.(quran.Target).Kind == quran.TargetSurah && items[0].value.(quran.Target).Number == surah.Number {
			continue
		}
		items = append(items, m.surahItems([]quran.Surah{surah})...)
	}
	if len([]rune(q)) < 3 {
		return items
	}
	for _, hit := range m.book.Search(m.verses, q, searchLimit) {
		items = append(items, m.hitItem(hit))
	}
	return items
}

func (m *Model) targetItem(target quran.Target) pickerItem {
	detail := ""
	start := m.book.Surah(target.Start.Surah)
	switch target.Kind {
	case quran.TargetSurah:
		detail = fmt.Sprintf("%s · %d ayahs", start.NameMeaning, start.AyahCount)
	case quran.TargetAyah:
		detail = start.NameEnglish
	default:
		end := m.book.Surah(target.End.Surah)
		detail = fmt.Sprintf("%s %s → %s %s", start.NameEnglish, target.Start, end.NameEnglish, target.End)
	}
	title := "→ " + target.String()
	if target.Kind == quran.TargetSurah {
		title = fmt.Sprintf("→ %d. %s", start.Number, start.NameEnglish)
	}
	return pickerItem{title: title, detail: detail, value: target}
}

func (m *Model) surahItems(surahs []quran.Surah) []pickerItem {
	items := make([]pickerItem, 0, len(surahs))
	for _, s := range surahs {
		target, err := m.book.Resolve(fmt.Sprint(s.Number))
		if err != nil {
			continue
		}
		items = append(items, pickerItem{
			title:  fmt.Sprintf("%3d. %s", s.Number, s.NameEnglish),
			detail: fmt.Sprintf("%s · %d · %s", s.NameMeaning, s.AyahCount, m.mode.Word(s.NameArabic)),
			value:  target,
		})
	}
	return items
}

func (m *Model) hitItem(hit quran.Hit) pickerItem {
	text := m.verses[hit.Key.Surah-1][hit.Key.Ayah-1]
	if hit.InArabic {
		text = hit.Clean
	}
	return pickerItem{
		title:  fmt.Sprintf("%-7s %s", hit.Key, m.book.Surah(hit.Surah).NameEnglish),
		detail: m.snippet(text, hit.InArabic),
		value:  quran.Target{Kind: quran.TargetAyah, Number: hit.Surah, Start: hit.Key, End: hit.Key},
	}
}

func (m *Model) snippet(text string, rtl bool) string {
	if !rtl {
		return text
	}
	words := strings.Fields(text)
	words = words[:min(len(words), 8)]
	for i, w := range words {
		words[i] = m.mode.Word(w)
	}
	if m.mode == render.Visual {
		for i, j := 0, len(words)-1; i < j; i, j = i+1, j-1 {
			words[i], words[j] = words[j], words[i]
		}
	}
	return strings.Join(words, " ")
}

func (m *Model) reciterItems(query string) []pickerItem {
	q := strings.ToLower(query)
	items := []pickerItem{}
	for _, r := range quran.Reciters {
		if !strings.Contains(strings.ToLower(r.Title()+" "+r.Slug), q) {
			continue
		}
		title := r.Name
		if r.Slug == m.reciter.Slug {
			title = "● " + title
		}
		items = append(items, pickerItem{title: title, detail: r.Style + " · " + r.Slug, value: r.Slug})
	}
	return items
}

func (m *Model) translationItems(query string) []pickerItem {
	q := strings.ToLower(query)
	items := []pickerItem{}
	for _, t := range quran.Translations {
		if !strings.Contains(strings.ToLower(t.Title()+" "+t.ID), q) {
			continue
		}
		title := t.Language
		if t.ID == m.translation.ID {
			title = "● " + title
		}
		items = append(items, pickerItem{title: title, detail: t.Name + " · " + t.ID, value: t.ID})
	}
	return items
}
