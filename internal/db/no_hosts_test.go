package db

import (
	"errors"
	"fmt"
	"testing"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

func TestIsNoHostsAvailable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"gocql no connections", gocql.ErrNoConnections, true},
		{"wrapped no connections", fmt.Errorf("query failed: %w", gocql.ErrNoConnections), true},
		{"generic error", errors.New("some other error"), false},
		{"timeout", gocql.ErrTimeoutNoResponse, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNoHostsAvailable(tt.err); got != tt.want {
				t.Errorf("IsNoHostsAvailable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
