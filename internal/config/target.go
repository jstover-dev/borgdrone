package config

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"strings"

	"codeberg.org/jstover/borgdrone/internal/borg"
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
// ARCHIVE:STORE format mirrors the CLI format for speciying targets
func (t Target) GetName() string {
	return t.ArchiveName + ":" + t.StoreName
}

// GetConfigPath returns the base path to where all files for this target are stored
// Currently this is $XDG_CONFIG_HOME/borg_drone/<archive>_<store>
func (t Target) GetConfigPath() string {
	configDir := configPath()
	return path.Join(configDir, t.ArchiveName+"_"+t.StoreName)
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
