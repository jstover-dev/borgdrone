package types

import (
	"strings"
	"errors"
)

type BorgTarget struct {
	archive string
	repository string
}

func (t *BorgTarget) UnmarshalText(b []byte) error {
	parts := strings.Split(string(b), ":")
	if len(parts) != 2 {
		return errors.New("does not match ARCHIVE:[REPO] format")
	} else {
		t.archive = parts[0]
		t.repository = parts[1]
	}
	return nil
}


