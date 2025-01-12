package config

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Stores struct {
		Filesystem map[string]string

		Ssh map[string]struct {
			Hostname string
			Username string
			Port     int
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

//go:embed default.yml
var defaultConfigData []byte

func ReadConfigFile(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return cfg, err
	}

	// Validate Stores
	stores := slices.Concat(
		slices.Collect(maps.Keys(cfg.Stores.Filesystem)),
		slices.Collect(maps.Keys(cfg.Stores.Ssh)),
	)

	storeKeys := make(map[string]struct{})
	for _, name := range stores {
		if _, has := storeKeys[name]; has {
			return cfg, fmt.Errorf("Invalid configuration: Duplicate Store name '%s' (%s)", name, path)
		}
		storeKeys[name] = struct{}{}
	}

	// Validate Targets
	for _, target := range cfg.Targets {

		// Check archive references
		if _, ok := cfg.Archives[target.Archive]; !ok {
			return cfg, fmt.Errorf("Invalid configuration: Invalid archive reference '%s' (%s)", target.Archive, path)
		}

		// Check store references
		stores := append(
			slices.Collect(maps.Keys(cfg.Stores.Filesystem)),
			slices.Collect(maps.Keys(cfg.Stores.Ssh))...,
		)
		if !slices.Contains(stores, target.Store) {
			return cfg, fmt.Errorf("Invalid configuration: Invalid store reference '%s' (%s)", target.Store, path)
		}
	}

	return cfg, nil
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
