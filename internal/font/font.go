package font

// Font represents a font asset available for download and installation.
type Font struct {
	Name        string
	Style       string
	Description string
	FileName    string
	URL         string
	Size        int64
}

// RecommendedFonts returns the curated bundle of Quranic fonts across scripts.
func RecommendedFonts() []Font {
	return []Font{
		{
			Name:        "Amiri Quran",
			Style:       "Madani Naskh",
			Description: "Classical Naskh typeface optimized for Quranic text by Khaled Hosny",
			FileName:    "AmiriQuran-Regular.ttf",
			URL:         "https://raw.githubusercontent.com/google/fonts/main/ofl/amiriquran/AmiriQuran-Regular.ttf",
			Size:        136920,
		},
		{
			Name:        "Scheherazade New",
			Style:       "Extended Naskh",
			Description: "Traditional Arabic Naskh typeface designed by SIL International",
			FileName:    "ScheherazadeNew-Regular.ttf",
			URL:         "https://raw.githubusercontent.com/google/fonts/main/ofl/scheherazadenew/ScheherazadeNew-Regular.ttf",
			Size:        331504,
		},
		{
			Name:        "Al Qalam Quran Majeed",
			Style:       "Indo-Pak Naskh",
			Description: "Standard 15/16-line Indo-Pak Quran script for South Asia",
			FileName:    "AlQalamQuranMajeed.ttf",
			URL:         "https://cdn.jsdelivr.net/gh/fawazahmed0/quran-api@1/fonts/al-qalam-quran-majeed.ttf",
			Size:        180260,
		},
		{
			Name:        "KFGQPC Nastaleeq",
			Style:       "Indo-Pak Nastaleeq",
			Description: "Official King Fahd Complex Nastaleeq font for Urdu and South Asian scripts",
			FileName:    "HafsNastaleeq-Regular.ttf",
			URL:         "https://cdn.jsdelivr.net/gh/fawazahmed0/quran-api@1/fonts/hafs-nastaleeq-ver10.ttf",
			Size:        252820,
		},
		{
			Name:        "Shaikh Hamdullah",
			Style:       "Ottoman / Turkish",
			Description: "Classic Ottoman calligraphic Quran font from Turkish Diyanet",
			FileName:    "KuranKerimHamdullah.ttf",
			URL:         "https://webdosya.diyanet.gov.tr/kuran/kuranikerim/dosyalar/font/KuranKerimFontHamdullah.ttf",
			Size:        78728,
		},
	}
}
