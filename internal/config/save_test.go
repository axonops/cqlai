package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readBack(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path) // #nosec G304 - test file
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	m := map[string]any{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	return m
}

func TestSaveWritesLoadedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cqlai.json")
	if err := os.WriteFile(path, []byte(`{"host":"old","port":9042}`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{Host: "new", Port: 9043, SourcePath: path}
	written, err := cfg.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if written != path {
		t.Fatalf("wrote %s, want %s", written, path)
	}

	got := readBack(t, path)
	if got["host"] != "new" {
		t.Errorf("host is %v, want new", got["host"])
	}
	if got["port"] != float64(9043) {
		t.Errorf("port is %v, want 9043", got["port"])
	}
}

// A key this build does not know about is not the price of editing one it does.
func TestSaveKeepsUnknownKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cqlai.json")
	body := `{"host":"old","somethingNew":42,"ssl":{"enabled":true,"futureOption":"x"}}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{Host: "new", SourcePath: path, SSL: &SSLConfig{Enabled: false, CAPath: "/ca.pem"}}
	if _, err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got := readBack(t, path)
	if got["somethingNew"] != float64(42) {
		t.Errorf("unknown top-level key lost: %v", got)
	}
	ssl, ok := got["ssl"].(map[string]any)
	if !ok {
		t.Fatalf("ssl is %v", got["ssl"])
	}
	if ssl["futureOption"] != "x" {
		t.Errorf("unknown nested key lost: %v", ssl)
	}
	if ssl["enabled"] != false {
		t.Errorf("enabled is %v, want false", ssl["enabled"])
	}
	if ssl["caPath"] != "/ca.pem" {
		t.Errorf("caPath is %v, want /ca.pem", ssl["caPath"])
	}
}

// Emptying a field has to empty it. omitempty leaves the key out of the
// marshalled config, and a merge that only ever copied keys in would keep the
// old value for ever.
func TestSaveClearsEmptiedFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cqlai.json")
	body := `{"host":"h","consistency":"QUORUM","ai":{"provider":"openai","apiKey":"k"}}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{Host: "h", SourcePath: path}
	if _, err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got := readBack(t, path)
	if _, ok := got["consistency"]; ok {
		t.Errorf("consistency survived being cleared: %v", got)
	}
	if _, ok := got["ai"]; ok {
		t.Errorf("ai survived being cleared: %v", got)
	}
}

func TestSaveCreatesFileWhenNoneWasLoaded(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := &Config{Host: "localhost", Port: 9042}
	written, err := cfg.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	want := filepath.Join(home, ".cqlai.json")
	if written != want {
		t.Fatalf("wrote %s, want %s", written, want)
	}
	if cfg.SourcePath != want {
		t.Errorf("SourcePath is %s, want %s", cfg.SourcePath, want)
	}

	info, err := os.Stat(written)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != SaveFileMode {
		t.Errorf("mode is %o, want %o", mode, SaveFileMode)
	}
}

func TestLoadConfigRecordsSourcePath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	path := filepath.Join(dir, ".cqlai.json")
	if err := os.WriteFile(path, []byte(`{"host":"h"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.SourcePath != path {
		t.Errorf("SourcePath is %q, want %q", cfg.SourcePath, path)
	}
}
