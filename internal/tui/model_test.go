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

func TestMouseWheel(t *testing.T) {
	m := newTestModel(t)
	// Open Surah 2 so we have many ayahs
	m.openSurah(2)
	if m.cursor != 0 {
		t.Fatalf("expected cursor 0, got %d", m.cursor)
	}

	// Mouse wheel down over reader (X = 50, Y = 10)
	m.Update(tea.MouseWheelMsg{
		X:      50,
		Y:      10,
		Button: tea.MouseWheelDown,
	})
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after wheel down, got %d", m.cursor)
	}

	// Mouse wheel up over reader
	m.Update(tea.MouseWheelMsg{
		X:      50,
		Y:      10,
		Button: tea.MouseWheelUp,
	})
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after wheel up, got %d", m.cursor)
	}

	// Mouse wheel down over sidebar (X = 10, Y = 10)
	initialSideOffset := m.sideOffset
	m.Update(tea.MouseWheelMsg{
		X:      10,
		Y:      10,
		Button: tea.MouseWheelDown,
	})
	if m.sideOffset != initialSideOffset+3 {
		t.Errorf("expected sidebar offset to increase by 3, got %d from %d", m.sideOffset, initialSideOffset)
	}
}

func TestMouseClick(t *testing.T) {
	m := newTestModel(t)
	// Click in sidebar at row 3 (header is 2 rows, so row 3 is first or second surah)
	m.Update(tea.MouseClickMsg{
		X:      5,
		Y:      3,
		Button: tea.MouseLeft,
	})
	if m.focus != focusSidebar {
		t.Errorf("expected sidebar to be focused after click")
	}

	// Click in reader area at X = 50, Y = 5
	m.Update(tea.MouseClickMsg{
		X:      50,
		Y:      5,
		Button: tea.MouseLeft,
	})
	if m.focus != focusReader {
		t.Errorf("expected reader to be focused after click")
	}
}

func TestSettingsOverlay(t *testing.T) {
	m := newTestModel(t)
	typeText(m, "S")
	if m.overlay == nil || m.overlay.kind != pickSettings {
		t.Fatalf("expected settings overlay to be open")
	}

	items := m.settingsItems("")
	foundFont := false
	foundArabic := false
	foundVolume := false
	for _, it := range items {
		if strings.Contains(it.title, "Quran Font") {
			foundFont = true
		}
		if strings.Contains(it.title, "Arabic Mode") {
			foundArabic = true
		}
		if strings.Contains(it.title, "Volume") {
			foundVolume = true
		}
	}
	if !foundFont {
		t.Errorf("expected settings to contain Quran Font")
	}
	if !foundArabic {
		t.Errorf("expected settings to contain Arabic Mode")
	}
	if !foundVolume {
		t.Errorf("expected settings to contain Volume")
	}

	// First item is Quran Font; press Enter to open pickFont
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.overlay == nil || m.overlay.kind != pickFont {
		t.Fatalf("expected font picker to open on enter, got %+v", m.overlay)
	}

	// Select Scheherazade New
	fontList := m.fontItems("")
	if len(fontList) < 5 {
		t.Fatalf("expected at least 5 font items, got %d", len(fontList))
	}
	m.overlay.cursor = 1
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.fontName != "Scheherazade New" {
		t.Errorf("expected fontName to be Scheherazade New, got %q", m.fontName)
	}

	// Open settings again and press Esc to dismiss
	typeText(m, "S")
	m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.overlay != nil {
		t.Errorf("expected overlay to be closed after Esc")
	}
}
