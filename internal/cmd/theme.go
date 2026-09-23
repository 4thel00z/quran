package cmd

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/fang"
)

// colorScheme gives the help and error output the reader's palette.
func colorScheme(c lipgloss.LightDarkFunc) fang.ColorScheme {
	gold := c(lipgloss.Color("#9A6B12"), lipgloss.Color("#E7BE62"))
	green := c(lipgloss.Color("#1F6E52"), lipgloss.Color("#4FB08A"))
	muted := c(lipgloss.Color("#6B6B6B"), lipgloss.Color("#8A8F8C"))
	text := c(lipgloss.Color("#2A2A2A"), lipgloss.Color("#E6E1D6"))
	return fang.ColorScheme{
		Base:           text,
		Title:          gold,
		Description:    muted,
		Codeblock:      c(lipgloss.Color("#EEF3F0"), lipgloss.Color("#18221E")),
		Program:        gold,
		DimmedArgument: muted,
		Comment:        muted,
		Flag:           green,
		FlagDefault:    muted,
		Command:        green,
		QuotedString:   gold,
		Argument:       text,
		Help:           muted,
		Dash:           muted,
		ErrorHeader:    [2]color.Color{c(lipgloss.Color("#FFFFFF"), lipgloss.Color("#1A1712")), c(lipgloss.Color("#B3261E"), lipgloss.Color("#F2857A"))},
		ErrorDetails:   c(lipgloss.Color("#B3261E"), lipgloss.Color("#F2857A")),
	}
}
