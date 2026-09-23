package quran

import "fmt"

type Reciter struct {
	Slug       string
	Name       string
	Style      string
	QuranComID int
	// Source is where the mirror copies the ayah files from.
	Source string
}

var Reciters = []Reciter{
	{Slug: "minshawi-murattal", Name: "Mohamed Siddiq al-Minshawi", Style: "Murattal", QuranComID: 9, Source: "https://verses.quran.com/Minshawi/Murattal/mp3/"},
	{Slug: "minshawi-mujawwad", Name: "Mohamed Siddiq al-Minshawi", Style: "Mujawwad", QuranComID: 8, Source: "https://verses.quran.com/Minshawi/Mujawwad/mp3/"},
	{Slug: "husary", Name: "Mahmoud Khalil al-Husary", Style: "Murattal", QuranComID: 6, Source: "https://mirrors.quranicaudio.com/everyayah/Husary_64kbps/"},
	{Slug: "husary-muallim", Name: "Mahmoud Khalil al-Husary", Style: "Muallim", QuranComID: 12, Source: "https://mirrors.quranicaudio.com/everyayah/Husary_Muallim_128kbps/"},
	{Slug: "alafasy", Name: "Mishari Rashid al-Afasy", Style: "Murattal", QuranComID: 7, Source: "https://verses.quran.com/Alafasy/mp3/"},
	{Slug: "abdulbaset-murattal", Name: "AbdulBaset AbdulSamad", Style: "Murattal", QuranComID: 2, Source: "https://verses.quran.com/AbdulBaset/Murattal/mp3/"},
	{Slug: "abdulbaset-mujawwad", Name: "AbdulBaset AbdulSamad", Style: "Mujawwad", QuranComID: 1, Source: "https://verses.quran.com/AbdulBaset/Mujawwad/mp3/"},
	{Slug: "sudais", Name: "Abdur-Rahman as-Sudais", Style: "Murattal", QuranComID: 3, Source: "https://verses.quran.com/Sudais/mp3/"},
	{Slug: "shatri", Name: "Abu Bakr al-Shatri", Style: "Murattal", QuranComID: 4, Source: "https://verses.quran.com/Shatri/mp3/"},
	{Slug: "rifai", Name: "Hani ar-Rifai", Style: "Murattal", QuranComID: 5, Source: "https://verses.quran.com/Rifai/mp3/"},
	{Slug: "shuraym", Name: "Saud ash-Shuraym", Style: "Murattal", QuranComID: 10, Source: "https://verses.quran.com/Shuraym/mp3/"},
	{Slug: "tablawi", Name: "Mohamed al-Tablawi", Style: "Murattal", QuranComID: 11, Source: "https://mirrors.quranicaudio.com/everyayah/Mohammad_al_Tablaway_128kbps/"},
}

func (r Reciter) Title() string {
	return r.Name + " · " + r.Style
}

func FindReciter(slug string) (Reciter, error) {
	for _, r := range Reciters {
		if r.Slug == slug {
			return r, nil
		}
	}
	return Reciter{}, fmt.Errorf("reciter %q: %w", slug, ErrNotFound)
}

// AudioFile is the file name of one ayah, e.g. 002255.mp3.
func AudioFile(key Key) string {
	return fmt.Sprintf("%03d%03d.mp3", key.Surah, key.Ayah)
}

// HostedPath is the ayah's path under the audio host's web root.
func (r Reciter) HostedPath(key Key) string {
	return "audio/" + r.Slug + "/" + AudioFile(key)
}

func (r Reciter) AudioURL(base string, key Key) string {
	return base + "/" + r.HostedPath(key)
}

func (r Reciter) SourceURL(key Key) string {
	return r.Source + AudioFile(key)
}
