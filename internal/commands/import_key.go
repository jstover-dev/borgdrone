package commands

import (
	"codeberg.org/jstover/borgdrone/internal/config"
	"codeberg.org/jstover/borgdrone/internal/logger"
)

type ImportKeyCmd struct {
	Target       SingleBorgTarget `arg:"required,positional"`
	Keyfile      string           `arg:"required"`
	PasswordFile string           `arg:"--password-file"`
}

func (cmd ImportKeyCmd) Run(cfg config.Config) int {
	target := cfg.GetTargets(cmd.Target.Archive, cmd.Target.Store)[0]
	ImportKey(target, cmd.Keyfile, cmd.PasswordFile)
	return 0
}

func ImportKey(target config.Target, keyFile string, passwordFile string) {
	logger.Info("Running ImportKey")
	logger.Info("Target = %+v", target)
	logger.Info("Key File = %s", keyFile)
	logger.Info("Password File = %s", passwordFile)
}
