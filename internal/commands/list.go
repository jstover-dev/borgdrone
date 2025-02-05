package commands

import (
	"codeberg.org/jstover/borgdrone/internal/config"
	"codeberg.org/jstover/borgdrone/internal/logger"
)

type ListCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

func (cmd ListCmd) Run(cfg config.Config) int {
	targets := cfg.GetTargets(cmd.Target.Archive, cmd.Target.Store)
	return List(targets)
}

func List(targets []config.Target) int {
	for _, target := range targets {
		if !target.IsInitialised() {
			logger.Warn("target '%s' has not been initialised", target.Name())
			continue
		}
		logger.Info("----- %s -----", target.Name())
		target.ExecBorg("list")
	}
	return 0
}
