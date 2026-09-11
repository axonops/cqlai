package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Writing configuration back to disk.
//
// LoadConfig reads a cqlshrc and then one of three JSON files, and until the
// PREFERENCES window there was nothing going the other way: every setting was
// changed by leaving cqlai and editing the file by hand.
//
// The file is not overwritten with a fresh marshal of Config. A file can hold
// keys this build does not know - written by a newer cqlai, or by hand for
// something not implemented yet - and rewriting it wholesale would drop them
// silently, as the cost of changing one unrelated field. What Save does is read
// the file back, replace the keys Config describes, and leave the rest alone.

// SaveFileMode is what the config file is written as. It holds a password and
// API keys.
const SaveFileMode = 0o600

// SavePath is the file Save writes to: the one this config was loaded from, or
// ~/.cqlai.json when nothing was found.
//
// The first candidate LoadConfig looks at is a relative "cqlai.json", which is
// whatever directory cqlai happened to start in. That is fine to read and a
// poor place to create, so a config that came from nowhere goes to the home
// directory rather than to the working one.
func (c *Config) SavePath() string {
	if c.SourcePath != "" {
		return c.SourcePath
	}
	return filepath.Join(os.Getenv("HOME"), ".cqlai.json")
}

// Save writes the configuration, and reports the file it wrote.
func (c *Config) Save() (string, error) {
	path := c.SavePath()

	existing := map[string]any{}
	if data, err := os.ReadFile(path); err == nil { // #nosec G304 - the path is the config file this config was loaded from
		if err := json.Unmarshal(data, &existing); err != nil {
			return "", fmt.Errorf("cannot read %s back: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("cannot read %s: %w", path, err)
	}

	updated, err := asMap(c)
	if err != nil {
		return "", err
	}
	mergeKnown(existing, updated, reflect.TypeOf(*c))

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')

	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return "", fmt.Errorf("cannot create %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, data, SaveFileMode); err != nil {
		return "", fmt.Errorf("cannot write %s: %w", path, err)
	}

	c.SourcePath = path
	return path, nil
}

// asMap marshals a value and reads it back as a map, so the merge below works
// on the same shape the file has.
func asMap(v any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	m := map[string]any{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// mergeKnown copies src over dst for every key the struct type t describes, and
// leaves every other key in dst untouched.
//
// A key t knows about but src does not have was emptied - omitempty leaves it
// out - and is removed from dst rather than kept. Editing a field to nothing
// has to mean nothing, or a value could never be cleared from the window.
func mergeKnown(dst, src map[string]any, t reflect.Type) {
	for i := range t.NumField() {
		field := t.Field(i)
		name := jsonName(field)
		if name == "" {
			continue
		}

		value, ok := src[name]
		if !ok {
			delete(dst, name)
			continue
		}

		// A nested object - ssl, ai, and the per-provider blocks under it -
		// gets the same treatment one level down, so a key inside one of those
		// that cqlai does not know about survives too.
		if nested := structType(field.Type); nested != nil {
			from, fromOK := value.(map[string]any)
			into, intoOK := dst[name].(map[string]any)
			if fromOK && intoOK {
				mergeKnown(into, from, nested)
				continue
			}
		}

		dst[name] = value
	}
}

// jsonName is the key a field is written under, or "" for one that is not
// written at all.
func jsonName(field reflect.StructField) string {
	if field.PkgPath != "" {
		return "" // unexported
	}
	tag := field.Tag.Get("json")
	name, _, _ := strings.Cut(tag, ",")
	switch name {
	case "-":
		return ""
	case "":
		return field.Name
	}
	return name
}

// structType returns the struct a field is, or points at, and nil for anything
// else.
func structType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() == reflect.Struct {
		return t
	}
	return nil
}
