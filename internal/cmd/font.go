package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/4thel00z/quran/internal/font"
)

func newFontCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "font",
		Short: "Manage and install Quranic fonts for terminal readers",
		Long: "Download and install high-quality Quranic fonts (Amiri Quran and Scheherazade New) " +
			"to improve Arabic typography in terminal emulators, and view terminal configuration guidance.",
		Example: "  quran font status\n  quran font install\n  quran font install --force",
	}

	cmd.AddCommand(
		newFontInstallCommand(),
		newFontStatusCommand(),
	)

	return cmd
}

func newFontInstallCommand() *cobra.Command {
	var (
		dirFlag   string
		forceFlag bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Download and install recommended Quran fonts",
		Long:  "Download Amiri Quran and Scheherazade New into the user font directory, update the font cache, and display terminal fallback guidance.",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir, err := font.TargetDir(dirFlag)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Installing fonts to: %s\n\n", targetDir)

			lastFont := ""
			err = font.Install(cmd.Context(), targetDir, forceFlag, func(name string, written, total int64) {
				if name != lastFont {
					if lastFont != "" {
						fmt.Fprintln(out)
					}
					lastFont = name
					fmt.Fprintf(out, "  Downloading %s...", name)
				}
			})
			if err != nil {
				return err
			}
			if lastFont != "" {
				fmt.Fprintln(out, " done.")
			}

			// Report final status
			statuses, err := font.CheckStatus(targetDir)
			if err != nil {
				return err
			}

			fmt.Fprintln(out, "\nFont Status:")
			for _, s := range statuses {
				state := "✓ Installed"
				if !s.Installed {
					state = "✗ Not installed"
				}
				fmt.Fprintf(out, "  %-24s [%s] (%s)\n", s.Font.Name, state, s.Font.Style)
			}

			// Refresh cache
			cacheMsg, cacheErr := font.RefreshCache(targetDir)
			if cacheErr != nil {
				fmt.Fprintf(out, "\nWarning: font cache update: %v\n", cacheErr)
			} else if cacheMsg != "" {
				fmt.Fprintf(out, "\n%s\n", cacheMsg)
			}

			// Print terminal guidance
			fmt.Fprintln(out)
			fmt.Fprint(out, font.TerminalGuide())

			return nil
		},
	}

	cmd.Flags().StringVar(&dirFlag, "dir", "", "custom destination directory for fonts")
	cmd.Flags().BoolVar(&forceFlag, "force", false, "force download even if fonts already exist")

	return cmd
}

func newFontStatusCommand() *cobra.Command {
	var dirFlag string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check font installation status and view terminal configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir, err := font.TargetDir(dirFlag)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			statuses, err := font.CheckStatus(targetDir)
			if err != nil {
				return err
			}

			fmt.Fprintf(out, "Target directory: %s\n\n", targetDir)
			fmt.Fprintln(out, "Quran Fonts:")
			for _, s := range statuses {
				state := "✓ Installed"
				if !s.Installed {
					state = "✗ Not installed"
				}
				fmt.Fprintf(out, "  %-24s [%s] (%s)\n    Path: %s\n    Info: %s\n", s.Font.Name, state, s.Font.Style, s.Path, s.Font.Description)
			}

			fmt.Fprintln(out)
			fmt.Fprint(out, font.TerminalGuide())

			return nil
		},
	}

	cmd.Flags().StringVar(&dirFlag, "dir", "", "custom destination directory for fonts")

	return cmd
}
