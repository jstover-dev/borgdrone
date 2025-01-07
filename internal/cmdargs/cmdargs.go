package cmdargs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/alexflint/go-arg"
)

func configPath() string {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		xdgConfigHome = path.Join(userHome, ".config")
	}
	return path.Join(xdgConfigHome, "borgdrone")
}

type BorgTarget struct {
	archive string
	repository string
}
func (t *BorgTarget) UnmarshalText(b []byte) error {
	fmt.Println(b)
	parts := strings.Split(string(b), ":")
	if len(parts) != 2 {
		return errors.New("does not match ARCHIVE:[REPO] format")
	} else {
		t.archive = parts[0]
		t.repository = parts[1]
	}
	return nil
}


type GenerateConfigCmd struct {
	Force bool `arg:"-f,--force"`
}

type ListTargetsCmd struct {
	Format string `arg:"-F,--format" default:"text"`
}

type InitialiseCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

type InfoCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

type ListCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

type CreateCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

type ExportKeysCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

type ImportKeysCmd struct {}

type CleanCmd struct {}

type Arguments struct {
	GenerateConfig *GenerateConfigCmd `arg:"subcommand:generate-config"`
	ListTargets *ListTargetsCmd `arg:"subcommand:list-targets"`
	Initialise *InitialiseCmd `arg:"subcommand:init"`
	Info *InfoCmd `arg:"subcommand:info"`
	List *ListCmd `arg:"subcommand:list"`
	Create *CreateCmd `arg:"subcommand:create"`
	ExportKeys *ExportKeysCmd `arg:"subcommand:export-keys"`
	ImportKeys *ImportKeysCmd `arg:"subcommand:import-keys"`
	Clean *CleanCmd `arg:"subcommand:clean"`

	ConfigFile string `arg:"-c,--config-file"`
}


func ParseArgs() *Arguments {
	var args Arguments
	arg.MustParse(&args)
	if args.ConfigFile == "" {
		args.ConfigFile = path.Join(configPath(), "config.yml")
	}
	return &args
}