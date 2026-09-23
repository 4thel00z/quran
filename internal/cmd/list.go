package cmd

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/spf13/cobra"

	"github.com/4thel00z/quran/internal/quran"
)

func newRecitersCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reciters",
		Short: "List reciters",
		Run: func(cmd *cobra.Command, args []string) {
			rows := [][]string{}
			for _, r := range quran.Reciters {
				rows = append(rows, []string{r.Slug, r.Name, r.Style})
			}
			fmt.Fprintln(cmd.OutOrStdout(), listTable([]string{"SLUG", "RECITER", "STYLE"}, rows))
		},
	}
}

func newTranslationsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "translations",
		Short: "List embedded translations",
		Run: func(cmd *cobra.Command, args []string) {
			rows := [][]string{}
			for _, t := range quran.Translations {
				rows = append(rows, []string{t.ID, t.Language, t.Name})
			}
			fmt.Fprintln(cmd.OutOrStdout(), listTable([]string{"ID", "LANGUAGE", "TRANSLATOR"}, rows))
		},
	}
}

func listTable(headers []string, rows [][]string) string {
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E0B45C")).Padding(0, 1)
	cell := lipgloss.NewStyle().Padding(0, 1)
	return table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#3E6B5A"))).
		Headers(headers...).
		Rows(rows...).
		StyleFunc(func(row int, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return header
			}
			return cell
		}).
		String()
}
