package cmdargs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"reflect"
	"strings"

	"codeberg.org/jstover/borgdrone/internal/commands"
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

// SingleBorgTarget is the same as BorgTarget except fails during Unmarshal if any field is unset
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

// list-targets
// ----------------------------------------------------------------------------
type ListTargetsCmd struct {
	Format string `arg:"-F,--format" default:"text"`
}

func (cmd ListTargetsCmd) Run(cfg config.Config) int {
	commands.ListTargets(cfg, cmd.Format)
	return 0
}

// initialise
// ----------------------------------------------------------------------------
type InitialiseCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

func (cmd InitialiseCmd) Run(cfg config.Config) int {
	targets := cfg.GetTargets(cmd.Target.Archive, cmd.Target.Store)
	commands.Initialise(targets)
	return 0
}

// info
// ----------------------------------------------------------------------------
type InfoCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

func (cmd InfoCmd) Run(cfg config.Config) int {
	targets := cfg.GetTargets(cmd.Target.Archive, cmd.Target.Store)
	commands.Info(targets)
	return 0
}

// list
// ----------------------------------------------------------------------------
type ListCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

func (cmd ListCmd) Run(cfg config.Config) int {
	targets := cfg.GetTargets(cmd.Target.Archive, cmd.Target.Store)
	commands.List(targets)
	return 0
}

// create
// ----------------------------------------------------------------------------
type CreateCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

func (cmd CreateCmd) Run(cfg config.Config) int {
	targets := cfg.GetTargets(cmd.Target.Archive, cmd.Target.Store)
	commands.Create(targets)
	return 0
}

// export-key
// ----------------------------------------------------------------------------
type ExportKeyCmd struct {
	Target BorgTarget `arg:"required,positional"`
}

func (cmd ExportKeyCmd) Run(cfg config.Config) int {
	targets := cfg.GetTargets(cmd.Target.Archive, cmd.Target.Store)
	commands.ExportKey(targets)
	return 0
}

// import-key
// ----------------------------------------------------------------------------
type ImportKeyCmd struct {
	Target       SingleBorgTarget `arg:"required,positional"`
	Keyfile      string           `arg:"required"`
	PasswordFile string           `arg:"--password-file"`
}

func (cmd ImportKeyCmd) Run(cfg config.Config) int {
	target := cfg.GetTargets(cmd.Target.Archive, cmd.Target.Store)[0]
	commands.ImportKey(target, cmd.Keyfile, cmd.PasswordFile)
	return 0
}

// clean
// ----------------------------------------------------------------------------
type CleanCmd struct{}

func (cmd CleanCmd) Run(cfg config.Config) int {
	targets := cfg.GetTargets("", "")
	commands.Clean(targets)
	return 0
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
