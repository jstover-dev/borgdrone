package config

import (
	"fmt"
	"testing"

	"codeberg.org/jstover/borgdrone/internal/borg"
)

func mockRun(env []string, args ...string) bool {
	fmt.Println("was called!")
	fmt.Println(env)
	fmt.Println(args)
	return true
}


var testTarget = Target{
	StoreName:        "testStore",
	ArchiveName:      "testArchive",
	Archive:          strucArchive(cfg.Archives[target.Archive]),
	Encryption:       target.Encryption,
	Compression:      target.Compresion,
	Compact:          target.Compact,
	OneFileSystem:    target.OneFileSystem,
	Prune:            PruneOptions(target.Prune),
	RcloneUploadPath: target.RcloneUploadPath,
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
}
t.SetDefaults()
return t

func TestBorgRun(t *testing.T) {
	borgRunner = borg.Runner{Run: mockRun}
	target := Target{}
	target.ExecBorg("hello", "world")
}
