package quran

import "fmt"

type Translation struct {
	ID       string
	Language string
	Name     string
	// RTL marks scripts written right to left.
	RTL bool
}

var Translations = []Translation{
	{ID: "en-sahih", Language: "English", Name: "Saheeh International"},
	{ID: "en-pickthall", Language: "English", Name: "Pickthall"},
	{ID: "en-yusufali", Language: "English", Name: "Yusuf Ali"},
	{ID: "en-hilali", Language: "English", Name: "Hilali & Khan"},
	{ID: "en-asad", Language: "English", Name: "Muhammad Asad"},
	{ID: "en-itani", Language: "English", Name: "Talal Itani"},
	{ID: "en-transliteration", Language: "Transliteration", Name: "Transliteration"},
	{ID: "de-bubenheim", Language: "Deutsch", Name: "Bubenheim & Elyas"},
	{ID: "de-aburida", Language: "Deutsch", Name: "Abu Rida"},
	{ID: "de-zaidan", Language: "Deutsch", Name: "Zaidan"},
	{ID: "fr-hamidullah", Language: "Français", Name: "Hamidullah"},
	{ID: "es-cortes", Language: "Español", Name: "Cortés"},
	{ID: "it-piccardo", Language: "Italiano", Name: "Piccardo"},
	{ID: "nl-siregar", Language: "Nederlands", Name: "Siregar"},
	{ID: "pt-elhayek", Language: "Português", Name: "El-Hayek"},
	{ID: "sv-bernstrom", Language: "Svenska", Name: "Bernström"},
	{ID: "tr-diyanet", Language: "Türkçe", Name: "Diyanet İşleri"},
	{ID: "ru-kuliev", Language: "Русский", Name: "Kuliev"},
	{ID: "bs-korkut", Language: "Bosanski", Name: "Korkut"},
	{ID: "sq-ahmeti", Language: "Shqip", Name: "Sherif Ahmeti"},
	{ID: "id-indonesian", Language: "Indonesia", Name: "Kementerian Agama"},
	{ID: "ms-basmeih", Language: "Melayu", Name: "Basmeih"},
	{ID: "ber-mensur", Language: "Tamazight", Name: "Mensur"},
	{ID: "so-abduh", Language: "Soomaali", Name: "Abduh"},
	{ID: "sw-barwani", Language: "Kiswahili", Name: "Al-Barwani"},
	{ID: "ha-gumi", Language: "Hausa", Name: "Gumi"},
	{ID: "bn-bengali", Language: "বাংলা", Name: "Muhiuddin Khan"},
	{ID: "hi-farooq", Language: "हिन्दी", Name: "Farooq Khan & Nadvi"},
	{ID: "zh-jian", Language: "中文", Name: "Ma Jian"},
	{ID: "ja-japanese", Language: "日本語", Name: "Japanese"},
	{ID: "ko-korean", Language: "한국어", Name: "Korean"},
	{ID: "ur-jalandhry", Language: "اردو", Name: "Jalandhry", RTL: true},
	{ID: "fa-makarem", Language: "فارسی", Name: "Makarem Shirazi", RTL: true},
	{ID: "ar-muyassar", Language: "العربية", Name: "التفسير الميسر", RTL: true},
	{ID: "ar-jalalayn", Language: "العربية", Name: "تفسير الجلالين", RTL: true},
}

func (t Translation) Title() string {
	return t.Language + " · " + t.Name
}

func FindTranslation(id string) (Translation, error) {
	for _, t := range Translations {
		if t.ID == id {
			return t, nil
		}
	}
	return Translation{}, fmt.Errorf("translation %q: %w", id, ErrNotFound)
}
