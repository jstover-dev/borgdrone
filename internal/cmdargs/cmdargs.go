package cmdargs

import (
	"log"
	"os"
	"path"
	"reflect"

	"codeberg.org/jstover/borgdrone/internal/commands"
	"codeberg.org/jstover/borgdrone/internal/config"
	"codeberg.org/jstover/borgdrone/internal/types"

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

type CommonArguments struct {
	ConfigFile string
}

type RunnableCommand interface {
	Run(cfg config.Config) int
}

// list-targets
type ListTargetsCmd struct {
	Format string `arg:"-F,--format" default:"text"`
}

func (cmd ListTargetsCmd) Run(cfg config.Config) int {
	commands.ListTargets(cmd.Format)
	return 0
}

// initialise
type InitialiseCmd struct {
	Target types.BorgTarget `arg:"required,positional"`
}

func (cmd InitialiseCmd) Run(cfg config.Config) int {
	commands.Initialise(cmd.Target)
	return 0
}

// info
type InfoCmd struct {
	Target types.BorgTarget `arg:"required,positional"`
}

func (cmd InfoCmd) Run(cfg config.Config) int {
	commands.Info(cmd.Target)
	return 0
}

// list
type ListCmd struct {
	Target types.BorgTarget `arg:"required,positional"`
}

func (cmd ListCmd) Run(cfg config.Config) int {
	commands.List(cmd.Target)
	return 0
}

// create
type CreateCmd struct {
	Target types.BorgTarget `arg:"required,positional"`
}

func (cmd CreateCmd) Run(cfg config.Config) int {
	commands.Create(cmd.Target)
	return 0
}

// export-key
type ExportKeyCmd struct {
	Target types.BorgTarget `arg:"required,positional"`
}

func (cmd ExportKeyCmd) Run(cfg config.Config) int {
	commands.ExportKey(cmd.Target)
	return 0
}

// import-key
type ImportKeyCmd struct {
	Target       types.BorgTarget `arg:"required,positional"`
	Keyfile      string           `arg:"required"`
	PasswordFile string           `arg:"--password-file"`
}

func (cmd ImportKeyCmd) Run(cfg config.Config) int {
	commands.ImportKey(cmd.Target, cmd.Keyfile, cmd.PasswordFile)
	return 0
}

// clean
type CleanCmd struct{}

func (cmd CleanCmd) Run(cfg config.Config) int {
	commands.Clean()
	return 0
}

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
