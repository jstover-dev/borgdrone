package commands

import (
	"encoding/json"
	"log"
	"strings"

	"codeberg.org/jstover/borgdrone/internal/config"
	"codeberg.org/jstover/borgdrone/internal/logger"
	"gopkg.in/yaml.v3"
)

type ListTargetsCmd struct {
	Format string `arg:"-F,--format" default:"text"`
}

func (cmd ListTargetsCmd) Run(cfg config.Config) int {
	ListTargets(cfg, cmd.Format)
	return 0
}

func ListTargets(cfg config.Config, format string) {
	switch format {
	case "json":
		data, err := json.MarshalIndent(cfg.TargetMap, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		logger.Info(string(data))

	case "yaml":
		data, err := yaml.Marshal(cfg.TargetMap)
		if err != nil {
			log.Fatal(err)
		}
		logger.Info(string(data))

	case "text":
		for name, target := range cfg.TargetMap {
			logger.Info(name)
			logger.Info("Include     | %s", strings.Join(target.Archive.Include, ", "))
			if len(target.Archive.Exclude) > 0 {
				logger.Info("Exclude     | %s", strings.Join(target.Archive.Exclude, ", "))
			}
			logger.Info("Repository  | %s [%s]", target.StoreName, target.BorgRepositoryPath())
			logger.Info("")
		}
	}
}
