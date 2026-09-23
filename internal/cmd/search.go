package cmd

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

func newSearchCommand(cfg *Config) *cobra.Command {
	limit := 20
	cmd := &cobra.Command{
		Use:   "search <words...>",
		Short: "Search the translation, or the Arabic text when the query is Arabic",
		Long: "Every word must appear in the ayah. Arabic queries ignore diacritics and letter variants " +
			"(أ إ آ ٱ → ا, ى → ي, ة → ه).",
		Example: "  quran search mercy patience\n  quran search -t de-bubenheim Geduld\n  quran search الحي القيوم",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newSession(*cfg)
			if err != nil {
				return err
			}
			hits := s.book.Search(s.verses, strings.Join(args, " "), limit)
			key := lipgloss.NewStyle().Foreground(lipgloss.Color("#4FB08A")).Bold(true)
			name := lipgloss.NewStyle().Foreground(lipgloss.Color("#9A9A94"))
			body := lipgloss.NewStyle().PaddingLeft(2).Width(98)
			out := cmd.OutOrStdout()
			for _, hit := range hits {
				text := s.verses[hit.Key.Surah-1][hit.Key.Ayah-1]
				label := fmt.Sprintf("%-8s", hit.Key)
				fmt.Fprintln(out, key.Render(label)+name.Render(s.book.Surah(hit.Key.Surah).NameEnglish))
				fmt.Fprintln(out, body.Render(text)+"\n")
			}
			if len(hits) == 0 {
				return fmt.Errorf("no ayah contains %q", strings.Join(args, " "))
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", limit, "maximum results")
	return cmd
}
