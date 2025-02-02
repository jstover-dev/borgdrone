package main

import (
	"log"

	"codeberg.org/jstover/borgdrone/internal/cmdargs"
	"codeberg.org/jstover/borgdrone/internal/config"
)

func main() {
	args := cmdargs.ParseArgs()
	config.WriteDefaultConfigFile(args.ConfigFile)

	cfg, err := config.ReadConfigFile(args.ConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	args.RunSubcommand(cfg)

}
