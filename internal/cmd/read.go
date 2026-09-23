package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"github.com/4thel00z/quran/internal/quran"
	"github.com/4thel00z/quran/internal/render"
)

func newReadCommand(cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:     "read <surah | surah:ayah | surah:from-to | juz N | hizb N | page N>",
		Short:   "Print ayahs with their translation",
		Example: "  quran read 1\n  quran read 2:255 -t de-bubenheim\n  quran read page 604 --arabic native",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newSession(*cfg)
			if err != nil {
				return err
			}
			target, err := s.target(args)
			if err != nil {
				return err
			}
			printer := newPrinter(s, cmd.OutOrStdout())
			for _, ayah := range s.book.TargetAyahs(*target) {
				printer.ayah(ayah)
			}
			return nil
		},
	}
}

type printStyles struct {
	title       lipgloss.Style
	key         lipgloss.Style
	arabic      lipgloss.Style
	marker      lipgloss.Style
	translation lipgloss.Style
	surah       lipgloss.Style
}

// printer writes ayahs to a plain stream, used by read and play --no-tui.
type printer struct {
	session
	out    io.Writer
	width  int
	styles printStyles
	surah  int
}

func newPrinter(s session, out io.Writer) *printer {
	width := 80
	if w, _, err := term.GetSize(os.Stdout.Fd()); err == nil && w > 20 {
		width = min(w, 100)
	}
	return &printer{
		session: s,
		out:     out,
		width:   width,
		styles: printStyles{
			title:       lipgloss.NewStyle().Foreground(lipgloss.Color("#E7BE62")).Bold(true),
			key:         lipgloss.NewStyle().Foreground(lipgloss.Color("#4FB08A")),
			arabic:      lipgloss.NewStyle(),
			marker:      lipgloss.NewStyle().Foreground(lipgloss.Color("#E7BE62")),
			translation: lipgloss.NewStyle().Foreground(lipgloss.Color("#9A9A94")),
			surah:       lipgloss.NewStyle().Foreground(lipgloss.Color("#E7BE62")).Bold(true).MarginTop(1),
		},
	}
}

func (p *printer) ayah(ayah quran.Ayah) {
	st := p.styles
	if ayah.Surah != p.surah {
		p.surah = ayah.Surah
		surah := p.book.Surah(ayah.Surah)
		fmt.Fprintln(p.out, st.surah.Render(fmt.Sprintf("%d. %s · %s", surah.Number, surah.NameEnglish, surah.NameMeaning)))
	}
	words := append(append([]string{}, ayah.Words...), render.AyahMarker(ayah.Ayah))
	marker := len(words) - 1
	style := func(i int) lipgloss.Style {
		if i == marker {
			return st.marker
		}
		return st.arabic
	}
	lines := []string{st.key.Render(ayah.Key.String())}
	lines = append(lines, render.RTLLines(words, p.width, p.mode, style, lipgloss.NewStyle())...)
	lines = append(lines, render.TextLines(p.verses[ayah.Surah-1][ayah.Ayah-1], p.width, p.mode, st.translation)...)
	fmt.Fprintln(p.out, strings.Join(lines, "\n")+"\n")
}
