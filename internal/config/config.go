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

// Target is a struct used to hold information about a single borg target
// Store and Archive are read from the separate YAML section and copied here
// This avoids the need to reference into the YAML parser struct
type Target struct {
	StoreName string
	StoreType StoreType
	Store     struct {
		Local string
		SSH   struct {
			Hostname string
			Username string
			Port     int
			Path     string
			SshKey   string
		}
	}

	ArchiveName string
	Archive     struct {
		Include []string
		Exclude []string
	}

	Encryption   string
	Compact      bool
	OneFileSytem bool
	Prune        struct {
		KeepDaily   int
		KeepWeekly  int
		KeepMonthly int
		KeepYearly  int
	}
	RcloneUploadPath string
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

// Config is the struct used for parsing the YAML configuration file
type Config struct {
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

	Archives map[string]struct {
		Include []string
		Exclude []string
	}

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
func (cfg Config) GetTarget(idx int) Target {
	target := cfg.Targets[idx]
	t := Target{
		StoreName:    target.Store,
		ArchiveName:  target.Archive,
		Archive:      cfg.Archives[target.Archive],
		Encryption:   target.Encryption,
		Compact:      target.Compact,
		OneFileSytem: target.OneFileSytem,
		Prune: struct {
			KeepDaily   int
			KeepWeekly  int
			KeepMonthly int
			KeepYearly  int
		}(target.Prune),
		RcloneUploadPath: target.RcloneUploadPath,
	}
	// Populate the appropriate Store and set StoreType
	if store, ok := cfg.Stores.Filesystem[target.Store]; ok {
		t.StoreType = LocalStore
		t.Store.Local = store
	} else if store, ok := cfg.Stores.Ssh[target.Store]; ok {
		t.StoreType = SSHStore
		t.Store.SSH.Hostname = store.Hostname
		t.Store.SSH.Username = store.Username
		t.Store.SSH.Port = store.Port
		t.Store.SSH.SshKey = store.SshKey
	}
	return t
}

//go:embed default.yml
var defaultConfigData []byte

func ReadConfigFile(path string) (map[string]Target, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	localStores := slices.Collect(maps.Keys(cfg.Stores.Filesystem))
	sshStores := slices.Collect(maps.Keys(cfg.Stores.Ssh))
	allStores := slices.Concat(localStores, sshStores)

	// Validate Stores
	storeKeys := make(map[string]struct{})
	for _, name := range allStores {
		if _, has := storeKeys[name]; has {
			return nil, fmt.Errorf("Invalid configuration: Duplicate Store name '%s' (%s)", name, path)
		}
		storeKeys[name] = struct{}{}
	}

	// Validate Targets
	for _, target := range cfg.Targets {

		// Check archive references
		if _, ok := cfg.Archives[target.Archive]; !ok {
			return nil, fmt.Errorf("Invalid configuration: Invalid archive reference '%s' (%s)", target.Archive, path)
		}

		// Check store references
		if !slices.Contains(allStores, target.Store) {
			return nil, fmt.Errorf("Invalid configuration: Invalid store reference '%s' (%s)", target.Store, path)
		}
	}

	// Generate map of Targets
	targets := make(map[string]Target)
	for idx := range cfg.Targets {
		t := cfg.GetTarget(idx)
		targets[t.GetName()] = t
		fmt.Println(t)
	}

	return targets, nil
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
