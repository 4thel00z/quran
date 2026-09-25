package font

import "strings"

// TerminalGuide returns configuration advice and snippets for popular terminal emulators.
func TerminalGuide() string {
	var b strings.Builder

	b.WriteString("Configure your terminal emulator to use your preferred Quranic font as a fallback:\n")
	b.WriteString("  Available styles: Amiri Quran (Madani), Scheherazade New (Naskh),\n")
	b.WriteString("                    Al Qalam Quran Majeed (Indo-Pak), KFGQPC Nastaleeq (Nastaleeq),\n")
	b.WriteString("                    Shaikh Hamdullah Mushaf (Turkish/Ottoman).\n\n")

	b.WriteString("  • Kitty (~/.config/kitty/kitty.conf):\n")
	b.WriteString("      symbol_map U+0600-U+06FF,U+0750-U+077F,U+08A0-U+08FF,U+FB50-U+FDFF,U+FE70-U+FEFF Amiri Quran\n")
	b.WriteString("      # Or replace 'Amiri Quran' with 'Al Qalam Quran Majeed' or 'Shaikh Hamdullah Mushaf'\n\n")

	b.WriteString("  • WezTerm (~/.wezterm.lua):\n")
	b.WriteString("      config.font = wezterm.font_with_fallback({\n")
	b.WriteString("        \"JetBrains Mono\",\n")
	b.WriteString("        \"Amiri Quran\",\n")
	b.WriteString("        \"Al Qalam Quran Majeed\",\n")
	b.WriteString("        \"Shaikh Hamdullah Mushaf\",\n")
	b.WriteString("      })\n\n")

	b.WriteString("  • Alacritty (~/.config/alacritty/alacritty.toml):\n")
	b.WriteString("      [font.normal]\n")
	b.WriteString("      family = \"Amiri Quran\" # or \"Al Qalam Quran Majeed\" / \"Shaikh Hamdullah Mushaf\"\n\n")

	b.WriteString("  • Ghostty (~/.config/ghostty/config):\n")
	b.WriteString("      font-family = \"JetBrains Mono\"\n")
	b.WriteString("      font-family = \"Amiri Quran\"\n")
	b.WriteString("      font-family = \"Al Qalam Quran Majeed\"\n\n")

	b.WriteString("  • iTerm2 / Terminal.app / GNOME Terminal:\n")
	b.WriteString("      Set font or fallback font to your preferred Quran font in profile settings.\n")

	return b.String()
}
