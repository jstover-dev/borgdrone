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

	"gopkg.in/yaml.v3"
)

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

	Archives map[string]struct {
		Include []string
		Exclude []string
	}

	Targets []struct {
		Archive       string
		Store         string
		Encryption    string
		Compresion    string
		Compact       bool
		OneFileSystem bool `yaml:"one_file_system"`
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
		Archive:          Archive(cfg.Archives[target.Archive]),
		Encryption:       target.Encryption,
		Compression:      target.Compresion,
		Compact:          target.Compact,
		OneFileSystem:     target.OneFileSystem,
		Prune:            PruneOptions(target.Prune),
		RcloneUploadPath: target.RcloneUploadPath,
	}
	if t.Encryption == "" {
		t.Encryption = "keyfile-blake2"
	}
	if t.Compression == "" {
		t.Compression = "lz4"
	}

	// Populate the appropriate Store and set StoreType
	if store, ok := cfg.Stores.Filesystem[t.StoreName]; ok {
		t.StoreType = LocalStore
		t.Store.Local = store
	} else if store, ok := cfg.Stores.Ssh[t.StoreName]; ok {
		t.StoreType = SSHStore
		t.Store.SSH = &SshStore{
			Hostname: store.Hostname,
			Username: store.Username,
			Port:     store.Port,
			SshKey:   store.SshKey,
		}
		if t.Store.SSH.Port == 0 {
			t.Store.SSH.Port = 22
		}
	}

	// Ensure uninitialised slices are not nil. Workaround for json serialising empty slices as null
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

// GetTargets returns an array of target objects matching the provided target spec
func (cfg Config) GetTargets(archive string, store string) []Target {
	targets := []Target{}

	for _, t := range cfg.TargetMap {
		if t.ArchiveName == archive || archive == "" {
			if t.StoreName == store || store == "" {
				targets = append(targets, t)
			}
		}
	}
	if len(targets) == 0 {
		log.Fatalf("No targets were found matching %s:%s", archive, store)
	}
	return targets
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

		// Check required values are present
		if target.Archive == "" {
			return Config{}, fmt.Errorf("Invalid Configuration: Target missing value: archive")
		}
		if target.Store == "" {
			return Config{}, fmt.Errorf("Invalid Configuration: Target missing value: store")
		}

		// Check archive references
		if _, ok := cfg.Archives[target.Archive]; !ok {
			return Config{}, fmt.Errorf("Invalid configuration: Invalid archive reference '%s' (%s)", target.Archive, path)
		}

		// Check store references
		if !slices.Contains(allStores, target.Store) {
			return Config{}, fmt.Errorf("Invalid configuration: Invalid store reference '%s' (%s)", target.Store, path)
		}
	}

	// Validate SSH Stores
	for name, store := range cfg.Stores.Ssh {
		if store.Hostname == "" {
			return Config{}, fmt.Errorf("Invalid Configuration: SSH Store '%s' missing required value: hostname", name)
		}
	}

	// Generate map of Targets
	targets := make(map[string]Target)
	for idx := range cfg.Targets {
		t := cfg.GetTarget(idx)
		targets[t.GetName()] = t
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

// ConfigPath is a helper function to determine the applications config+data path
// TODO: Separate config from data as per XDG spec
func ConfigPath() string {
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
