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
	
	if enabled {
		// Create the log file immediately when debug is enabled
		cwd, _ := os.Getwd()
		logPath := cwd + "/cqlai_debug.log"
		fmt.Fprintf(os.Stderr, "[INFO] Debug logging enabled. Log file: %s\n", logPath)
	}
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
	
	cwd, _ := os.Getwd()
	logPath := cwd + "/cqlai_debug.log"

	// Check if file exists to print message only once
	_, statErr := os.Stat(logPath)
	isNewFile := os.IsNotExist(statErr)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600) // #nosec G304: Potential file inclusion via variable
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARNING] Could not open debug log: %v\n", err)
		return
	}
	defer logFile.Close()

	// Notify user on first creation
	if isNewFile {
		fmt.Fprintf(os.Stderr, "[INFO] Created debug log file: %s\n", logPath)
	}

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
