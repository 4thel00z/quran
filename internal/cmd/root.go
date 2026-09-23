// Package cmd wires the quran command line.
package cmd

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"github.com/4thel00z/quran/internal/quran"
	"github.com/4thel00z/quran/internal/render"
)

var version = "dev"

const (
	defaultBaseURL     = "https://quran.host"
	defaultReciter     = "minshawi-murattal"
	defaultTranslation = "en-sahih"
)

// Config is shared by every subcommand; flags override the QURAN_* env vars.
type Config struct {
	BaseURL     string
	Reciter     string
	Translation string
	CacheDir    string
	Arabic      string
}

func (c Config) reciter() (quran.Reciter, error) {
	return quran.FindReciter(c.Reciter)
}

func (c Config) translation() (quran.Translation, error) {
	return quran.FindTranslation(c.Translation)
}

func envOr(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func defaultCacheDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return dir + "/quran"
}

func newRoot() *cobra.Command {
	cfg := &Config{}
	root := &cobra.Command{
		Use:   "quran [surah | surah:ayah | juz N | page N]",
		Short: "Read and listen to the Quran in your terminal",
		Long: "Read the Quran with translations and listen ayah by ayah, with the recited word highlighted.\n" +
			"Text and timings come from Tarteel's quran-assets; audio is streamed from " + defaultBaseURL + ".",
		Example: "  quran\n  quran 18\n  quran 2:255 --reciter husary\n  quran juz 30 --translation de-bubenheim",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTUI(cmd.Context(), *cfg, args)
		},
	}
	flags := root.PersistentFlags()
	flags.StringVar(&cfg.BaseURL, "base-url", envOr("QURAN_BASE_URL", defaultBaseURL), "audio host ($QURAN_BASE_URL)")
	flags.StringVarP(&cfg.Reciter, "reciter", "r", envOr("QURAN_RECITER", defaultReciter), "reciter slug ($QURAN_RECITER)")
	flags.StringVarP(&cfg.Translation, "translation", "t", envOr("QURAN_TRANSLATION", defaultTranslation), "translation id ($QURAN_TRANSLATION)")
	flags.StringVar(&cfg.Arabic, "arabic", envOr("QURAN_ARABIC", string(render.Auto)), "visual, native or auto ($QURAN_ARABIC)")
	flags.StringVar(&cfg.CacheDir, "cache-dir", envOr("QURAN_CACHE_DIR", defaultCacheDir()), "ayah cache, empty disables ($QURAN_CACHE_DIR)")
	root.AddCommand(
		newPlayCommand(cfg),
		newReadCommand(cfg),
		newSearchCommand(cfg),
		newRecitersCommand(),
		newTranslationsCommand(),
		newMirrorCommand(cfg),
	)
	return root
}

func Execute(ctx context.Context) error {
	return fang.Execute(ctx, newRoot(), fang.WithVersion(version), fang.WithColorSchemeFunc(colorScheme))
}
