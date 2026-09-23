package assets

import (
	"testing"

	"github.com/4thel00z/quran/internal/quran"
)

func TestBookIsComplete(t *testing.T) {
	book, err := Book()
	if err != nil {
		t.Fatal(err)
	}
	if len(book.Surahs) != quran.SurahCount || len(book.Ayahs) != quran.AyahCount {
		t.Fatalf("got %d surahs, %d ayahs", len(book.Surahs), len(book.Ayahs))
	}
	last := book.Ayahs[len(book.Ayahs)-1]
	if last.Juz != quran.JuzCount || last.Hizb != quran.HizbCount || last.Page != quran.PageCount {
		t.Fatalf("last ayah juz=%d hizb=%d page=%d", last.Juz, last.Hizb, last.Page)
	}
	for _, a := range book.Ayahs {
		if a.Juz == 0 || a.Hizb == 0 || a.Page == 0 || a.Clean == "" {
			t.Fatalf("%s missing metadata: %+v", a.Key, a)
		}
	}
}

func TestTranslationsCoverEveryAyah(t *testing.T) {
	book, err := Book()
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range quran.Translations {
		verses, err := Translation(tr.ID)
		if err != nil {
			t.Fatal(err)
		}
		for i, s := range book.Surahs {
			if len(verses[i]) != s.AyahCount {
				t.Fatalf("%s surah %d: %d verses, want %d", tr.ID, s.Number, len(verses[i]), s.AyahCount)
			}
		}
	}
}

// Word timings index into Ayah.Words; a mismatch breaks word highlighting.
func TestSegmentsAlignWithWords(t *testing.T) {
	book, err := Book()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range quran.Reciters {
		segments, err := Segments(r.Slug)
		if err != nil {
			t.Fatal(err)
		}
		missing, overflow := 0, 0
		for _, a := range book.Ayahs {
			timings := segments.For(a.Key)
			if len(timings) == 0 {
				missing++
				continue
			}
			if timings[len(timings)-1][1] > len(a.Words) {
				overflow++
			}
		}
		if missing > 10 || overflow > 10 {
			t.Errorf("%s: %d ayahs without timings, %d overflowing", r.Slug, missing, overflow)
		}
		t.Logf("%s: %d missing, %d overflowing", r.Slug, missing, overflow)
	}
}
