package commands

import (
	"errors"
	"io/fs"
	"os"

	"codeberg.org/jstover/borgdrone/internal/config"
	"codeberg.org/jstover/borgdrone/internal/logger"
)

type CleanCmd struct{}

func (cmd CleanCmd) Run(cfg config.Config) int {
	targets := cfg.GetTargets("", "")
	Clean(targets)
	return 0
}

func Clean(targets []config.Target) {
	keys := []string{}
	for _, t := range targets {
		keys = append(keys, t.GetKeyfile())
		keys = append(keys, t.GetPaperKeyfile())
	}
	var n = 0
	for _, k := range keys {
		err := os.Remove(k)
		if !errors.Is(err, fs.ErrNotExist) {
			n++
			logger.Info("Removed %s", k)
		}
	}
	logger.Info("%d files removed", n)
}
