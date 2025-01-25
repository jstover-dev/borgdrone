package config

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"codeberg.org/jstover/borgdrone/internal/borg"

	"gopkg.in/yaml.v3"
)

type StoreType string

const (
	LocalStore StoreType = "Local"
	SSHStore   StoreType = "SSH"
)

// SshStore contains values needed for remote borg repositories
type SshStore struct {
	Hostname string `json:",omitempty" yaml:",omitempty"`
	Username string `json:",omitempty" yaml:",omitempty"`
	Port     int    `json:",omitempty" yaml:",omitempty"`
	Path     string `json:",omitempty" yaml:",omitempty"`
	SshKey   string `json:",omitempty" yaml:",omitempty"`
}

// Archive contains values needed for locating files to backup
type Archive struct {
	Include []string `json:",omitempty" yaml:",omitempty"`
	Exclude []string `json:",omitempty" yaml:",omitempty"`
}

// PruneOptions contains options for pruning old versions of borg backups
type PruneOptions struct {
	KeepDaily   int `json:",omitempty" yaml:",omitempty"`
	KeepWeekly  int `json:",omitempty" yaml:",omitempty"`
	KeepMonthly int `json:",omitempty" yaml:",omitempty"`
	KeepYearly  int `json:",omitempty" yaml:",omitempty"`
}

// Target is a struct used to hold information about a single borg target
// Store and Archive are read from the separate YAML section and copied here
// This avoids the need to reference into the YAML parser struct
type Target struct {
	StoreName string    `json:"-" yaml:"-"`
	StoreType StoreType `json:"-" yaml:"-"`
	Store     struct {
		Local string    `json:",omitempty" yaml:",omitempty"`
		SSH   *SshStore `json:",omitempty" yaml:",omitempty"`
	}
	ArchiveName      string `json:"-" yaml:"-"`
	Archive          Archive
	Encryption       string
	Compact          bool
	OneFileSytem     bool
	Prune            PruneOptions
	RcloneUploadPath string `json:",omitempty" yaml:",omitempty"`
}

// GetName Returns a human-readable label for this target
// ARCHIVE:STORE mirrors the CLI format for speciying targets
func (t Target) GetName() string {
	return t.ArchiveName + ":" + t.StoreName
}

// GetConfigPath returns the base path to where all files for this target are stored
// Currently this is $XDG_CONFIG_HOME/borg_drone/<archive>_<store>
func (t Target) GetConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	return path.Join(configDir, "borgdrone", t.ArchiveName+"_"+t.StoreName)
}

// GetPasswordFile returns the path to the file containing the borg repository password
func (t Target) GetPasswordFile() string {
	return path.Join(t.GetConfigPath(), "passwd")
}

// GetKeyfile returns the path to the (binary) keyfile
func (t Target) GetKeyfile() string {
	return path.Join(t.GetConfigPath(), "keyfile.bin")
}

// GetPaperKeyfile returns the path to the "paper" (text) keyfile
func (t Target) GetPaperKeyfile() string {
	return path.Join(t.GetConfigPath(), "keyfile.txt")
}

// IsInitialised will return true if this target has already been initialised (keys/passwords are generated)
func (t Target) IsInitialised() bool {
	if _, err := os.Stat(path.Join(t.GetConfigPath(), ".initialised")); err == nil {
		return true
	} else {
		return false
	}
}

// GetBorgRepositoryPath returns a repo path usable by Borg
// For Local filesystem targets, this is just a path
// For SSH targets, this is an ssh:// URL
func (t Target) GetBorgRepositoryPath() string {
	switch t.StoreType {

	case LocalStore:
		return path.Join(t.Store.Local + t.ArchiveName)

	case SSHStore:
		store := t.Store.SSH
		// Use user@host syntax if username was provided
		username := store.Username
		if username != "" {
			username += "@"
		}
		// Ensure Relative paths start with ./
		path := store.Path
		if !strings.HasPrefix(path, ".") {
			if strings.HasPrefix(path, "/") {
				path = strings.TrimLeft(path, "/")
			} else {
				path = "./" + path
			}
		}
		return fmt.Sprintf("ssh://%s%s:%d/%s", username, store.Hostname, store.Port, path)

	default:
		panic("Unknown Store Type: " + t.StoreType)
	}
}

// GetEnvironment returns borg.Environment object used to set the subprocess environment variables
func (t Target) GetEnvironment() borg.Environment {
	env := borg.Environment{
		PassCommand:             "cat " + t.GetPasswordFile(),
		RelocatedRepoAccessIsOk: true,
		Repo:                    t.GetBorgRepositoryPath(),
		Rsh:                     "",
	}
	if t.StoreType == SSHStore {
		rsh := "ssh -o VisualHostKey=no"
		if t.Store.SSH.SshKey != "" {
			rsh += " -i " + t.Store.SSH.SshKey
		}
		env.Rsh = rsh
	}
	return env
}

// ConfigYaml is the struct used for parsing the YAML configuration file
type ConfigYaml struct {
	Stores struct {
		Filesystem map[string]string

		Ssh map[string]struct {
			Hostname string
			Username string
			Port     int
			Path     string
			SshKey   string `yaml:"ssh_key"`
		}
	}

	Archives map[string]Archive

	Targets []struct {
		Archive      string
		Store        string
		Encryption   string
		Compact      bool
		OneFileSytem bool `yaml:"one_file_system"`
		Prune        struct {
			KeepDaily   int `yaml:"keep_daily"`
			KeepWeekly  int `yaml:"keep_weekly"`
			KeepMonthly int `yaml:"keep_monthly"`
			KeepYearly  int `yaml:"keep_yearly"`
		}
		RcloneUploadPath string `yaml:"rclone_upload_path"`
	}
}

// GetTarget reads a target configuration by its positional index and returns a Target object
func (cfg ConfigYaml) GetTarget(idx int) Target {
	target := cfg.Targets[idx]
	t := Target{
		StoreName:        target.Store,
		ArchiveName:      target.Archive,
		Archive:          cfg.Archives[target.Archive],
		Encryption:       target.Encryption,
		Compact:          target.Compact,
		OneFileSytem:     target.OneFileSytem,
		Prune:            PruneOptions(target.Prune),
		RcloneUploadPath: target.RcloneUploadPath,
	}
	// Populate the appropriate Store and set StoreType
	if store, ok := cfg.Stores.Filesystem[target.Store]; ok {
		t.StoreType = LocalStore
		t.Store.Local = store
	} else if store, ok := cfg.Stores.Ssh[target.Store]; ok {
		t.StoreType = SSHStore
		t.Store.SSH = &SshStore{
			Hostname: store.Hostname,
			Username: store.Username,
			Port:     store.Port,
			SshKey:   store.SshKey,
		}
	}
	// Ensure uninitialised slices are not nil. Work around for json serialising as null
	if len(t.Archive.Include) == 0 {
		t.Archive.Include = []string{}
	}
	if len(t.Archive.Exclude) == 0 {
		t.Archive.Exclude = []string{}
	}
	return t
}

// Config is the main configuration struct which is passed into subcommands
// Currently only contains the map of valid targets, but could be used for global program configuration
type Config struct {
	TargetMap map[string]Target
}

//go:embed default.yml
var defaultConfigData []byte

func ReadConfigFile(path string) (Config, error) {
	var cfg ConfigYaml
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, err
	}

	localStores := slices.Collect(maps.Keys(cfg.Stores.Filesystem))
	sshStores := slices.Collect(maps.Keys(cfg.Stores.Ssh))
	allStores := slices.Concat(localStores, sshStores)

	// Validate Stores
	storeKeys := make(map[string]struct{})
	for _, name := range allStores {
		if _, has := storeKeys[name]; has {
			return Config{}, fmt.Errorf("Invalid configuration: Duplicate Store name '%s' (%s)", name, path)
		}
		storeKeys[name] = struct{}{}
	}

	// Validate Targets
	for _, target := range cfg.Targets {

		// Check archive references
		if _, ok := cfg.Archives[target.Archive]; !ok {
			return Config{}, fmt.Errorf("Invalid configuration: Invalid archive reference '%s' (%s)", target.Archive, path)
		}

		// Check store references
		if !slices.Contains(allStores, target.Store) {
			return Config{}, fmt.Errorf("Invalid configuration: Invalid store reference '%s' (%s)", target.Store, path)
		}
	}

	// Generate map of Targets
	targets := make(map[string]Target)
	for idx := range cfg.Targets {
		t := cfg.GetTarget(idx)
		targets[t.GetName()] = t
		fmt.Println(t)
	}

	return Config{TargetMap: targets}, nil
}

func WriteDefaultConfigFile(path string) int {

	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if errors.Is(err, os.ErrExist) {
		return 0
	} else if err != nil {
		log.Fatal(err)
	}
	n, err := file.Write(defaultConfigData)
	if err != nil {
		log.Fatal(err)
	}
	return n
}
