package config

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
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
	Compression      string
	Compact          bool
	OneFileSystem    bool
	Prune            PruneOptions
	RcloneUploadPath string `json:",omitempty" yaml:",omitempty"`
}

func (t Target) SetDefaults() {
	if t.Encryption == "" {
		t.Encryption = "keyfile-blake2"
	}
	if t.Compression == "" {
		t.Compression = "lz4"
	}
	if t.Store.SSH != nil {
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
}

// Name Returns a human-readable label for this target
// ARCHIVE:STORE format mirrors the CLI format for speciying targets
func (t Target) Name() string {
	return t.ArchiveName + ":" + t.StoreName
}

// configPath returns the base path to where all files for this target are stored
// Currently this is $XDG_CONFIG_HOME/borg_drone/<archive>_<store>
func (t Target) configPath() string {
	configDir := ConfigPath()
	return path.Join(configDir, t.ArchiveName+"_"+t.StoreName)
}

// PasswordFile returns the path to the file containing the borg repository password
func (t Target) PasswordFile() string {
	return path.Join(t.configPath(), "passwd")
}

// Keyfile returns the path to the (binary) keyfile
func (t Target) Keyfile() string {
	return path.Join(t.configPath(), "keyfile.bin")
}

// PaperKeyfile returns the path to the "paper" (text) keyfile
func (t Target) PaperKeyfile() string {
	return path.Join(t.configPath(), "keyfile.txt")
}

// IsInitialised will return true if this target has already been initialised (keys/passwords are generated)
func (t Target) IsInitialised() bool {
	_, err := os.Stat(path.Join(t.configPath(), ".initialised"))
	return err == nil
}

// BorgRepositoryPath returns a repo path usable by Borg
// For Local filesystem targets, this is just a path
// For SSH targets, this is an ssh:// URL
func (t Target) BorgRepositoryPath() string {
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

// CreatePasswordFile
func (t Target) CreatePasswordFile() {
	os.MkdirAll(t.configPath(), 0700)
	file, err := os.OpenFile(t.PasswordFile(), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
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

	fmt.Println("Created " + t.PasswordFile())
}

// MarkInitialised
func (t Target) MarkInitialised() {
	_, err := os.Create(path.Join(t.configPath(), ".initialised"))
	if err != nil {
		log.Fatal(err)
	}
}

// Run executes a borg command against the current target repository
func (t Target) ExecBorg(args ...string) bool {
	env := []string{
		"BORG_RELOCATED_REPO_ACCESS_IS_OK=yes",
	}
	env = append(env, fmt.Sprintf("BORG_PASSCOMMAND=cat %s", t.PasswordFile()))
	env = append(env, fmt.Sprintf("BORG_REPO=%s", t.BorgRepositoryPath()))

	if t.StoreType == SSHStore {
		rshOptions := []string{"-o VisualHostKey=no"}
		if t.Store.SSH.SshKey != "" {
			rshOptions = append(rshOptions, "-i "+t.Store.SSH.SshKey)
		}
		env = append(env, fmt.Sprintf("BORG_RSH=ssh %s", strings.Join(rshOptions, " ")))
	}

	return borg.Run(env, args...)
}
