package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type pickerKind int

const (
	pickSearch pickerKind = iota
	pickReciter
	pickTranslation
	pickSettings
)

type settingID int

const (
	settingReciter settingID = iota
	settingTranslation
	settingArabic
	settingSidebar
	settingTranslate
	settingRepeat
	settingAutoplay
	settingVolume
)

type pickerItem struct {
	title  string
	detail string
	// value is what choosing the item acts on: a quran.Target, a reciter
	// slug or a translation id.
	value any
}

// picker is a filterable list shown over the reader.
type picker struct {
	kind   pickerKind
	title  string
	input  textinput.Model
	items  []pickerItem
	cursor int
	offset int
	// source rebuilds items from the query.
	source func(query string) []pickerItem
}

const pickerRows = 12

func newPicker(kind pickerKind, title string, placeholder string, source func(string) []pickerItem) *picker {
	input := textinput.New()
	input.Prompt = "❯ "
	input.Placeholder = placeholder
	input.Focus()
	p := &picker{kind: kind, title: title, input: input, source: source}
	p.refresh()
	return p
}

func (p *picker) refresh() {
	p.items = p.source(p.input.Value())
	p.cursor, p.offset = 0, 0
}

func (p *picker) selected() (pickerItem, bool) {
	if len(p.items) == 0 {
		return pickerItem{}, false
	}
	return p.items[p.cursor], true
}

func (p *picker) move(delta int) {
	if len(p.items) == 0 {
		return
	}
	p.cursor = min(len(p.items)-1, max(0, p.cursor+delta))
	if p.cursor < p.offset {
		p.offset = p.cursor
	}
	if p.cursor >= p.offset+pickerRows {
		p.offset = p.cursor - pickerRows + 1
	}
}

func (p *picker) update(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "up", "ctrl+p", "ctrl+k":
		p.move(-1)
		return nil
	case "down", "ctrl+n", "ctrl+j":
		p.move(1)
		return nil
	case "pgup":
		p.move(-pickerRows)
		return nil
	case "pgdown":
		p.move(pickerRows)
		return nil
	}
	before := p.input.Value()
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	if p.input.Value() != before {
		p.refresh()
	}
	return cmd
}

func (p *picker) view(s styles, width int) string {
	inner := width - 4
	p.input.SetWidth(inner - 2)
	rows := []string{s.overlayTitle.Render(p.title), p.input.View(), s.rule.Render(strings.Repeat("─", inner))}
	if len(p.items) == 0 {
		rows = append(rows, s.muted.Render("no matches"))
	}
	end := min(len(p.items), p.offset+pickerRows)
	for i := p.offset; i < end; i++ {
		rows = append(rows, pickerRow(s, p.items[i], inner, i == p.cursor))
	}
	if len(p.items) > pickerRows {
		rows = append(rows, s.faint.Render(ansi.Truncate(countLabel(p.cursor+1, len(p.items)), inner, "…")))
	}
	return s.overlay.Width(width).Render(strings.Join(rows, "\n"))
}

func pickerRow(s styles, item pickerItem, width int, selected bool) string {
	title := ansi.Truncate(item.title, width-2, "…")
	detailWidth := width - 3 - ansi.StringWidth(title)
	detail := ""
	if detailWidth > 4 && item.detail != "" {
		detail = ansi.Truncate(item.detail, detailWidth, "…")
	}
	gap := strings.Repeat(" ", max(1, width-2-ansi.StringWidth(title)-ansi.StringWidth(detail)))
	if selected {
		return s.itemSelected.Width(width).Render("▸ " + title + gap + detail)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, "  ", s.itemTitle.Render(title), gap, s.itemDetail.Render(detail))
}

func countLabel(position int, total int) string {
	return fmt.Sprintf("%d/%d", position, total)
}
