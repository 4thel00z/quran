package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/4thel00z/quran/internal/assets"
	"github.com/4thel00z/quran/internal/quran"
)

func newMirrorCommand(cfg *Config) *cobra.Command {
	all := false
	cmd := &cobra.Command{
		Use:   "mirror [reciter...]",
		Short: "Print source URLs and hosted paths for mirroring audio",
		Long: "Prints one line per ayah: the upstream URL and the path under the web root that " +
			"--base-url serves it from. deploy/mirror.sh feeds this to curl on the server.",
		Example: "  quran mirror husary minshawi-murattal\n  quran mirror --all",
		RunE: func(cmd *cobra.Command, args []string) error {
			book, err := assets.Book()
			if err != nil {
				return err
			}
			reciters, err := mirrorReciters(all, args, cfg.Reciter)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, r := range reciters {
				for _, a := range book.Ayahs {
					fmt.Fprintln(out, r.SourceURL(a.Key), r.HostedPath(a.Key))
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "every known reciter")
	return cmd
}

func mirrorReciters(all bool, slugs []string, fallback string) ([]quran.Reciter, error) {
	if all {
		return quran.Reciters, nil
	}
	if len(slugs) == 0 {
		slugs = []string{fallback}
	}
	reciters := []quran.Reciter{}
	for _, slug := range slugs {
		r, err := quran.FindReciter(slug)
		if err != nil {
			return nil, err
		}
		reciters = append(reciters, r)
	}
	return reciters, nil
}
