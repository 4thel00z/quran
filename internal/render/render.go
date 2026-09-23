// Package render lays out Arabic and translation text for the terminal.
package render

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/4thel00z/quran/internal/arabic"
)

// Mode selects how right-to-left text reaches the terminal.
type Mode string

const (
	// Visual shapes letters and reorders each line; works in any terminal.
	Visual Mode = "visual"
	// Native writes logical order for terminals with bidi support.
	Native Mode = "native"
	Auto   Mode = "auto"
)

func ParseMode(text string) (Mode, error) {
	mode := Mode(strings.ToLower(text))
	if !slices.Contains([]Mode{Visual, Native, Auto}, mode) {
		return "", fmt.Errorf("arabic mode %q: want visual, native or auto", text)
	}
	if mode != Auto {
		return mode, nil
	}
	return detectMode(), nil
}

// detectMode picks Native for terminals known to implement bidi.
func detectMode() Mode {
	if os.Getenv("TERM_PROGRAM") == "Apple_Terminal" || os.Getenv("KONSOLE_VERSION") != "" || os.Getenv("VTE_VERSION") != "" {
		return Native
	}
	return Visual
}

func (m Mode) Toggle() Mode {
	if m == Visual {
		return Native
	}
	return Visual
}

// Word prepares one right-to-left word for display.
func (m Mode) Word(word string) string {
	if m == Native {
		return word
	}
	return arabic.Visual(word)
}

// AyahMarker is the end-of-ayah ornament with Arabic-Indic digits.
func AyahMarker(number int) string {
	return "﴿" + arabicDigits(number) + "﴾"
}

func arabicDigits(number int) string {
	digits := []rune(fmt.Sprint(number))
	for i, d := range digits {
		digits[i] = d - '0' + '٠'
	}
	return string(digits)
}

// RTLLines wraps words right to left into lines of at most width cells,
// each right aligned. style renders word i; pad renders the filler.
func RTLLines(words []string, width int, mode Mode, style func(i int) lipgloss.Style, pad lipgloss.Style) []string {
	type cell struct {
		index int
		text  string
		width int
	}
	lines := [][]cell{}
	current := []cell{}
	used := 0
	for i, word := range words {
		shown := mode.Word(word)
		w := ansi.StringWidth(shown)
		if len(current) > 0 && used+1+w > width {
			lines = append(lines, current)
			current, used = nil, 0
		}
		if len(current) > 0 {
			used++
		}
		current = append(current, cell{index: i, text: shown, width: w})
		used += w
	}
	if len(current) > 0 {
		lines = append(lines, current)
	}
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if mode == Visual {
			slices.Reverse(line)
		}
		parts := make([]string, 0, len(line))
		lineWidth := len(line) - 1
		for _, c := range line {
			parts = append(parts, style(c.index).Render(c.text))
			lineWidth += c.width
		}
		filler := pad.Render(strings.Repeat(" ", max(0, width-lineWidth)))
		result = append(result, filler+strings.Join(parts, pad.Render(" ")))
	}
	return result
}

// TextLines wraps a translation; right-to-left scripts are right aligned.
func TextLines(text string, width int, mode Mode, style lipgloss.Style) []string {
	if !arabic.IsRTL(text) {
		wrapped := ansi.Wordwrap(text, width, "")
		lines := strings.Split(wrapped, "\n")
		for i, line := range lines {
			lines[i] = style.Render(line)
		}
		return lines
	}
	words := strings.Fields(text)
	return RTLLines(words, width, mode, func(int) lipgloss.Style { return style }, lipgloss.NewStyle())
}
