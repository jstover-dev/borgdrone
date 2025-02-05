package commands

import (
	"os"
	"path/filepath"
	"strings"

	"codeberg.org/jstover/borgdrone/internal/config"
	"codeberg.org/jstover/borgdrone/internal/logger"
)

type CreateCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

func (cmd CreateCmd) Run(cfg config.Config) int {
	targets := cfg.GetTargets(cmd.Target.Archive, cmd.Target.Store)
	return Create(targets)
}

func Create(targets []config.Target) int {

	expand := func(path string) string {
		if !strings.HasPrefix(path, "~/") {
			return path
		}
		dirname, _ := os.UserHomeDir()
		return filepath.Join(dirname, path[2:])
	}

	logger.Info("Running Create")
	for _, target := range targets {
		logger.Info("----- %s -----", target.Name())
		argv := []string{"create", "--stats", "--compression", target.Compression}
		if target.OneFileSystem {
			argv = append(argv, "--one-file-system")
		}
		for _, p := range target.Archive.Exclude {
			argv = append(argv, "--exclude")
			argv = append(argv, expand(p))
		}
		argv = append(argv, "::{now}")
		for _, p := range target.Archive.Include {
			argv = append(argv, expand(p))
		}
		logger.Info("%+v", argv)
		target.ExecBorg(argv...)
	}
	return 0
}
