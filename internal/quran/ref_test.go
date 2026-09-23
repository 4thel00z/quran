package quran_test

import (
	"testing"

	"github.com/4thel00z/quran/internal/assets"
	"github.com/4thel00z/quran/internal/quran"
)

func TestResolve(t *testing.T) {
	book, err := assets.Book()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		query string
		kind  quran.TargetKind
		start quran.Key
		end   quran.Key
	}{
		{query: "18", kind: quran.TargetSurah, start: quran.Key{Surah: 18, Ayah: 1}, end: quran.Key{Surah: 18, Ayah: 110}},
		{query: "2:255", kind: quran.TargetAyah, start: quran.Key{Surah: 2, Ayah: 255}, end: quran.Key{Surah: 2, Ayah: 255}},
		{query: "2.255 - 257", kind: quran.TargetRange, start: quran.Key{Surah: 2, Ayah: 255}, end: quran.Key{Surah: 2, Ayah: 257}},
		{query: "juz 30", kind: quran.TargetJuz, start: quran.Key{Surah: 78, Ayah: 1}, end: quran.Key{Surah: 114, Ayah: 6}},
		{query: "j2", kind: quran.TargetJuz, start: quran.Key{Surah: 2, Ayah: 142}, end: quran.Key{Surah: 2, Ayah: 252}},
		{query: "hizb 60", kind: quran.TargetHizb, start: quran.Key{Surah: 87, Ayah: 1}, end: quran.Key{Surah: 114, Ayah: 6}},
		{query: "page 1", kind: quran.TargetPage, start: quran.Key{Surah: 1, Ayah: 1}, end: quran.Key{Surah: 1, Ayah: 7}},
		{query: "p604", kind: quran.TargetPage, start: quran.Key{Surah: 112, Ayah: 1}, end: quran.Key{Surah: 114, Ayah: 6}},
		{query: "Al-Kahf", kind: quran.TargetSurah, start: quran.Key{Surah: 18, Ayah: 1}, end: quran.Key{Surah: 18, Ayah: 110}},
		{query: "yasin", kind: quran.TargetSurah, start: quran.Key{Surah: 36, Ayah: 1}, end: quran.Key{Surah: 36, Ayah: 83}},
		{query: "الكهف", kind: quran.TargetSurah, start: quran.Key{Surah: 18, Ayah: 1}, end: quran.Key{Surah: 18, Ayah: 110}},
	}
	for _, c := range cases {
		t.Run(c.query, func(t *testing.T) {
			got, err := book.Resolve(c.query)
			if err != nil {
				t.Fatal(err)
			}
			if got.Kind != c.kind || got.Start != c.start || got.End != c.end {
				t.Fatalf("got %+v", got)
			}
		})
	}
	for _, bad := range []string{"115", "2:287", "juz 31", "2:5-3", "zzzz"} {
		if _, err := book.Resolve(bad); err == nil {
			t.Errorf("Resolve(%q) succeeded", bad)
		}
	}
}

func TestSearch(t *testing.T) {
	book, err := assets.Book()
	if err != nil {
		t.Fatal(err)
	}
	translation, err := assets.Translation("en-sahih")
	if err != nil {
		t.Fatal(err)
	}
	hits := book.Search(translation, "ever-living sustainer kursi", 50)
	if !containsKey(hits, quran.Key{Surah: 2, Ayah: 255}) {
		t.Fatalf("2:255 missing from %d hits", len(hits))
	}
	arabic := book.Search(translation, "الله لا اله الا هو الحي القيوم", 50)
	if !containsKey(arabic, quran.Key{Surah: 2, Ayah: 255}) {
		t.Fatalf("2:255 missing from %d arabic hits", len(arabic))
	}
}

func containsKey(hits []quran.Hit, key quran.Key) bool {
	for _, h := range hits {
		if h.Key == key {
			return true
		}
	}
	return false
}
