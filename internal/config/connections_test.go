package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAConnectionIsNamedAfterItsHostWhenItIsNotNamed.
//
// A name is asked for and rarely wanted - one cluster per host is the ordinary
// case - so an empty one falls back to the host rather than refusing to save.
func TestAConnectionIsNamedAfterItsHostWhenItIsNotNamed(t *testing.T) {
	assert.Equal(t, "production", ConnectionName(Config{Name: "production", Host: "10.0.0.1"}))
	assert.Equal(t, "10.0.0.1", ConnectionName(Config{Host: "10.0.0.1"}))
	assert.Equal(t, "10.0.0.1", ConnectionName(Config{Name: "  ", Host: "10.0.0.1"}))
	assert.Empty(t, ConnectionName(Config{}))
}

// TestAConnectionKeepsTheConnectionSettingsAndNothingElse.
//
// They are kept in the same file as everything else, and writing the whole
// configuration under each one would put a copy of the API keys and the output
// format beside every host.
func TestAConnectionKeepsTheConnectionSettingsAndNothingElse(t *testing.T) {
	kept := ConnectionSettings(Config{
		Host: "10.0.0.1", Port: 9042, Keyspace: "shop", Username: "ada", Password: "secret",
		ConnectTimeout: 5, RequestTimeout: 10,
		SSL: &SSLConfig{Enabled: true, CAPath: "/etc/ca.pem"},

		PageSize:     100,
		OutputFormat: "JSON",
		AI:           &AIConfig{APIKey: "not this"},
	})

	assert.Equal(t, "10.0.0.1", kept.Host)
	assert.Equal(t, "shop", kept.Keyspace)
	assert.Equal(t, "/etc/ca.pem", kept.SSL.CAPath)

	assert.Zero(t, kept.PageSize)
	assert.Empty(t, kept.OutputFormat)
	assert.Nil(t, kept.AI)
}

// TestSavingAConnectionTwiceReplacesIt rather than keeping two of the name.
func TestSavingAConnectionTwiceReplacesIt(t *testing.T) {
	cfg := &Config{}

	assert.Equal(t, "production", cfg.SaveConnection(Config{Name: "production", Host: "10.0.0.1"}))
	assert.Equal(t, "staging", cfg.SaveConnection(Config{Name: "staging", Host: "10.0.1.1"}))
	assert.Equal(t, "production", cfg.SaveConnection(Config{Name: "production", Host: "10.0.0.9"}))

	require.Len(t, cfg.Connections, 2)
	assert.Equal(t, "10.0.0.9", cfg.Connections[0].Host)

	production, found := cfg.Connection("production")
	require.True(t, found)
	assert.Equal(t, "10.0.0.9", production.Host)

	_, found = cfg.Connection("nowhere")
	assert.False(t, found)
}

// TestUsingAConnectionMakesItTheOneCqlaiStartsWith.
func TestUsingAConnectionMakesItTheOneCqlaiStartsWith(t *testing.T) {
	cfg := &Config{
		Host: "old", Port: 1, Username: "someone", PageSize: 500,
		SSL: &SSLConfig{Enabled: true},
	}

	cfg.UseConnection(Config{Name: "production", Host: "10.0.0.1", Port: 9042})

	assert.Equal(t, "production", cfg.Name)
	assert.Equal(t, "10.0.0.1", cfg.Host)
	assert.Equal(t, 9042, cfg.Port)
	assert.Empty(t, cfg.Username, "the connection has none, so neither has the shell")
	assert.Nil(t, cfg.SSL, "the connection is not over TLS")
	assert.Equal(t, 500, cfg.PageSize, "everything that is not the connection is left alone")
}

// TestConnectionsAreWrittenToTheFileAndReadBack.
func TestConnectionsAreWrittenToTheFileAndReadBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cqlai.json")
	cfg := &Config{SourcePath: path, Host: "10.0.0.1", PageSize: 100}
	cfg.SaveConnection(Config{Name: "production", Host: "10.0.0.1", Port: 9042})
	cfg.SaveConnection(Config{Host: "10.0.1.1"})

	written, err := cfg.Save()
	require.NoError(t, err)
	assert.Equal(t, path, written)

	data, err := os.ReadFile(path) //nolint:gosec // the test wrote it
	require.NoError(t, err)
	assert.Contains(t, string(data), `"connections"`)

	read := &Config{}
	require.NoError(t, json.Unmarshal(data, read))
	require.Len(t, read.Connections, 2)
	assert.Equal(t, "production", read.Connections[0].Name)
	assert.Equal(t, 9042, read.Connections[0].Port)
	assert.Equal(t, "10.0.1.1", read.Connections[1].Name, "named after its host")

	// A connection holds no connections of its own.
	assert.NotContains(t, string(data), `"connections":[{"name":"production","connections"`)
}

// TestTheFileConnectsToTheFirstOfItsConnections.
//
// The settings at the top level and the list below it are the same thing said
// twice. The list decides: the top level is written to keep older builds - and
// anything reading the file by hand - working.
func TestTheFileConnectsToTheFirstOfItsConnections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cqlai.json")
	require.NoError(t, os.WriteFile(path, []byte(`{
		"host": "left-over", "port": 1, "pageSize": 100,
		"connections": [
			{"name": "production", "host": "10.0.0.1", "port": 9042, "keyspace": "shop"},
			{"name": "staging", "host": "10.0.1.1"}
		]
	}`), 0o600))

	cfg, err := LoadConfig(path)
	require.NoError(t, err)

	assert.Equal(t, "10.0.0.1", cfg.Host, "the first connection is the default")
	assert.Equal(t, 9042, cfg.Port)
	assert.Equal(t, "shop", cfg.Keyspace)
	assert.Equal(t, 100, cfg.PageSize, "and everything else is the file's")
}

// TestAFileWithNoConnectionsStillConnects, which is every file written before
// there was a list of them.
func TestAFileWithNoConnectionsStillConnects(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cqlai.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"host": "10.0.0.1", "port": 9042}`), 0o600))

	cfg, err := LoadConfig(path)
	require.NoError(t, err)

	assert.Equal(t, "10.0.0.1", cfg.Host)
	assert.Equal(t, 9042, cfg.Port)
}

// TestMakingAConnectionTheDefaultPutsItFirst.
func TestMakingAConnectionTheDefaultPutsItFirst(t *testing.T) {
	cfg := &Config{Connections: []Config{
		{Name: "production", Host: "10.0.0.1"},
		{Name: "staging", Host: "10.0.1.1"},
	}}

	assert.Equal(t, "staging", cfg.MakeDefault(Config{Name: "staging", Host: "10.0.1.1", Keyspace: "trial"}))

	require.Len(t, cfg.Connections, 2)
	assert.Equal(t, "staging", cfg.Connections[0].Name)
	assert.Equal(t, "trial", cfg.Connections[0].Keyspace, "with what was changed about it")
	assert.Equal(t, "production", cfg.Connections[1].Name)

	assert.Equal(t, "10.0.1.1", cfg.Host, "and it is what the file connects to")
}

// TestTheConfigurationLivesBesideCqlshrc.
//
// cqlai reads a cqlshrc out of ~/.cassandra already, and its own file used to
// go somewhere else: one directory to find rather than two. The places it used
// to go are still read, so upgrading changes nothing.
func TestTheConfigurationLivesBesideCqlshrc(t *testing.T) {
	t.Setenv("HOME", "/home/someone")

	assert.Equal(t, "/home/someone/.cassandra/cqlai.json", DefaultConfigPath())
	assert.Equal(t, []string{
		"cqlai.json",
		"/home/someone/.cassandra/cqlai.json",
		"/home/someone/.cqlai.json",
		"/home/someone/.config/cqlai/config.json",
	}, ConfigPaths())

	// A configuration that came from nowhere is written to the first of the
	// person's own places, not into whatever directory cqlai started in.
	assert.Equal(t, DefaultConfigPath(), (&Config{}).SavePath())
	assert.Equal(t, "/elsewhere/cqlai.json", (&Config{SourcePath: "/elsewhere/cqlai.json"}).SavePath())
}

// TestTheFileIsFoundWhereItHasAlwaysBeen.
func TestTheFileIsFoundWhereItHasAlwaysBeen(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	older := filepath.Join(home, ".cqlai.json")
	require.NoError(t, os.WriteFile(older, []byte(`{"host":"from-the-old-place"}`), 0o600))

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "from-the-old-place", cfg.Host)
	assert.Equal(t, older, cfg.SourcePath, "and is written back where it was found")
}

// TestTheNewPlaceIsReadFirst, for anyone who has both.
func TestTheNewPlaceIsReadFirst(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, os.MkdirAll(filepath.Join(home, ".cassandra"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".cassandra", "cqlai.json"),
		[]byte(`{"host":"beside-cqlshrc"}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".cqlai.json"),
		[]byte(`{"host":"from-the-old-place"}`), 0o600))

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "beside-cqlshrc", cfg.Host)
}
