package commands

import (
	"codeberg.org/jstover/borgdrone/internal/config"
	"codeberg.org/jstover/borgdrone/internal/logger"
)

type ExampleConfigCmd struct{}

func (cmd ExampleConfigCmd) Run(cfg config.Config) int {
	logger.Info(string(config.ExampleConf))
	return 0
}
