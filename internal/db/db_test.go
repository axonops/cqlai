package db

import (
	"testing"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// TestGetCluster verifies that GetCluster() exposes the cluster configuration
func TestGetCluster(t *testing.T) {
	// Create a mock cluster config
	cluster := gocql.NewCluster("127.0.0.1:9042")
	cluster.Consistency = gocql.Quorum
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "testuser",
		Password: "testpass",
	}

	// Create a session with this cluster
	// Note: We can't actually create a real session without Cassandra running,
	// so we'll create a Session struct manually for testing
	session := &Session{
		cluster:  cluster,
		username: "testuser",
	}

	// Test GetCluster()
	retrievedCluster := session.GetCluster()

	if retrievedCluster == nil {
		t.Fatal("GetCluster() returned nil")
	}

	if retrievedCluster != cluster {
		t.Error("GetCluster() did not return the same cluster reference")
	}

	// Verify cluster properties are preserved
	if retrievedCluster.Consistency != gocql.Quorum {
		t.Errorf("Expected consistency Quorum, got %v", retrievedCluster.Consistency)
	}

	if auth, ok := retrievedCluster.Authenticator.(gocql.PasswordAuthenticator); ok {
		if auth.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got '%s'", auth.Username)
		}
		if auth.Password != "testpass" {
			t.Errorf("Expected password 'testpass', got '%s'", auth.Password)
		}
	} else {
		t.Error("Authenticator is not a PasswordAuthenticator")
	}
}

// TestNewSessionFromClusterNilCluster verifies error handling for nil cluster
func TestNewSessionFromClusterNilCluster(t *testing.T) {
	_, err := NewSessionFromCluster(nil, "testuser", false)
	if err == nil {
		t.Error("Expected error for nil cluster, got nil")
	}

	expectedMsg := "cluster config cannot be nil"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestSessionSharesAuthUnit verifies cluster auth preservation (unit test, no connection)
func TestSessionSharesAuthUnit(t *testing.T) {
	// Create cluster with authentication
	cluster := gocql.NewCluster("127.0.0.1:9042")
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "testuser",
		Password: "testpass",
	}

	// Create a session wrapper (without actual connection)
	s1 := &Session{
		cluster:  cluster,
		username: "testuser",
	}

	// Get cluster for second session
	retrievedCluster := s1.GetCluster()

	// Verify auth is preserved
	if auth, ok := retrievedCluster.Authenticator.(gocql.PasswordAuthenticator); ok {
		if auth.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got '%s'", auth.Username)
		}
		if auth.Password != "testpass" {
			t.Errorf("Expected password 'testpass', got '%s'", auth.Password)
		}
	} else {
		t.Error("Authenticator not preserved in cluster config")
	}
}
