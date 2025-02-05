package commands

import (
	"codeberg.org/jstover/borgdrone/internal/borg"
	"codeberg.org/jstover/borgdrone/internal/config"
	"codeberg.org/jstover/borgdrone/internal/logger"
)

type InitialiseCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

func (cmd InitialiseCmd) Run(cfg config.Config) int {
	targets := cfg.GetTargets(cmd.Target.Archive, cmd.Target.Store)
	Initialise(targets)
	return 0
}

func Initialise(targets []config.Target) {
	logger.Info("Runnning Initialise")
	for _, target := range targets {
		if target.IsInitialised() {
			logger.Warn("%s already initialised", target.GetName())
			continue
		}
		logger.Info("Initialising " + target.GetName())
		target.CreatePasswordFile()
		runner := borg.Runner{Env: target.GetEnvironment()}
		if runner.Run("init", "--encryption", target.Encryption) {
			target.MarkInitialised()
		}
	}
}
