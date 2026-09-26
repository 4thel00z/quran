package cmd

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/4thel00z/quran/internal/assets"
	"github.com/4thel00z/quran/internal/audio"
	"github.com/4thel00z/quran/internal/quran"
	"github.com/4thel00z/quran/internal/render"
	"github.com/4thel00z/quran/internal/tui"
)

// session is everything a command needs, resolved from Config.
type session struct {
	book        quran.Book
	reciter     quran.Reciter
	translation quran.Translation
	verses      [][]string
	mode        render.Mode
}

func newSession(cfg Config) (session, error) {
	book, err := assets.Book()
	if err != nil {
		return session{}, err
	}
	reciter, err := cfg.reciter()
	if err != nil {
		return session{}, err
	}
	translation, err := cfg.translation()
	if err != nil {
		return session{}, err
	}
	verses, err := assets.Translation(translation.ID)
	if err != nil {
		return session{}, err
	}
	mode, err := render.ParseMode(cfg.Arabic)
	if err != nil {
		return session{}, err
	}
	return session{book: book, reciter: reciter, translation: translation, verses: verses, mode: mode}, nil
}

// target resolves args like ["juz", "30"]; no args means none.
func (s session) target(args []string) (*quran.Target, error) {
	if len(args) == 0 {
		return nil, nil
	}
	target, err := s.book.Resolve(strings.Join(args, " "))
	if err != nil {
		return nil, err
	}
	return &target, nil
}

func runTUI(ctx context.Context, cfg Config, args []string) error {
	return startTUI(ctx, cfg, args, false)
}

func startTUI(ctx context.Context, cfg Config, args []string, play bool) error {
	s, err := newSession(cfg)
	if err != nil {
		return err
	}
	target, err := s.target(args)
	if err != nil {
		return err
	}
	model, err := tui.New(tui.Options{
		Book:        s.book,
		Reciter:     s.reciter,
		Translation: s.translation,
		Mode:        s.mode,
		Fetcher:     audio.NewFetcher(cfg.BaseURL, cfg.CacheDir),
		Player:      audio.NewPlayer(),
		Target:      target,
		Play:        play && target != nil,
		Config:      cfg.FileConfig,
	})
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(model, tea.WithContext(ctx)).Run()
	if err == tea.ErrProgramKilled {
		return nil
	}
	return err
}
