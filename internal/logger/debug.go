package logger

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"
)

var (
	debugEnabled bool
	debugMutex   sync.RWMutex
)

// SetDebugEnabled enables or disables debug logging
func SetDebugEnabled(enabled bool) {
	debugMutex.Lock()
	defer debugMutex.Unlock()
	debugEnabled = enabled
}

// IsDebugEnabled returns whether debug logging is enabled
func IsDebugEnabled() bool {
	debugMutex.RLock()
	defer debugMutex.RUnlock()
	return debugEnabled
}

// DebugToFile logs debug messages to a file
func DebugToFile(context string, message string) {
	if !IsDebugEnabled() {
		return
	}

	var logPath string
	logPath = os.Getenv("CQLAI_DEBUG_LOG_PATH")
	if logPath == "" {
		cwd, _ := os.Getwd()
		logPath = cwd + "/cqlai_debug.log"
	}
	// Check if file exists to print message only once
	_, statErr := os.Stat(logPath) //nolint:gosec // G703: log path is derived from known config
	isNewFile := os.IsNotExist(statErr)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600) // #nosec G304 G703 - Log path is derived from known config
	if err != nil {
		return
	}
	defer logFile.Close()

	// isNewFile is no longer used but keep for potential future use
	_ = isNewFile

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	fmt.Fprintf(logFile, "[%s] Context: %s | %s\n", timestamp, context, maskPasswords(message))
	_ = logFile.Sync()
}

// DebugfToFile logs formatted debug messages to a file
func DebugfToFile(context string, format string, args ...interface{}) {
	if !IsDebugEnabled() {
		return
	}
	message := fmt.Sprintf(format, args...)
	DebugToFile(context, message)
}

// Passwords in what is logged. Statements are logged as typed, and a cqlshrc
// line by line, so a CREATE ROLE or a [authentication] section would put a
// password in the debug log for anyone who can read the file.
var (
	// PASSWORD = 'x', PASSWORD 'x' and HASHED PASSWORD = $$x$$ in CQL.
	cqlPassword = regexp.MustCompile(`(?i)(\bpassword\s*=?\s*)('(?:[^']|'')*'|\$\$[\s\S]*?\$\$)`)
	// password = x in a cqlshrc or credentials file, and "password": "x".
	keyPassword  = regexp.MustCompile(`(?i)(\bpassword\s*[=:]\s*)([^\s'"$][^\r\n]*)`)
	jsonPassword = regexp.MustCompile(`(?i)("password"\s*:\s*)"(?:[^"\\]|\\.)*"`)
)

// maskPasswords is a message with every password in it replaced by ***.
func maskPasswords(message string) string {
	message = cqlPassword.ReplaceAllString(message, "${1}'***'")
	message = jsonPassword.ReplaceAllString(message, `${1}"***"`)
	return keyPassword.ReplaceAllString(message, "${1}***")
}
