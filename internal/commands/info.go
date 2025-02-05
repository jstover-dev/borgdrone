package commands

import (
	"codeberg.org/jstover/borgdrone/internal/config"
	"codeberg.org/jstover/borgdrone/internal/logger"
)

type InfoCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

func (cmd InfoCmd) Run(cfg config.Config) int {
	targets := cfg.GetTargets(cmd.Target.Archive, cmd.Target.Store)
	Info(targets)
	return 0
}

func Info(targets []config.Target) {
	for _, target := range targets {
		if !target.IsInitialised() {
			logger.Warn("target '%s' has not been initialised", target.Name())
			continue
		}
		logger.Info("----- %s -----\n", target.Name())
		target.ExecBorg("info")
	}
}
