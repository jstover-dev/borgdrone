package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

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

	expand := func(path string) string {
		if !strings.HasPrefix(path, "~/") {
			return path
		}
		dirname, _ := os.UserHomeDir()
		return filepath.Join(dirname, path[2:])
	}

	logger.Info("Running Create")
	for _, target := range targets {
		logger.Info("----- %s -----", target.GetName())
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
		runner := borg.Runner{Env: target.GetEnvironment()}
		runner.Run(argv...)
	}
}

func ExportKey(targets []config.Target) {
	passwords := make(map[string]string)
	exported := []string{}

	for _, target := range targets {
		if !target.IsInitialised() {
			logger.Warn("target '%s' has not been initialised", target.GetName())
			continue
		}

		key := target.GetKeyfile()
		pkey := target.GetPaperKeyfile()

		runner := borg.Runner{Env: target.GetEnvironment()}
		runner.Run("key", "export", "--paper", "::", pkey)
		logger.Debug("Exported %s", pkey)

		runner.Run("key", "export", "::", pkey)
		logger.Debug("Exported %s", pkey)

		pw, err := os.ReadFile(target.GetPasswordFile())
		if err != nil {
			logger.Fatal(err.Error(), 3)
		}

		exported = append(exported, key)
		exported = append(exported, pkey)
		passwords[target.GetName()] = string(pw)
	}

	if len(passwords) > 0 {
		logger.Warn("Repository passwords. You should back up these values to a safe location:")
		w := tabwriter.NewWriter(logger.NewWriter(logger.LevelInfo), 1, 4, 4, ' ', 0)
		for repo, pw := range passwords {
			fmt.Fprintf(w, "\t%s\t%s\n", repo, pw)
		}
		w.Flush()
	}
	logger.Info("")

	if len(exported) > 0 {
		logger.Warn("MAKE SURE TO BACKUP THESE FILES, AND THEN REMOVE FROM THE LOCAL FILESYSTEM!")
		logger.Warn("You can delete these files by running: `borgdrone clean")
		for _, f := range exported {
			logger.Info("\t%s", f)
		}
	}
	logger.Info("")
}

func ImportKey(target config.Target, keyFile string, passwordFile string) {
	logger.Info("Running ImportKey")
	logger.Info("Target = %+v", target)
	logger.Info("Key File = %s", keyFile)
	logger.Info("Password File = %s", passwordFile)
}

func Clean(targets []config.Target) {
	keys := []string{}
	for _, t := range targets {
		keys = append(keys, t.GetKeyfile())
		keys = append(keys, t.GetPaperKeyfile())
	}
	var n = 0
	for _, k := range keys {
		err := os.Remove(k)
		if !errors.Is(err, fs.ErrNotExist) {
			n++
			logger.Info("Removed %s", k)
		}
	}
	logger.Info("%d files removed", n)
}
