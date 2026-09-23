package main

import (
	"context"
	"os"

	"github.com/4thel00z/quran/internal/cmd"
)

func main() {
	if err := cmd.Execute(context.Background()); err != nil {
		os.Exit(1)
	}
}
