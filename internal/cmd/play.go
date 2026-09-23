package cmd

import (
	"context"
	"fmt"
	"io"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/4thel00z/quran/internal/audio"
	"github.com/4thel00z/quran/internal/quran"
)

func newPlayCommand(cfg *Config) *cobra.Command {
	headless := false
	cmd := &cobra.Command{
		Use:   "play <surah | surah:ayah | surah:from-to | juz N | hizb N | page N>",
		Short: "Recite a surah, ayah, range, juz, hizb or page",
		Example: "  quran play 36\n  quran play 2:255 -r husary\n  quran play juz 30 --no-tui\n" +
			"  quran play 18:1-10 -t de-bubenheim",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !headless {
				return startTUI(cmd.Context(), *cfg, args, true)
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()
			return playHeadless(ctx, *cfg, args, cmd.OutOrStdout())
		},
	}
	cmd.Flags().BoolVar(&headless, "no-tui", false, "print each ayah as it is recited instead of opening the reader")
	return cmd
}

func playHeadless(ctx context.Context, cfg Config, args []string, out io.Writer) error {
	s, err := newSession(cfg)
	if err != nil {
		return err
	}
	target, err := s.target(args)
	if err != nil {
		return err
	}
	fetcher := audio.NewFetcher(cfg.BaseURL, cfg.CacheDir)
	player := audio.NewPlayer()
	defer player.Stop()
	ayahs := s.book.TargetAyahs(*target)
	printer := newPrinter(s, out)
	fmt.Fprintln(out, printer.styles.title.Render(s.reciter.Title()))
	for i, ayah := range ayahs {
		data, err := fetcher.Fetch(ctx, s.reciter, ayah.Key)
		if err != nil {
			return err
		}
		done, err := player.Play(data)
		if err != nil {
			return err
		}
		if i+1 < len(ayahs) {
			go prefetch(ctx, fetcher, s.reciter, ayahs[i+1].Key)
		}
		printer.ayah(ayah)
		select {
		case <-done:
		case <-ctx.Done():
			return nil
		}
	}
	return nil
}

func prefetch(ctx context.Context, fetcher *audio.Fetcher, reciter quran.Reciter, key quran.Key) {
	_, _ = fetcher.Fetch(ctx, reciter, key)
}
