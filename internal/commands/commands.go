package commands

import (
	"encoding/json"
	"log"
	"strings"

	"codeberg.org/jstover/borgdrone/internal/borg"
	"codeberg.org/jstover/borgdrone/internal/config"
	"codeberg.org/jstover/borgdrone/internal/logger"
	"gopkg.in/yaml.v3"
)

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
			logger.Info("Repository  | %s [%s]", target.StoreName, target.GetBorgRepositoryPath())
			logger.Info("")
		}
	}
}

func Initialise(targets []config.Target) {
	logger.Info("Runnning Initialise")
	for _, target := range targets {
		if target.IsInitialised() {
			logger.Warn(target.GetName(), "already initialised")
			continue
		}
		logger.Info("Initialising " + target.GetName())
		target.CreatePasswordFile()
		runner := borg.Runner{Env: target.GetEnvironment()}
		runner.Run("init", "--encryption", target.Encryption)
		target.MarkInitialised()
	}
}

func Info(targets []config.Target) {
	for _, target := range targets {
		if !target.IsInitialised() {
			logger.Warn("target '%s' has not been initialised", target.GetName())
			continue
		}
		logger.Info("----- %s -----\n", target.GetName())
		runner := borg.Runner{Env: target.GetEnvironment()}
		runner.Run("info")
	}
}

func List(targets []config.Target) {
	for _, target := range targets {
		if !target.IsInitialised() {
			logger.Warn("target '%s' has not been initialised", target.GetName())
			continue
		}
		logger.Info("----- %s -----", target.GetName())
		runner := borg.Runner{Env: target.GetEnvironment()}
		runner.Run("list")
	}
}

func Create(targets []config.Target) {
	logger.Info("Running Create")
	for _, target := range targets {
		logger.Info("Target = %+v", target)
	}
}

func ExportKey(targets []config.Target) {
	logger.Info("Running ExportKey")
	for _, target := range targets {
		logger.Info("Target = %+v", target)
	}
}

func ImportKey(target config.Target, keyFile string, passwordFile string) {
	logger.Info("Running ImportKey")
	logger.Info("Target = %+v", target)
	logger.Info("Key File = %s", keyFile)
	logger.Info("Password File = %s", passwordFile)
}

func Clean() {
	logger.Info("Running Clean...")
}
