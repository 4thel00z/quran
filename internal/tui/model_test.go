package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/4thel00z/quran/internal/assets"
	"github.com/4thel00z/quran/internal/audio"
	"github.com/4thel00z/quran/internal/quran"
	"github.com/4thel00z/quran/internal/render"
)

func newTestModel(t *testing.T) *Model {
	t.Helper()
	book, err := assets.Book()
	if err != nil {
		t.Fatal(err)
	}
	reciter, err := quran.FindReciter("husary")
	if err != nil {
		t.Fatal(err)
	}
	translation, err := quran.FindTranslation("en-sahih")
	if err != nil {
		t.Fatal(err)
	}
	m, err := New(Options{
		Book:        book,
		Reciter:     reciter,
		Translation: translation,
		Mode:        render.Visual,
		Fetcher:     audio.NewFetcher("http://127.0.0.1:1", ""),
		Player:      audio.NewPlayer(),
	})
	if err != nil {
		t.Fatal(err)
	}
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return m
}

func typeText(m *Model, text string) {
	for _, r := range text {
		m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

func TestSearchJumpsToAyah(t *testing.T) {
	m := newTestModel(t)
	typeText(m, "/")
	if m.overlay == nil {
		t.Fatal("search did not open")
	}
	typeText(m, "18:10")
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.overlay != nil || m.cursorKey() != (quran.Key{Surah: 18, Ayah: 10}) {
		t.Fatalf("cursor at %s", m.cursorKey())
	}
	if view := m.View().Content; !strings.Contains(view, "18:10") || !strings.Contains(view, "Al-Kahf") {
		t.Fatal("view does not show 18:10 in Al-Kahf")
	}
}

func TestJuzJumpBoundsPlayback(t *testing.T) {
	m := newTestModel(t)
	typeText(m, "/juz 30")
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.cursorKey() != (quran.Key{Surah: 78, Ayah: 1}) || m.rangeEnd == nil || *m.rangeEnd != (quran.Key{Surah: 114, Ayah: 6}) {
		t.Fatalf("cursor %s range end %v", m.cursorKey(), m.rangeEnd)
	}
	if next, ok := m.nextKey(quran.Key{Surah: 114, Ayah: 6}); ok {
		t.Fatalf("playback continues past juz 30 to %s", next)
	}
	m.repeat = repeatRange
	if next, _ := m.nextKey(quran.Key{Surah: 114, Ayah: 6}); next != (quran.Key{Surah: 78, Ayah: 1}) {
		t.Fatalf("range repeat goes to %s", next)
	}
}

func TestSurahNavigation(t *testing.T) {
	m := newTestModel(t)
	typeText(m, "l")
	typeText(m, "G")
	if m.cursorKey() != (quran.Key{Surah: 2, Ayah: 286}) {
		t.Fatalf("cursor at %s", m.cursorKey())
	}
	typeText(m, "j")
	if m.cursorKey() != (quran.Key{Surah: 2, Ayah: 286}) {
		t.Fatalf("cursor moved past the last ayah to %s", m.cursorKey())
	}
}
