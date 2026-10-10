package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPasswordsAreNotLogged(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"CREATE ROLE r WITH PASSWORD = 'hunter2' AND LOGIN = true", "CREATE ROLE r WITH PASSWORD = '***' AND LOGIN = true"},
		{"CREATE USER u WITH PASSWORD 'hun''ter2' SUPERUSER", "CREATE USER u WITH PASSWORD '***' SUPERUSER"},
		{"ALTER ROLE r WITH HASHED PASSWORD = $$abc$$", "ALTER ROLE r WITH HASHED PASSWORD = '***'"},
		{"Line 4: password = hunter2", "Line 4: password = ***"},
		{"Line 4: password: hunter2", "Line 4: password: ***"},
		{`{"username": "u", "password": "hun\"ter2"}`, `{"username": "u", "password": "***"}`},
		// Not a password, so left alone.
		{"Overriding password with command-line option", "Overriding password with command-line option"},
		{"SELECT * FROM ks.t WHERE id = 'hunter2'", "SELECT * FROM ks.t WHERE id = 'hunter2'"},
	} {
		if got := maskPasswords(tc.in); got != tc.want {
			t.Errorf("maskPasswords(%q)\n got %q\nwant %q", tc.in, got, tc.want)
		}
	}
}

func TestTheLogFileHasNoPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "debug.log")
	t.Setenv("CQLAI_DEBUG_LOG_PATH", path)
	SetDebugEnabled(true)
	defer SetDebugEnabled(false)

	DebugfToFile("ExecuteCQLQuery", "Called with query: %s", "CREATE ROLE r WITH PASSWORD = 'hunter2'")
	DebugfToFile("CQLSHRC", "Line %d: %s", 3, "password = hunter2")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "hunter2") {
		t.Errorf("the password is in the log:\n%s", data)
	}
}
