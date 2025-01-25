package cmdargs

import (
	"log"
	"os"
	"path"
	"reflect"

	"codeberg.org/jstover/borgdrone/internal/bdTypes"
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
	Target bdTypes.BorgTarget `arg:"required,positional"`
}

func (cmd InitialiseCmd) Run(cfg config.Config) int {
	commands.Initialise(cfg, cmd.Target)
	return 0
}

// info
// ----------------------------------------------------------------------------
type InfoCmd struct {
	Target bdTypes.BorgTarget `arg:"required,positional"`
}

func (cmd InfoCmd) Run(cfg config.Config) int {
	commands.Info(cmd.Target)
	return 0
}

// list
// ----------------------------------------------------------------------------
type ListCmd struct {
	Target bdTypes.BorgTarget `arg:"required,positional"`
}

func (cmd ListCmd) Run(cfg config.Config) int {
	commands.List(cmd.Target)
	return 0
}

// create
// ----------------------------------------------------------------------------
type CreateCmd struct {
	Target bdTypes.BorgTarget `arg:"required,positional"`
}

func (cmd CreateCmd) Run(cfg config.Config) int {
	commands.Create(cmd.Target)
	return 0
}

// export-key
// ----------------------------------------------------------------------------
type ExportKeyCmd struct {
	Target bdTypes.BorgTarget `arg:"required,positional"`
}

func (cmd ExportKeyCmd) Run(cfg config.Config) int {
	commands.ExportKey(cmd.Target)
	return 0
}

// import-key
// ----------------------------------------------------------------------------
type ImportKeyCmd struct {
	Target       bdTypes.BorgTarget `arg:"required,positional"`
	Keyfile      string             `arg:"required"`
	PasswordFile string             `arg:"--password-file"`
}

func (cmd ImportKeyCmd) Run(cfg config.Config) int {
	commands.ImportKey(cmd.Target, cmd.Keyfile, cmd.PasswordFile)
	return 0
}

// clean
// ----------------------------------------------------------------------------
type CleanCmd struct{}

func (cmd CleanCmd) Run(cfg config.Config) int {
	commands.Clean()
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
	arg.MustParse(&args)

	// If --config-file is not provided, set to the default location.
	if args.ConfigFile == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			log.Fatal(err)
		}
		args.ConfigFile = path.Join(configDir, "borgdrone", "config.yml")
	}

	return &args
}
