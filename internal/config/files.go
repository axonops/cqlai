package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Dir is ~/.cassandra, where cqlai keeps its files beside cqlsh's rather than
// as hidden files in the home directory.
func Dir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cassandra"
	}
	return filepath.Join(home, ".cassandra")
}

// HomeFile is a file directly in the home directory, where an older cqlai
// kept the ones that are now in Dir.
func HomeFile(name string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return name
	}
	return filepath.Join(home, name)
}

// MoveIn readies path for use: it makes the directory path is in if there is
// none, and moves the file at old to path if old is there and path is not.
//
// Moving keeps what the file held. For the MCP token that is the point: the
// client was set up with it, and a new one would stop the client working. A
// file at both places is left as it is, and path is the one used.
func MoveIn(path, old string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("cannot make %s: %w", filepath.Dir(path), err)
	}
	if old == "" || old == path {
		return nil
	}
	if _, err := os.Stat(path); err == nil || !errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if _, err := os.Stat(old); err != nil {
		return nil
	}
	if err := os.Rename(old, path); err != nil {
		return fmt.Errorf("cannot move %s to %s: %w", old, path, err)
	}
	return nil
}
