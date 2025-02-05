package main

import (
	"codeberg.org/jstover/borgdrone/internal/commands"
	"codeberg.org/jstover/borgdrone/internal/config"
	"codeberg.org/jstover/borgdrone/internal/logger"
)

func main() {
	args := commands.ParseArgs()
	config.WriteDefaultConfigFile(args.ConfigFile)

	cfg, err := config.ReadConfigFile(args.ConfigFile)
	if err != nil {
		logger.Fatal(err.Error(), 1)
	}

	args.RunSubcommand(cfg)

}
