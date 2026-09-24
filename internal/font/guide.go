package font

import "strings"

// TerminalGuide returns configuration advice and snippets for popular terminal emulators.
func TerminalGuide() string {
	var b strings.Builder

	b.WriteString("Configure your terminal emulator to use Amiri Quran or Scheherazade New as a fallback font:\n\n")

	b.WriteString("  • Kitty (~/.config/kitty/kitty.conf):\n")
	b.WriteString("      symbol_map U+0600-U+06FF,U+0750-U+077F,U+08A0-U+08FF,U+FB50-U+FDFF,U+FE70-U+FEFF Amiri Quran\n\n")

	b.WriteString("  • WezTerm (~/.wezterm.lua):\n")
	b.WriteString("      config.font = wezterm.font_with_fallback({\n")
	b.WriteString("        \"JetBrains Mono\",\n")
	b.WriteString("        \"Amiri Quran\",\n")
	b.WriteString("        \"Scheherazade New\",\n")
	b.WriteString("      })\n\n")

	b.WriteString("  • Alacritty (~/.config/alacritty/alacritty.toml):\n")
	b.WriteString("      [font.normal]\n")
	b.WriteString("      family = \"Amiri Quran\"\n")
	b.WriteString("      # Or keep your primary font and let system fontconfig handle Arabic fallback\n\n")

	b.WriteString("  • Ghostty (~/.config/ghostty/config):\n")
	b.WriteString("      font-family = \"JetBrains Mono\"\n")
	b.WriteString("      font-family = \"Amiri Quran\"\n\n")

	b.WriteString("  • iTerm2 / Terminal.app / GNOME Terminal:\n")
	b.WriteString("      Set font or fallback font to \"Amiri Quran\" or \"Scheherazade New\" in profile settings.\n")

	return b.String()
}
