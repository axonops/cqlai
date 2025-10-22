package logger

import (
	"fmt"
	"os"
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
	_, statErr := os.Stat(logPath)
	isNewFile := os.IsNotExist(statErr)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600) // #nosec G304: Potential file inclusion via variable
	if err != nil {
		return
	}
	defer logFile.Close()

	// isNewFile is no longer used but keep for potential future use
	_ = isNewFile

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	fmt.Fprintf(logFile, "[%s] Context: %s | %s\n", timestamp, context, message)
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
