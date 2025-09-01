package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCQLSHRC(t *testing.T) {
	// Create a temporary cqlshrc file
	tmpDir := t.TempDir()
	cqlshrcPath := filepath.Join(tmpDir, "cqlshrc")

	cqlshrcContent := `; Test CQLSHRC file
[connection]
hostname = testhost.example.com
port = 9043
ssl = true

[authentication]
keyspace = test_keyspace
credentials = ~/.cassandra/credentials

[auth_provider]
module = cassandra.auth
classname = PlainTextAuthProvider
username = testuser

[ssl]
certfile = ~/certs/ca.pem
userkey = ~/certs/client.key
usercert = ~/certs/client.cert
validate = false
`

	if err := os.WriteFile(cqlshrcPath, []byte(cqlshrcContent), 0600); err != nil {
		t.Fatalf("Failed to create test cqlshrc file: %v", err)
	}

	// Test loading the cqlshrc file
	config := &Config{
		Host: "localhost",
		Port: 9042,
	}

	if err := loadCQLSHRC(cqlshrcPath, config); err != nil {
		t.Fatalf("Failed to load cqlshrc: %v", err)
	}

	// Verify the loaded configuration
	if config.Host != "testhost.example.com" {
		t.Errorf("Expected host to be 'testhost.example.com', got '%s'", config.Host)
	}

	if config.Port != 9043 {
		t.Errorf("Expected port to be 9043, got %d", config.Port)
	}

	if config.Keyspace != "test_keyspace" {
		t.Errorf("Expected keyspace to be 'test_keyspace', got '%s'", config.Keyspace)
	}

	if config.Username != "testuser" {
		t.Errorf("Expected username to be 'testuser', got '%s'", config.Username)
	}

	if config.AuthProvider == nil {
		t.Error("Expected AuthProvider to be set")
	} else {
		if config.AuthProvider.Module != "cassandra.auth" {
			t.Errorf("Expected AuthProvider.Module to be 'cassandra.auth', got '%s'", config.AuthProvider.Module)
		}
		if config.AuthProvider.ClassName != "PlainTextAuthProvider" {
			t.Errorf("Expected AuthProvider.ClassName to be 'PlainTextAuthProvider', got '%s'", config.AuthProvider.ClassName)
		}
	}

	if config.SSL == nil {
		t.Error("Expected SSL config to be set")
	} else {
		if !config.SSL.Enabled {
			t.Error("Expected SSL to be enabled")
		}

		expectedCAPath := filepath.Join(os.Getenv("HOME"), "certs/ca.pem")
		if config.SSL.CAPath != expectedCAPath {
			t.Errorf("Expected CAPath to be '%s', got '%s'", expectedCAPath, config.SSL.CAPath)
		}

		expectedKeyPath := filepath.Join(os.Getenv("HOME"), "certs/client.key")
		if config.SSL.KeyPath != expectedKeyPath {
			t.Errorf("Expected KeyPath to be '%s', got '%s'", expectedKeyPath, config.SSL.KeyPath)
		}

		expectedCertPath := filepath.Join(os.Getenv("HOME"), "certs/client.cert")
		if config.SSL.CertPath != expectedCertPath {
			t.Errorf("Expected CertPath to be '%s', got '%s'", expectedCertPath, config.SSL.CertPath)
		}

		if !config.SSL.InsecureSkipVerify {
			t.Error("Expected InsecureSkipVerify to be true (validate=false)")
		}
	}
}

func TestLoadCredentialsFile(t *testing.T) {
	// Create a temporary credentials file
	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, "credentials")

	credContent := `; Credentials file
[PlainTextAuthProvider]
username = creduser
password = credpass123
`

	if err := os.WriteFile(credPath, []byte(credContent), 0600); err != nil {
		t.Fatalf("Failed to create test credentials file: %v", err)
	}

	// Test loading the credentials file
	config := &Config{}

	if err := loadCredentialsFile(credPath, config); err != nil {
		t.Fatalf("Failed to load credentials file: %v", err)
	}

	// Verify the loaded credentials
	if config.Username != "creduser" {
		t.Errorf("Expected username to be 'creduser', got '%s'", config.Username)
	}

	if config.Password != "credpass123" {
		t.Errorf("Expected password to be 'credpass123', got '%s'", config.Password)
	}
}