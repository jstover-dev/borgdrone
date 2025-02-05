package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"codeberg.org/jstover/borgdrone/internal/config"
	"codeberg.org/jstover/borgdrone/internal/logger"
)

type ExportKeyCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

func (cmd ExportKeyCmd) Run(cfg config.Config) int {
	targets := cfg.GetTargets(cmd.Target.Archive, cmd.Target.Store)
	ExportKey(targets)
	return 0
}

func ExportKey(targets []config.Target) {
	passwords := make(map[string]string)
	exported := []string{}

	for _, target := range targets {
		if !target.IsInitialised() {
			logger.Warn("target '%s' has not been initialised", target.Name())
			continue
		}

		key := target.Keyfile()
		pkey := target.PaperKeyfile()

		target.ExecBorg("key", "export", "--paper", "::", pkey)
		logger.Debug("Exported %s", pkey)

		target.ExecBorg("key", "export", "::", pkey)
		logger.Debug("Exported %s", pkey)

		pw, err := os.ReadFile(target.PasswordFile())
		if err != nil {
			logger.Fatal(err.Error(), 3)
		}

		exported = append(exported, key)
		exported = append(exported, pkey)
		passwords[target.Name()] = string(pw)
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
