package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type palette struct {
	text      color.Color
	arabic    color.Color
	muted     color.Color
	faint     color.Color
	accent    color.Color
	gold      color.Color
	onGold    color.Color
	danger    color.Color
	selection color.Color
}

func newPalette(dark bool) palette {
	pick := lipgloss.LightDark(dark)
	return palette{
		text:      pick(lipgloss.Color("#2A2A2A"), lipgloss.Color("#E6E1D6")),
		arabic:    pick(lipgloss.Color("#1C1C1C"), lipgloss.Color("#F4EEDF")),
		muted:     pick(lipgloss.Color("#6B6B6B"), lipgloss.Color("#9A9A94")),
		faint:     pick(lipgloss.Color("#B8B8B0"), lipgloss.Color("#4A4F4C")),
		accent:    pick(lipgloss.Color("#1F6E52"), lipgloss.Color("#4FB08A")),
		gold:      pick(lipgloss.Color("#9A6B12"), lipgloss.Color("#E7BE62")),
		onGold:    pick(lipgloss.Color("#FFFFFF"), lipgloss.Color("#1A1712")),
		danger:    pick(lipgloss.Color("#B3261E"), lipgloss.Color("#F2857A")),
		selection: pick(lipgloss.Color("#E4EFE9"), lipgloss.Color("#1D2B25")),
	}
}

type styles struct {
	palette
	title        lipgloss.Style
	subtitle     lipgloss.Style
	muted        lipgloss.Style
	faint        lipgloss.Style
	arabic       lipgloss.Style
	recited      lipgloss.Style
	marker       lipgloss.Style
	translation  lipgloss.Style
	key          lipgloss.Style
	keyPlaying   lipgloss.Style
	gutter       lipgloss.Style
	gutterCursor lipgloss.Style
	gutterPlay   lipgloss.Style
	banner       lipgloss.Style
	bannerArabic lipgloss.Style
	bismillah    lipgloss.Style
	sidebar      lipgloss.Style
	sideItem     lipgloss.Style
	sideCurrent  lipgloss.Style
	sideCursor   lipgloss.Style
	sideArabic   lipgloss.Style
	rule         lipgloss.Style
	status       lipgloss.Style
	statusError  lipgloss.Style
	pill         lipgloss.Style
	pillOn       lipgloss.Style
	progressDone lipgloss.Style
	progressTodo lipgloss.Style
	overlay      lipgloss.Style
	overlayTitle lipgloss.Style
	itemTitle    lipgloss.Style
	itemDetail   lipgloss.Style
	itemSelected lipgloss.Style
	helpKey      lipgloss.Style
	helpDesc     lipgloss.Style
}

func newStyles(dark bool) styles {
	p := newPalette(dark)
	base := lipgloss.NewStyle()
	return styles{
		palette:      p,
		title:        base.Foreground(p.gold).Bold(true),
		subtitle:     base.Foreground(p.text),
		muted:        base.Foreground(p.muted),
		faint:        base.Foreground(p.faint),
		arabic:       base.Foreground(p.arabic),
		recited:      base.Foreground(p.onGold).Background(p.gold).Bold(true),
		marker:       base.Foreground(p.gold),
		translation:  base.Foreground(p.muted),
		key:          base.Foreground(p.faint),
		keyPlaying:   base.Foreground(p.gold).Bold(true),
		gutter:       base.Foreground(p.faint),
		gutterCursor: base.Foreground(p.accent).Bold(true),
		gutterPlay:   base.Foreground(p.gold).Bold(true),
		banner:       base.Foreground(p.muted),
		bannerArabic: base.Foreground(p.gold).Bold(true),
		bismillah:    base.Foreground(p.accent),
		sidebar:      base.BorderStyle(lipgloss.NormalBorder()).BorderRight(true).BorderForeground(p.faint),
		sideItem:     base.Foreground(p.muted),
		sideCurrent:  base.Foreground(p.gold).Bold(true),
		sideCursor:   base.Foreground(p.text).Background(p.selection).Bold(true),
		sideArabic:   base.Foreground(p.faint),
		rule:         base.Foreground(p.faint),
		status:       base.Foreground(p.muted),
		statusError:  base.Foreground(p.danger),
		pill:         base.Foreground(p.muted).Padding(0, 1),
		pillOn:       base.Foreground(p.onGold).Background(p.accent).Padding(0, 1),
		progressDone: base.Foreground(p.gold),
		progressTodo: base.Foreground(p.faint),
		overlay:      base.Border(lipgloss.RoundedBorder()).BorderForeground(p.accent).Padding(0, 1),
		overlayTitle: base.Foreground(p.gold).Bold(true),
		itemTitle:    base.Foreground(p.text),
		itemDetail:   base.Foreground(p.muted),
		itemSelected: base.Foreground(p.text).Background(p.selection).Bold(true),
		helpKey:      base.Foreground(p.accent).Bold(true),
		helpDesc:     base.Foreground(p.muted),
	}
}
