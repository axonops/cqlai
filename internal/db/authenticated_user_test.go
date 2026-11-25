package db

import (
	"testing"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticatedUser(t *testing.T) {
	loggedIn := &gocql.ClusterConfig{Authenticator: gocql.PasswordAuthenticator{Username: "alice", Password: "x"}}

	tests := []struct {
		name    string
		session *Session
		want    string
	}{
		{"logged in", &Session{username: "alice", cluster: loggedIn}, "alice"},
		{"a long Vault user", &Session{username: "v_token_readonly_qjfz8r6gk75quopana9p_1795464675", cluster: loggedIn},
			"v_token_readonly_qjfz8r6gk75quopana9p_1795464675"},
		{"a username with no password does not log in", &Session{username: "cassandra", cluster: &gocql.ClusterConfig{}}, ""},
		{"no cluster", &Session{username: "alice"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.session.AuthenticatedUser())
		})
	}
}
