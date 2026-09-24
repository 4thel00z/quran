package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/4thel00z/quran/internal/quran"
	"github.com/4thel00z/quran/internal/render"
)

const (
	sidebarWidth    = 32
	minSidebarWidth = 96
	headerHeight    = 2
	footerHeight    = 3
	gutterWidth     = 3
	bismillah       = "بِسْمِ ٱللَّهِ ٱلرَّحْمَـٰنِ ٱلرَّحِيمِ"
)

func (m *Model) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = "quran · " + m.book.Surah(m.surah).NameEnglish
	if m.width == 0 || m.height == 0 {
		return v
	}
	body := m.readerView()
	if m.sidebarVisible() {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.sidebarView(), body)
	}
	screen := lipgloss.JoinVertical(lipgloss.Left, m.headerView(), body, m.footerView())
	if m.overlay != nil {
		screen = m.withOverlay(screen, m.overlay.view(m.styles, min(84, m.readerWidth()-4)))
	}
	if m.showHelp {
		screen = m.withOverlay(screen, m.helpView())
	}
	v.SetContent(screen)
	return v
}

// withOverlay centers box over the reader, never over the sidebar divider.
func (m *Model) withOverlay(screen string, box string) string {
	left := m.width - m.readerWidth()
	x := left + max(0, (m.readerWidth()-lipgloss.Width(box))/2)
	y := max(0, (m.height-lipgloss.Height(box))/3)
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(screen),
		lipgloss.NewLayer(box).X(x).Y(y).Z(1),
	).Render()
}

func (m *Model) sidebarVisible() bool {
	return m.showSidebar && m.width >= minSidebarWidth
}

func (m *Model) readerWidth() int {
	if !m.sidebarVisible() {
		return m.width
	}
	return m.width - sidebarWidth - 1
}

func (m *Model) textWidth() int {
	return max(20, m.readerWidth()-gutterWidth-2)
}

func (m *Model) bodyHeight() int {
	return max(1, m.height-headerHeight-footerHeight)
}

func (m *Model) headerView() string {
	s := m.styles
	surah := m.book.Surah(m.surah)
	left := s.title.Render("۞ quran") + s.faint.Render("  ·  ") +
		s.subtitle.Render(fmt.Sprintf("%d. %s", surah.Number, surah.NameEnglish)) +
		s.muted.Render(" "+surah.NameMeaning)
	ayah := m.currentAyah()
	right := s.muted.Render(fmt.Sprintf("Juz %d · Hizb %d · Page %d", ayah.Juz, ayah.Hizb, ayah.Page))
	if m.rangeStart != nil {
		right = s.marker.Render(fmt.Sprintf("▶ %s → %s", m.rangeStart, m.rangeEnd)) + s.faint.Render("  ") + right
	}
	gap := max(1, m.width-lipgloss.Width(left)-lipgloss.Width(right)-2)
	line := " " + left + strings.Repeat(" ", gap) + right + " "
	return line + "\n" + s.rule.Render(strings.Repeat("─", m.width))
}

// currentAyah is the ayah whose metadata the header shows.
func (m *Model) currentAyah() quran.Ayah {
	key := m.cursorKey()
	if m.play.active && m.play.key.Surah == m.surah {
		key = m.play.key
	}
	index, err := m.book.Index(key)
	if err != nil {
		return m.book.Ayahs[0]
	}
	return m.book.Ayahs[index]
}

func (m *Model) sidebarView() string {
	s := m.styles
	rows := m.bodyHeight()
	if m.focus != focusSidebar {
		m.sideOffset = min(max(0, m.surah-1-rows/2), max(0, quran.SurahCount-rows))
	}
	lines := make([]string, 0, rows)
	inner := sidebarWidth - 1
	for i := m.sideOffset; i < min(quran.SurahCount, m.sideOffset+rows); i++ {
		surah := m.book.Surahs[i]
		number := fmt.Sprintf("%3d ", surah.Number)
		name := m.mode.Word(surah.NameArabic)
		english := ansi.Truncate(surah.NameEnglish, inner-len(number)-ansi.StringWidth(name)-2, "…")
		gap := strings.Repeat(" ", max(1, inner-len(number)-ansi.StringWidth(english)-ansi.StringWidth(name)-1))
		switch {
		case m.focus == focusSidebar && i == m.sideCursor:
			lines = append(lines, s.sideCursor.Width(inner).Render(number+english+gap+name))
		case surah.Number == m.surah:
			lines = append(lines, s.sideCurrent.Render(number+english)+gap+s.marker.Render(name))
		default:
			lines = append(lines, s.faint.Render(number)+s.sideItem.Render(english)+gap+s.sideArabic.Render(name))
		}
	}
	for len(lines) < rows {
		lines = append(lines, "")
	}
	divider := lipgloss.Border{Right: m.mode.Divider()}
	return s.sidebar.Border(divider, false, true, false, false).Width(sidebarWidth).Height(rows).Render(strings.Join(lines, "\n"))
}

func (m *Model) readerView() string {
	rows := m.bodyHeight()
	lines := make([]string, 0, rows)
	offset := m.scroll
	for i := -1; i < m.book.Surah(m.surah).AyahCount && len(lines) < rows; i++ {
		block := m.block(i)
		if offset >= len(block) {
			offset -= len(block)
			continue
		}
		lines = append(lines, block[offset:]...)
		offset = 0
	}
	if len(lines) > rows {
		lines = lines[:rows]
	}
	for len(lines) < rows {
		lines = append(lines, "")
	}
	return lipgloss.NewStyle().Width(m.readerWidth()).PaddingLeft(1).Render(strings.Join(lines, "\n"))
}

// ensureVisible scrolls so the cursor ayah is on screen.
func (m *Model) ensureVisible() {
	if m.width == 0 {
		return
	}
	top := 0
	for i := -1; i < m.cursor; i++ {
		top += len(m.block(i))
	}
	height := len(m.block(m.cursor))
	rows := m.bodyHeight()
	if m.cursor == 0 {
		top = 0
	}
	if top < m.scroll {
		m.scroll = top
	}
	if top+height > m.scroll+rows {
		m.scroll = top + min(height, rows) - rows
		m.scroll = max(m.scroll, top+height-rows)
	}
	if height > rows {
		m.scroll = top
	}
}

// scrollToCursor puts the cursor ayah a quarter of the way down the reader.
func (m *Model) scrollToCursor() {
	if m.width == 0 {
		return
	}
	top := 0
	for i := -1; i < m.cursor; i++ {
		top += len(m.block(i))
	}
	if m.cursor == 0 {
		top = 0
	}
	m.scroll = max(0, top-m.bodyHeight()/4)
}

// block renders ayah i of the open surah; -1 is the surah banner.
func (m *Model) block(i int) []string {
	if lines, ok := m.blocks[i]; ok {
		return lines
	}
	lines := m.ayahBlock(i)
	if i < 0 {
		lines = m.bannerBlock()
	}
	m.blocks[i] = lines
	return lines
}

func (m *Model) bannerBlock() []string {
	s := m.styles
	surah := m.book.Surah(m.surah)
	width := m.textWidth() + gutterWidth
	center := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
	ornament := s.marker.Render("۞")
	lines := []string{
		"",
		center.Render(s.bannerArabic.Render(m.mode.Word("سورة "+surah.NameArabic))),
		center.Render(s.subtitle.Render(surah.NameEnglish) + s.muted.Render(" · "+surah.NameMeaning)),
		center.Render(s.faint.Render(fmt.Sprintf("%s · %d ayahs · revealed #%d", surah.Revelation, surah.AyahCount, surah.RevelationOrder))),
		center.Render(s.faint.Render(strings.Repeat("─", 12)) + " " + ornament + " " + s.faint.Render(strings.Repeat("─", 12))),
	}
	if surah.Number != 1 && surah.Number != 9 {
		words := strings.Fields(bismillah)
		reciting := m.play.active && m.play.bismillah && m.play.key.Surah == surah.Number
		style := func(w int) lipgloss.Style {
			if reciting && w == m.play.word {
				return s.recited
			}
			return s.bismillah
		}
		for _, line := range render.RTLLines(words, width, m.mode, style, lipgloss.NewStyle()) {
			lines = append(lines, center.Render(strings.TrimLeft(line, " ")))
		}
	}
	return append(lines, "")
}

func (m *Model) ayahBlock(i int) []string {
	s := m.styles
	key := quran.Key{Surah: m.surah, Ayah: i + 1}
	index, err := m.book.Index(key)
	if err != nil {
		return nil
	}
	ayah := m.book.Ayahs[index]
	playing := m.play.active && m.play.key == key
	words := append(append([]string{}, ayah.Words...), render.AyahMarker(key.Ayah))
	markerIndex := len(words) - 1
	style := func(w int) lipgloss.Style {
		switch {
		case w == markerIndex:
			return s.marker
		case playing && w == m.play.word:
			return s.recited
		}
		return s.arabic
	}
	width := m.textWidth()
	lines := []string{m.ayahHeading(ayah, playing, i == m.cursor, width)}
	lines = append(lines, render.RTLLines(words, width, m.mode, style, lipgloss.NewStyle())...)
	if m.showTranslate {
		text := m.verses[key.Surah-1][key.Ayah-1]
		lines = append(lines, render.TextLines(text, width, m.mode, s.translation)...)
	}
	gutter := s.gutter.Render("  ")
	switch {
	case playing:
		gutter = s.gutterPlay.Render("┃ ")
	case i == m.cursor:
		gutter = s.gutterCursor.Render("┃ ")
	}
	for j, line := range lines {
		lines[j] = gutter + line
	}
	return append(lines, "")
}

func (m *Model) ayahHeading(ayah quran.Ayah, playing bool, selected bool, width int) string {
	s := m.styles
	label := s.key.Render(ayah.Key.String())
	switch {
	case playing:
		label = s.keyPlaying.Render("♪ " + ayah.Key.String())
	case selected:
		label = s.gutterCursor.Render(ayah.Key.String())
	}
	marks := []string{}
	if ayah.Ayah > 1 || ayah.Surah > 1 {
		previous := m.book.Ayahs[max(0, m.mustIndex(ayah.Key)-1)]
		if previous.Juz != ayah.Juz {
			marks = append(marks, fmt.Sprintf("Juz %d", ayah.Juz))
		}
		if previous.Hizb != ayah.Hizb {
			marks = append(marks, fmt.Sprintf("Hizb %d", ayah.Hizb))
		}
		if previous.Page != ayah.Page {
			marks = append(marks, fmt.Sprintf("p.%d", ayah.Page))
		}
	}
	right := s.faint.Render(strings.Join(marks, " · "))
	fill := max(1, width-lipgloss.Width(label)-lipgloss.Width(right)-2)
	return label + " " + s.rule.Render(strings.Repeat("┄", fill)) + " " + right
}

func (m *Model) mustIndex(key quran.Key) int {
	index, err := m.book.Index(key)
	if err != nil {
		return 0
	}
	return index
}

func (m *Model) footerView() string {
	s := m.styles
	rule := s.rule.Render(strings.Repeat("─", m.width))
	return rule + "\n" + m.playerLine() + "\n" + m.statusLine()
}

func (m *Model) playerLine() string {
	s := m.styles
	icon, key := s.muted.Render("■"), s.muted.Render("stopped")
	position, duration := time.Duration(0), time.Duration(0)
	switch {
	case m.play.loading:
		icon, key = s.marker.Render("◌"), s.subtitle.Render(m.play.key.String())
	case m.play.active && m.player.Paused():
		icon, key = s.marker.Render("⏸"), s.subtitle.Render(m.play.key.String())
		position, duration = m.player.Position(), m.player.Duration()
	case m.play.active:
		icon, key = s.gutterPlay.Render("▶"), s.subtitle.Render(m.play.key.String())
		position, duration = m.player.Position(), m.player.Duration()
	}
	reciter := s.title.Render(m.reciter.Name) + s.muted.Render(" · "+m.reciter.Style)
	pills := strings.Join([]string{
		pill(s, "autoplay", m.autoplay),
		pill(s, m.repeat.String(), m.repeat != repeatOff),
		s.muted.Render(fmt.Sprintf("vol %d%%", m.player.Percent())),
	}, " ")
	times := s.muted.Render(clock(position) + " / " + clock(duration))
	left := " " + icon + " " + key + "  " + reciter + "  "
	right := "  " + times + "  " + pills + " "
	if m.width-lipgloss.Width(left)-lipgloss.Width(right) < 12 {
		right = "  " + times + " "
	}
	barWidth := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if barWidth < 8 {
		return ansi.Truncate(left+right, m.width, "…")
	}
	return left + progressBar(s, position, duration, barWidth) + right
}

func pill(s styles, label string, on bool) string {
	if !on {
		return s.pill.Render(label)
	}
	return s.pillOn.Render(label)
}

func progressBar(s styles, position time.Duration, duration time.Duration, width int) string {
	done := 0
	if duration > 0 {
		done = min(width, int(float64(width)*float64(position)/float64(duration)))
	}
	if done == 0 {
		return s.progressTodo.Render(strings.Repeat("─", width))
	}
	return s.progressDone.Render(strings.Repeat("━", done-1)+"●") + s.progressTodo.Render(strings.Repeat("─", width-done))
}

func clock(d time.Duration) string {
	seconds := int(d.Round(time.Second).Seconds())
	return fmt.Sprintf("%d:%02d", seconds/60, seconds%60)
}

func (m *Model) statusLine() string {
	s := m.styles
	hints := s.faint.Render(" space play · enter play here · n/p next/prev · / search · R reciter · T translation · ? help · q quit")
	if m.status == "" {
		return ansi.Truncate(hints, m.width, "…")
	}
	style := s.status
	if m.failed {
		style = s.statusError
	}
	return ansi.Truncate(" "+style.Render(m.status)+s.faint.Render("  ·  ? help"), m.width, "…")
}

type binding struct {
	keys string
	desc string
}

var bindings = [][]binding{
	{
		{"space", "play / pause"},
		{"enter", "play ayah under cursor"},
		{"n / p", "next / previous ayah"},
		{"s", "stop"},
		{"c", "jump to reciting ayah"},
		{"r", "repeat: off · ayah · range"},
		{"a", "autoplay next ayah"},
		{"+ / -", "volume"},
	},
	{
		{"j / k", "next / previous ayah"},
		{"h / l", "previous / next surah"},
		{"g / G", "first / last ayah"},
		{"tab", "surah list"},
		{"/", "search: 2:255, juz 30, text"},
		{"R / T", "reciter / translation"},
		{"t / b", "toggle translation / sidebar"},
		{"A", "arabic: visual ↔ native"},
	},
}

func (m *Model) helpView() string {
	s := m.styles
	columns := []string{}
	for _, column := range bindings {
		rows := []string{}
		for _, b := range column {
			rows = append(rows, s.helpKey.Width(8).Render(b.keys)+" "+s.helpDesc.Render(b.desc))
		}
		columns = append(columns, strings.Join(rows, "\n"))
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, columns[0], "    ", columns[1])
	footer := s.faint.Render("text: Tarteel quran-assets · audio: quran.host · ? to close")
	return s.overlay.Render(s.overlayTitle.Render("Keys") + "\n\n" + body + "\n\n" + footer)
}
