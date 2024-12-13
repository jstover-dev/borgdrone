package borgdrone

import (
	"fmt"
	"maps"
	"os"
	"slices"

	"gopkg.in/yaml.v3"
)

type ConfigFile struct {
    Stores struct {

        Filesystem map[string] string

        Ssh map[string] struct {
            Hostname string
            Username string
            Port int
            SshKey string `yaml:"ssh_key"`
        }

    }

    Archives map[string]struct {
        Include []string
        Exclude []string
    }

    Targets []struct {
        Archive string
        Store string
        Encryption string
        Compact bool
        OneFileSytem bool      `yaml:"one_file_system"`
        Prune struct {
            KeepDaily int      `yaml:"keep_daily"`
            KeepWeekly int     `yaml:"keep_weekly"`
            KeepMonthly int    `yaml:"keep_monthly"`
            KeepYearly int     `yaml:"keep_yearly"`
        }
        RcloneUploadPath string `yaml:"rclone_upload_path"`
    }
}

func ReadConfigFile(path string) (ConfigFile, error) {
    var cfg ConfigFile
    data, err := os.ReadFile(path)
    if err != nil {
        return cfg, err
    }
    err = yaml.Unmarshal(data, &cfg)
    if err != nil {
        return cfg, err
    }

    // Validation
    for _, target := range cfg.Targets {

        // Check archive references
        if _, ok := cfg.Archives[target.Archive]; !ok {
            return cfg, fmt.Errorf("ReadConfigFile: Invalid archive reference '%s'", target.Archive)
        }

        // Check store references
        stores := append(
            slices.Collect(maps.Keys(cfg.Stores.Filesystem)),
            slices.Collect(maps.Keys(cfg.Stores.Ssh))...,
        )
        if !slices.Contains(stores, target.Store) {
            return cfg, fmt.Errorf("ReadConfigFile: Invalid store reference '%s'", target.Store)
        }
    }

    return cfg, nil
}
