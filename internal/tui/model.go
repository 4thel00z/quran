// Package tui is the interactive reader and player.
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/4thel00z/quran/internal/assets"
	"github.com/4thel00z/quran/internal/audio"
	"github.com/4thel00z/quran/internal/config"
	"github.com/4thel00z/quran/internal/quran"
	"github.com/4thel00z/quran/internal/render"
)

// Options configure a session; Target and Play pick where it opens.
type Options struct {
	Book        quran.Book
	Reciter     quran.Reciter
	Translation quran.Translation
	Mode        render.Mode
	Fetcher     *audio.Fetcher
	Player      *audio.Player
	Target      *quran.Target
	Play        bool
	Config      config.Config
}

type focus int

const (
	focusReader focus = iota
	focusSidebar
)

type repeatMode int

const (
	repeatOff repeatMode = iota
	repeatAyah
	repeatRange
)

func (r repeatMode) next() repeatMode {
	return (r + 1) % 3
}

func (r repeatMode) String() string {
	switch r {
	case repeatAyah:
		return "repeat ayah"
	case repeatRange:
		return "repeat range"
	}
	return "repeat off"
}

type playback struct {
	active  bool
	loading bool
	key     quran.Key
	word    int
	// bismillah is set while the opening basmala of key's surah plays.
	bismillah bool
}

var basmala = quran.Key{Surah: 1, Ayah: 1}

// needsBismillah reports whether key opens a surah recited after a basmala.
func needsBismillah(key quran.Key) bool {
	return key.Ayah == 1 && key.Surah != 1 && key.Surah != 9
}

// audioKey is the recording currently loaded or playing.
func (p playback) audioKey() quran.Key {
	if p.bismillah {
		return basmala
	}
	return p.key
}

type Model struct {
	book        quran.Book
	reciter     quran.Reciter
	segments    quran.Segments
	translation quran.Translation
	verses      [][]string
	mode        render.Mode
	fetcher     *audio.Fetcher
	player      *audio.Player

	styles styles
	dark   bool
	width  int
	height int

	surah  int
	cursor int
	scroll int
	blocks map[int][]string

	focus         focus
	sideCursor    int
	sideOffset    int
	showSidebar   bool
	showTranslate bool
	showHelp      bool

	play     playback
	gen      int
	tickID   int
	autoplay bool
	repeat   repeatMode
	// rangeStart and rangeEnd bound continuous playback after a jump to a
	// juz, hizb, page or ayah range.
	rangeStart *quran.Key
	rangeEnd   *quran.Key

	// scrollPending defers scrollToCursor until the window size is known.
	scrollPending bool

	overlay *picker
	status  string
	failed  bool
	start   *quran.Target
	playNow bool
}

func New(opts Options) (*Model, error) {
	segments, err := assets.Segments(opts.Reciter.Slug)
	if err != nil {
		return nil, err
	}
	verses, err := assets.Translation(opts.Translation.ID)
	if err != nil {
		return nil, err
	}

	showSidebar := true
	if opts.Config.ShowSidebar != nil {
		showSidebar = *opts.Config.ShowSidebar
	}
	showTranslate := true
	if opts.Config.ShowTranslation != nil {
		showTranslate = *opts.Config.ShowTranslation
	}
	if opts.Config.Volume != nil {
		opts.Player.SetLevel(*opts.Config.Volume)
	}
	var rep repeatMode
	switch opts.Config.Repeat {
	case "ayah":
		rep = repeatAyah
	case "range":
		rep = repeatRange
	}

	m := &Model{
		book:          opts.Book,
		reciter:       opts.Reciter,
		segments:      segments,
		translation:   opts.Translation,
		verses:        verses,
		mode:          opts.Mode,
		fetcher:       opts.Fetcher,
		player:        opts.Player,
		dark:          true,
		styles:        newStyles(true),
		surah:         1,
		blocks:        map[int][]string{},
		showSidebar:   showSidebar,
		showTranslate: showTranslate,
		repeat:        rep,
		autoplay:      true,
		play:          playback{word: -1},
		start:         opts.Target,
		playNow:       opts.Play,
	}
	return m, nil
}

func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.RequestBackgroundColor}
	if m.start == nil {
		return tea.Batch(cmds...)
	}
	m.jump(*m.start)
	if m.playNow {
		cmds = append(cmds, m.playKey(m.cursorKey()))
	}
	return tea.Batch(cmds...)
}

type audioMsg struct {
	gen  int
	key  quran.Key
	data []byte
	err  error
}

type doneMsg struct{ gen int }

type tickMsg struct{ id int }

type prefetchedMsg struct{}

const tickInterval = 50 * time.Millisecond

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.dark = msg.IsDark()
		m.styles = newStyles(m.dark)
		m.invalidate()
		return m, nil
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.invalidate()
		m.ensureVisible()
		if m.scrollPending {
			m.scrollPending = false
			m.scrollToCursor()
		}
		return m, nil
	case audioMsg:
		return m, m.onAudio(msg)
	case doneMsg:
		if msg.gen != m.gen {
			return m, nil
		}
		if m.play.bismillah {
			return m, m.load(m.play.key, false)
		}
		return m, m.onFinished()
	case tickMsg:
		if msg.id != m.tickID || !m.play.active || m.play.loading {
			return m, nil
		}
		m.updateWord()
		return m, m.tick()
	case prefetchedMsg:
		return m, nil
	case tea.MouseWheelMsg:
		return m, m.onWheel(msg)
	case tea.MouseClickMsg:
		return m, m.onClick(msg)
	case tea.KeyPressMsg:
		if m.overlay != nil {
			return m, m.onOverlayKey(msg)
		}
		return m, m.onKey(msg)
	}
	return m, nil
}

func (m *Model) onWheel(msg tea.MouseWheelMsg) tea.Cmd {
	if m.overlay != nil {
		switch msg.Button {
		case tea.MouseWheelUp:
			m.overlay.move(-1)
		case tea.MouseWheelDown:
			m.overlay.move(1)
		}
		return nil
	}
	if m.sidebarVisible() && msg.X < sidebarWidth {
		switch msg.Button {
		case tea.MouseWheelUp:
			m.moveSide(-3)
		case tea.MouseWheelDown:
			m.moveSide(3)
		}
		return nil
	}
	switch msg.Button {
	case tea.MouseWheelUp:
		m.moveCursor(-1)
	case tea.MouseWheelDown:
		m.moveCursor(1)
	}
	return nil
}

func (m *Model) onClick(msg tea.MouseClickMsg) tea.Cmd {
	if msg.Button != tea.MouseLeft {
		return nil
	}

	if m.showHelp {
		m.showHelp = false
		return nil
	}

	if m.overlay != nil {
		boxWidth := min(84, m.readerWidth()-4)
		left := m.width - m.readerWidth()
		boxX := left + max(0, (m.readerWidth()-boxWidth)/2)

		totalRows := 3 + min(len(m.overlay.items), pickerRows)
		if len(m.overlay.items) > pickerRows {
			totalRows++
		}
		boxHeight := totalRows + 2
		boxY := max(0, (m.height-boxHeight)/3)

		if msg.X < boxX || msg.X >= boxX+boxWidth || msg.Y < boxY || msg.Y >= boxY+boxHeight {
			m.overlay = nil
			return nil
		}

		itemRow := msg.Y - (boxY + 3)
		visibleCount := min(len(m.overlay.items)-m.overlay.offset, pickerRows)
		if itemRow >= 0 && itemRow < visibleCount {
			clickedIdx := m.overlay.offset + itemRow
			if clickedIdx == m.overlay.cursor {
				item, ok := m.overlay.selected()
				kind := m.overlay.kind
				if kind != pickSettings {
					m.overlay = nil
				}
				if ok {
					return m.choose(kind, item)
				}
			} else {
				m.overlay.cursor = clickedIdx
			}
		}
		return nil
	}

	if m.sidebarVisible() && msg.X < sidebarWidth {
		row := msg.Y - headerHeight
		if row >= 0 && row < m.bodyHeight() {
			surahIdx := m.sideOffset + row
			if surahIdx >= 0 && surahIdx < quran.SurahCount {
				m.sideCursor = surahIdx
				m.openSurah(surahIdx + 1)
				m.focus = focusSidebar
			}
		}
		return nil
	}

	if msg.X >= (m.width - m.readerWidth()) {
		row := msg.Y - headerHeight
		if row >= 0 && row < m.bodyHeight() {
			m.focus = focusReader
			ayahIdx := m.ayahAtRow(row)
			if ayahIdx >= 0 {
				if ayahIdx == m.cursor {
					m.clearRange()
					return m.playKey(m.cursorKey())
				}
				m.setCursor(ayahIdx)
			}
		}
		return nil
	}

	return nil
}

func (m *Model) ayahAtRow(row int) int {
	if row < 0 || row >= m.bodyHeight() {
		return -1
	}
	offset := m.scroll
	currentLine := 0
	for i := -1; i < m.book.Surah(m.surah).AyahCount; i++ {
		block := m.block(i)
		if offset >= len(block) {
			offset -= len(block)
			continue
		}
		visibleInBlock := len(block) - offset
		offset = 0
		if row >= currentLine && row < currentLine+visibleInBlock {
			if i == -1 {
				return 0
			}
			return i
		}
		currentLine += visibleInBlock
		if currentLine >= m.bodyHeight() {
			break
		}
	}
	return -1
}

func (m *Model) onKey(msg tea.KeyPressMsg) tea.Cmd {
	key := msg.String()
	if m.focus == focusSidebar {
		if cmd, handled := m.onSidebarKey(key); handled {
			return cmd
		}
	}
	switch key {
	case "q", "ctrl+c":
		m.player.Stop()
		return tea.Quit
	case "?":
		m.showHelp = !m.showHelp
	case "tab":
		m.toggleFocus()
	case "b":
		m.showSidebar = !m.showSidebar
		m.focus = focusReader
		m.invalidate()
		m.saveConfig()
	case "j", "down":
		m.moveCursor(1)
	case "k", "up":
		m.moveCursor(-1)
	case "ctrl+d", "pgdown":
		m.moveCursor(10)
	case "ctrl+u", "pgup":
		m.moveCursor(-10)
	case "g", "home":
		m.setCursor(0)
	case "G", "end":
		m.setCursor(m.book.Surah(m.surah).AyahCount - 1)
	case "l", "right", "]":
		m.openSurah(m.surah + 1)
	case "h", "left", "[":
		m.openSurah(m.surah - 1)
	case "space":
		return m.togglePlay()
	case "enter":
		m.clearRange()
		return m.playKey(m.cursorKey())
	case "n":
		return m.step(1)
	case "p":
		return m.step(-1)
	case "s":
		m.stop()
	case "c":
		m.followPlayback()
	case "r":
		m.repeat = m.repeat.next()
		m.notify(m.repeat.String())
		m.saveConfig()
	case "a":
		m.autoplay = !m.autoplay
		m.notify(onOff("autoplay", m.autoplay))
		m.saveConfig()
	case "t":
		m.showTranslate = !m.showTranslate
		m.invalidate()
		m.saveConfig()
	case "A":
		m.mode = m.mode.Toggle()
		m.invalidate()
		m.notify("arabic: " + string(m.mode))
		m.saveConfig()
	case "+", "=":
		m.player.SetLevel(m.player.Level() + 1)
		m.notify(fmt.Sprintf("volume %d%%", m.player.Percent()))
		m.saveConfig()
	case "-", "_":
		m.player.SetLevel(m.player.Level() - 1)
		m.notify(fmt.Sprintf("volume %d%%", m.player.Percent()))
		m.saveConfig()
	case "/":
		m.overlay = newPicker(pickSearch, "Search", "2:255 · juz 30 · hizb 5 · page 12 · kahf · text", m.searchItems)
	case "R":
		m.overlay = newPicker(pickReciter, "Reciter", "filter reciters", m.reciterItems)
	case "T":
		m.overlay = newPicker(pickTranslation, "Translation", "filter translations", m.translationItems)
	case "S", ",":
		m.overlay = newPicker(pickSettings, "Settings", "filter settings", m.settingsItems)
	}
	return nil
}

func (m *Model) onSidebarKey(key string) (tea.Cmd, bool) {
	switch key {
	case "j", "down":
		m.moveSide(1)
	case "k", "up":
		m.moveSide(-1)
	case "ctrl+d", "pgdown":
		m.moveSide(10)
	case "ctrl+u", "pgup":
		m.moveSide(-10)
	case "g", "home":
		m.moveSide(-quran.SurahCount)
	case "G", "end":
		m.moveSide(quran.SurahCount)
	case "enter", "l", "right":
		m.openSurah(m.sideCursor + 1)
		m.focus = focusReader
	case "esc":
		m.focus = focusReader
	default:
		return nil, false
	}
	return nil, true
}

func (m *Model) onOverlayKey(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.overlay = nil
		return nil
	case "enter":
		item, ok := m.overlay.selected()
		kind := m.overlay.kind
		if kind != pickSettings {
			m.overlay = nil
		}
		if !ok {
			return nil
		}
		return m.choose(kind, item)
	case "left", "h", "-":
		if m.overlay.kind == pickSettings {
			item, ok := m.overlay.selected()
			if ok && item.value == settingVolume {
				m.player.SetLevel(m.player.Level() - 1)
				m.saveConfig()
				cur := m.overlay.cursor
				m.overlay = newPicker(pickSettings, "Settings", "filter settings", m.settingsItems)
				m.overlay.cursor = cur
				return nil
			}
		}
	case "right", "l", "+", "=":
		if m.overlay.kind == pickSettings {
			item, ok := m.overlay.selected()
			if ok && item.value == settingVolume {
				m.player.SetLevel(m.player.Level() + 1)
				m.saveConfig()
				cur := m.overlay.cursor
				m.overlay = newPicker(pickSettings, "Settings", "filter settings", m.settingsItems)
				m.overlay.cursor = cur
				return nil
			}
		}
	}
	return m.overlay.update(msg)
}

func (m *Model) choose(kind pickerKind, item pickerItem) tea.Cmd {
	switch kind {
	case pickReciter:
		return m.setReciter(item.value.(string))
	case pickTranslation:
		m.setTranslation(item.value.(string))
		return nil
	case pickSettings:
		id := item.value.(settingID)
		switch id {
		case settingReciter:
			m.overlay = newPicker(pickReciter, "Reciter", "filter reciters", m.reciterItems)
			return nil
		case settingTranslation:
			m.overlay = newPicker(pickTranslation, "Translation", "filter translations", m.translationItems)
			return nil
		case settingArabic:
			m.mode = m.mode.Toggle()
			m.invalidate()
			m.saveConfig()
		case settingSidebar:
			m.showSidebar = !m.showSidebar
			m.focus = focusReader
			m.invalidate()
			m.saveConfig()
		case settingTranslate:
			m.showTranslate = !m.showTranslate
			m.invalidate()
			m.saveConfig()
		case settingRepeat:
			m.repeat = m.repeat.next()
			m.saveConfig()
		case settingAutoplay:
			m.autoplay = !m.autoplay
			m.saveConfig()
		case settingVolume:
			lvl := m.player.Level() + 2
			if lvl > 10 {
				lvl = 0
			}
			m.player.SetLevel(lvl)
			m.saveConfig()
		}
		cur := 0
		if m.overlay != nil {
			cur = m.overlay.cursor
		}
		m.overlay = newPicker(pickSettings, "Settings", "filter settings", m.settingsItems)
		m.overlay.cursor = cur
		return nil
	}
	target := item.value.(quran.Target)
	m.jump(target)
	return nil
}

func (m *Model) toggleFocus() {
	if !m.showSidebar {
		return
	}
	if m.focus == focusSidebar {
		m.focus = focusReader
		return
	}
	m.focus = focusSidebar
	m.sideCursor = m.surah - 1
	m.moveSide(0)
}

func (m *Model) moveSide(delta int) {
	m.sideCursor = min(quran.SurahCount-1, max(0, m.sideCursor+delta))
	rows := max(1, m.bodyHeight())
	if m.sideCursor < m.sideOffset {
		m.sideOffset = m.sideCursor
	}
	if m.sideCursor >= m.sideOffset+rows {
		m.sideOffset = m.sideCursor - rows + 1
	}
}

func (m *Model) cursorKey() quran.Key {
	return quran.Key{Surah: m.surah, Ayah: m.cursor + 1}
}

func (m *Model) moveCursor(delta int) {
	m.setCursor(m.cursor + delta)
}

func (m *Model) setCursor(index int) {
	count := m.book.Surah(m.surah).AyahCount
	index = min(count-1, max(0, index))
	if index == m.cursor {
		return
	}
	delete(m.blocks, m.cursor)
	delete(m.blocks, index)
	m.cursor = index
	m.ensureVisible()
}

func (m *Model) openSurah(number int) {
	if number < 1 || number > quran.SurahCount {
		return
	}
	if number != m.surah {
		m.surah = number
		m.blocks = map[int][]string{}
		m.scroll = 0
	}
	m.cursor = 0
	m.sideCursor = number - 1
	m.moveSide(0)
	m.ensureVisible()
}

func (m *Model) jump(target quran.Target) {
	m.openSurah(target.Start.Surah)
	m.setCursor(target.Start.Ayah - 1)
	m.scrollPending = m.width == 0
	m.scrollToCursor()
	m.clearRange()
	if target.Kind != quran.TargetSurah && target.Kind != quran.TargetAyah {
		start, end := target.Start, target.End
		m.rangeStart, m.rangeEnd = &start, &end
	}
	m.notify(target.String())
}

func (m *Model) clearRange() {
	m.rangeStart, m.rangeEnd = nil, nil
}

// followPlayback moves the cursor to the ayah being recited.
func (m *Model) followPlayback() {
	if !m.play.active {
		return
	}
	m.openSurah(m.play.key.Surah)
	m.setCursor(m.play.key.Ayah - 1)
}

func (m *Model) togglePlay() tea.Cmd {
	if !m.play.active {
		return m.playKey(m.cursorKey())
	}
	if m.play.loading {
		return nil
	}
	if m.player.TogglePause() {
		m.notify("paused")
		return nil
	}
	m.notify("")
	return m.tick()
}

func (m *Model) step(delta int) tea.Cmd {
	if !m.play.active {
		m.moveCursor(delta)
		return nil
	}
	key, ok := m.neighbour(m.play.key, delta)
	if !ok {
		return nil
	}
	return m.playKey(key)
}

func (m *Model) neighbour(key quran.Key, delta int) (quran.Key, bool) {
	index, err := m.book.Index(key)
	if err != nil {
		return quran.Key{}, false
	}
	index += delta
	if index < 0 || index >= len(m.book.Ayahs) {
		return quran.Key{}, false
	}
	return m.book.Ayahs[index].Key, true
}

func (m *Model) playKey(key quran.Key) tea.Cmd {
	return m.load(key, needsBismillah(key))
}

func (m *Model) load(key quran.Key, bismillah bool) tea.Cmd {
	m.gen++
	m.player.Stop()
	m.markPlaying(key)
	m.play = playback{active: true, loading: true, key: key, word: -1, bismillah: bismillah}
	m.notify("loading " + key.String() + " …")
	gen, reciter, fetcher, audioKey := m.gen, m.reciter, m.fetcher, m.play.audioKey()
	return func() tea.Msg {
		data, err := fetcher.Fetch(context.Background(), reciter, audioKey)
		return audioMsg{gen: gen, key: key, data: data, err: err}
	}
}

func (m *Model) markPlaying(key quran.Key) {
	if m.play.key.Surah == m.surah {
		delete(m.blocks, m.play.key.Ayah-1)
		delete(m.blocks, -1)
	}
	if key.Surah == m.surah {
		delete(m.blocks, key.Ayah-1)
		delete(m.blocks, -1)
	}
	if m.play.active && m.play.key.Surah == m.surah && m.play.key.Ayah-1 == m.cursor {
		m.openSurah(key.Surah)
		m.setCursor(key.Ayah - 1)
	}
}

func (m *Model) onAudio(msg audioMsg) tea.Cmd {
	if msg.gen != m.gen {
		return nil
	}
	if msg.err != nil {
		m.stop()
		m.fail(msg.err)
		return nil
	}
	done, err := m.player.Play(msg.data)
	if err != nil {
		m.stop()
		m.fail(err)
		return nil
	}
	m.play.loading = false
	m.notify("")
	gen := m.gen
	cmds := []tea.Cmd{
		func() tea.Msg {
			<-done
			return doneMsg{gen: gen}
		},
		m.tick(),
	}
	next, ok := m.nextKey(msg.key)
	if m.play.bismillah {
		next, ok = msg.key, true
	}
	if ok {
		reciter, fetcher := m.reciter, m.fetcher
		cmds = append(cmds, func() tea.Msg {
			_, _ = fetcher.Fetch(context.Background(), reciter, next)
			return prefetchedMsg{}
		})
	}
	return tea.Batch(cmds...)
}

// nextKey is the ayah continuous playback moves to after key.
func (m *Model) nextKey(key quran.Key) (quran.Key, bool) {
	if m.repeat == repeatAyah {
		return key, true
	}
	if !m.autoplay {
		return quran.Key{}, false
	}
	if m.rangeEnd != nil && key == *m.rangeEnd {
		if m.repeat == repeatRange {
			return *m.rangeStart, true
		}
		return quran.Key{}, false
	}
	if m.rangeEnd == nil && m.repeat == repeatRange && key.Ayah == m.book.Surah(key.Surah).AyahCount {
		return quran.Key{Surah: key.Surah, Ayah: 1}, true
	}
	return m.neighbour(key, 1)
}

func (m *Model) onFinished() tea.Cmd {
	next, ok := m.nextKey(m.play.key)
	if !ok {
		m.stop()
		return nil
	}
	return m.playKey(next)
}

func (m *Model) stop() {
	m.gen++
	m.player.Stop()
	m.markPlaying(quran.Key{})
	m.play = playback{word: -1}
}

func (m *Model) tick() tea.Cmd {
	m.tickID++
	id := m.tickID
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{id: id} })
}

func (m *Model) updateWord() {
	key := m.play.audioKey()
	index, err := m.book.Index(key)
	if err != nil {
		return
	}
	words := len(m.book.Ayahs[index].Words)
	ms := int(m.player.Position().Milliseconds())
	word := quran.WordAt(m.segments.For(key), ms, words)
	if word == m.play.word {
		return
	}
	m.play.word = word
	if m.play.key.Surah != m.surah {
		return
	}
	if m.play.bismillah {
		delete(m.blocks, -1)
		return
	}
	delete(m.blocks, m.play.key.Ayah-1)
}

func (m *Model) setReciter(slug string) tea.Cmd {
	r, err := quran.FindReciter(slug)
	if err != nil {
		m.fail(err)
		return nil
	}
	segments, err := assets.Segments(slug)
	if err != nil {
		m.fail(err)
		return nil
	}
	m.reciter, m.segments = r, segments
	m.notify("reciter: " + r.Title())
	m.saveConfig()
	if !m.play.active {
		return nil
	}
	return m.playKey(m.play.key)
}

func (m *Model) setTranslation(id string) {
	t, err := quran.FindTranslation(id)
	if err != nil {
		m.fail(err)
		return
	}
	verses, err := assets.Translation(id)
	if err != nil {
		m.fail(err)
		return
	}
	m.translation, m.verses = t, verses
	m.showTranslate = true
	m.invalidate()
	m.notify("translation: " + t.Title())
	m.saveConfig()
}

func (m *Model) notify(text string) {
	m.status, m.failed = text, false
}

func (m *Model) fail(err error) {
	m.status, m.failed = err.Error(), true
}

func (m *Model) invalidate() {
	m.blocks = map[int][]string{}
}

func onOff(label string, on bool) string {
	if !on {
		return label + " off"
	}
	return label + " on"
}

func onOffLabel(b bool) string {
	if b {
		return "On"
	}
	return "Off"
}

func (m *Model) saveConfig() {
	showSide := m.showSidebar
	showTrans := m.showTranslate
	vol := m.player.Level()
	repeatStr := "off"
	switch m.repeat {
	case repeatAyah:
		repeatStr = "ayah"
	case repeatRange:
		repeatStr = "range"
	}
	cfg := config.Config{
		Reciter:         m.reciter.Slug,
		Translation:     m.translation.ID,
		Arabic:          string(m.mode),
		ShowSidebar:     &showSide,
		ShowTranslation: &showTrans,
		Repeat:          repeatStr,
		Volume:          &vol,
	}
	_ = config.Save(cfg)
}

func (m *Model) settingsItems(query string) []pickerItem {
	items := []pickerItem{
		{
			title:  "Reciter: " + m.reciter.Name,
			detail: "enter to change",
			value:  settingReciter,
		},
		{
			title:  "Translation: " + m.translation.Name,
			detail: "enter to change",
			value:  settingTranslation,
		},
		{
			title:  "Arabic Mode: " + string(m.mode),
			detail: "enter to toggle (auto · visual · native)",
			value:  settingArabic,
		},
		{
			title:  "Sidebar: " + onOffLabel(m.showSidebar),
			detail: "enter to toggle",
			value:  settingSidebar,
		},
		{
			title:  "Translation Text: " + onOffLabel(m.showTranslate),
			detail: "enter to toggle",
			value:  settingTranslate,
		},
		{
			title:  "Repeat Mode: " + m.repeat.String(),
			detail: "enter to cycle (off · ayah · range)",
			value:  settingRepeat,
		},
		{
			title:  "Autoplay: " + onOffLabel(m.autoplay),
			detail: "enter to toggle",
			value:  settingAutoplay,
		},
		{
			title:  fmt.Sprintf("Volume: %d%%", m.player.Percent()),
			detail: "left/right or enter to adjust",
			value:  settingVolume,
		},
	}
	if query == "" {
		return items
	}
	var filtered []pickerItem
	q := strings.ToLower(query)
	for _, it := range items {
		if strings.Contains(strings.ToLower(it.title), q) || strings.Contains(strings.ToLower(it.detail), q) {
			filtered = append(filtered, it)
		}
	}
	return filtered
}
