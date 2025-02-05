package commands

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"reflect"
	"strings"

	"codeberg.org/jstover/borgdrone/internal/config"
	"github.com/alexflint/go-arg"
)

// RunnableCommand is the interface which all subcommands implement.
// This allows all subcommands to have a .Run() method with a consistent signature.
// Subcommand-specific args are passed into the real command function by their respective implementations
type RunnableCommand interface {
	Run(cfg config.Config) int
}

// BorgTarget holds the ARCHIVE:REPO target specified as CLI argument
type BorgTarget struct {
	Archive string
	Store   string
}

func parseBorgTarget(b []byte) (BorgTarget, error) {
	parts := strings.Split(string(b), ":")
	if len(parts) != 2 {
		return BorgTarget{}, errors.New("does not match [ARCHIVE]:[STORE] format")
	}
	return BorgTarget{Archive: parts[0], Store: parts[1]}, nil
}

// UnmarshalText parses the ARCHIVE:REPO bytestring into BorgTarget fields
func (t *BorgTarget) UnmarshalText(b []byte) error {
	target, err := parseBorgTarget(b)
	if err != nil {
		return err
	}
	*t = target
	return nil
}

// SingleBorgTarget is the same as BorgTarget except fails during Unmarshal if any field is unset.
// It therefore must represent exactly one target
type SingleBorgTarget BorgTarget

func (t *SingleBorgTarget) UnmarshalText(b []byte) error {
	target, err := parseBorgTarget(b)
	if err != nil {
		return err
	}
	if target.Archive == "" || target.Store == "" {
		return errors.New(fmt.Sprintf("'%s:%s' does not match ARCHIVE:STORE format. Empty values are not allowed.", target.Archive, target.Store))
	}
	*t = SingleBorgTarget(target)
	return nil
}

// Arguments struct defines the CLI Interface
type Arguments struct {
	ListTargets *ListTargetsCmd `arg:"subcommand:list-targets"`
	Initialise  *InitialiseCmd  `arg:"subcommand:init"`
	Info        *InfoCmd        `arg:"subcommand:info"`
	List        *ListCmd        `arg:"subcommand:list"`
	Create      *CreateCmd      `arg:"subcommand:create"`
	ExportKey   *ExportKeyCmd   `arg:"subcommand:export-key"`
	ImportKey   *ImportKeyCmd   `arg:"subcommand:import-key"`
	Clean       *CleanCmd       `arg:"subcommand:clean"`

	ConfigFile string `arg:"-c,--config-file"`
}

// RunSubCommand method finds the CLI subcommand specified and calls it's Run() method
func (args *Arguments) RunSubcommand(cfg config.Config) int {
	subCommands := []RunnableCommand{
		args.ListTargets,
		args.Initialise,
		args.Info,
		args.List,
		args.Create,
		args.ExportKey,
		args.ImportKey,
		args.Clean,
	}
	for _, cmd := range subCommands {
		if !reflect.ValueOf(cmd).IsNil() {
			return cmd.Run(cfg)
		}
	}

	return 1
}

// ParseArgs is the main function used to initiate CLI argument parsing
func ParseArgs() *Arguments {
	var args Arguments
	p := arg.MustParse(&args)

	// If --config-file is not provided, set to the default location.
	if args.ConfigFile == "" {
		args.ConfigFile = path.Join(config.ConfigPath(), "borgdrone.yml")
	}

	// Argument Validation
	if args.ImportKey != nil {
		target := args.ImportKey.Target
		if target.Archive == "" || target.Store == "" {
			log.Fatal("Key import ")
		}
		fmt.Println("")
	}

	if p.Subcommand() == nil {
		p.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	return &args
}
