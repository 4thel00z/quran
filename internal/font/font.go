package font

// Font represents a font asset available for download and installation.
type Font struct {
	Name        string
	Description string
	FileName    string
	URL         string
	Size        int64
}

// RecommendedFonts returns the curated bundle of Quranic fonts.
func RecommendedFonts() []Font {
	return []Font{
		{
			Name:        "Amiri Quran",
			Description: "Classical Naskh typeface optimized for Quranic text by Khaled Hosny",
			FileName:    "AmiriQuran-Regular.ttf",
			URL:         "https://raw.githubusercontent.com/google/fonts/main/ofl/amiriquran/AmiriQuran-Regular.ttf",
			Size:        136920,
		},
		{
			Name:        "Scheherazade New",
			Description: "Traditional Arabic Naskh typeface designed by SIL International",
			FileName:    "ScheherazadeNew-Regular.ttf",
			URL:         "https://raw.githubusercontent.com/google/fonts/main/ofl/scheherazadenew/ScheherazadeNew-Regular.ttf",
			Size:        331504,
		},
	}
}
