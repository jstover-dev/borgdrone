package bdTypes

import (
	"errors"
	"strings"
)

type BorgTarget struct {
	Archive    string
	Repository string
}

func (t *BorgTarget) UnmarshalText(b []byte) error {
	parts := strings.Split(string(b), ":")
	if len(parts) != 2 {
		return errors.New("does not match ARCHIVE:[REPO] format")
	} else {
		t.Archive = parts[0]
		t.Repository = parts[1]
	}
	return nil
}
