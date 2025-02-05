package config

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
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
	Compression      string
	Compact          bool
	OneFileSystem     bool
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
	configDir := ConfigPath()
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
		return path.Join(t.Store.Local, t.ArchiveName)

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

// GetEnvironment
func (t Target) GetEnvironment() []string {
	e := []string{
		"BORG_RELOCATED_REPO_ACCESS_IS_OK=yes",
	}
	e = append(e, fmt.Sprintf("BORG_PASSCOMMAND=cat %s", t.GetPasswordFile()))
	e = append(e, fmt.Sprintf("BORG_REPO=%s", t.GetBorgRepositoryPath()))

	if t.StoreType == SSHStore {
		rshOptions := []string{"-o VisualHostKey=no"}
		if t.Store.SSH.SshKey != "" {
			rshOptions = append(rshOptions, "-i "+t.Store.SSH.SshKey)
		}
		e = append(e, fmt.Sprintf("BORG_RSH=ssh %s", strings.Join(rshOptions, " ")))
	}
	return e
}

// CreatePasswordFile
func (t Target) CreatePasswordFile() {
	os.MkdirAll(t.GetConfigPath(), 0700)
	file, err := os.OpenFile(t.GetPasswordFile(), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return
	} else if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.WriteString("SECRETPASSWORD")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Created " + t.GetPasswordFile())
}

// MarkInitialised
func (t Target) MarkInitialised() {
	_, err := os.Create(path.Join(t.GetConfigPath(), ".initialised"))
	if err != nil {
		log.Fatal(err)
	}
}
