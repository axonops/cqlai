package ui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxHistorySize = 1000
	historyFile    = "history"
)

// HistoryManager manages command history persistence
type HistoryManager struct {
	historyPath string
	history     []string
}

// NewHistoryManager creates a new history manager
func NewHistoryManager() (*HistoryManager, error) {
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	// Create ~/.cqlai directory if it doesn't exist
	cqlaiDir := filepath.Join(home, ".cqlai")
	if err := os.MkdirAll(cqlaiDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create .cqlai directory: %v", err)
	}

	historyPath := filepath.Join(cqlaiDir, historyFile)
	
	hm := &HistoryManager{
		historyPath: historyPath,
		history:     []string{},
	}

	// Load existing history
	if err := hm.loadHistory(); err != nil {
		// Log error but don't fail - history will start fresh
		fmt.Fprintf(os.Stderr, "Warning: could not load history: %v\n", err)
	}

	return hm, nil
}

// loadHistory loads history from file
func (hm *HistoryManager) loadHistory() error {
	file, err := os.Open(hm.historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's okay
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			hm.history = append(hm.history, line)
		}
	}

	// Keep only the most recent entries if history is too large
	if len(hm.history) > maxHistorySize {
		hm.history = hm.history[len(hm.history)-maxHistorySize:]
	}

	return scanner.Err()
}

// SaveCommand adds a command to history and saves to file
func (hm *HistoryManager) SaveCommand(command string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil
	}

	// Don't add duplicate consecutive commands
	if len(hm.history) > 0 && hm.history[len(hm.history)-1] == command {
		return nil
	}

	// Add to in-memory history
	hm.history = append(hm.history, command)

	// Keep history size under control
	if len(hm.history) > maxHistorySize {
		hm.history = hm.history[1:]
	}

	// Append to file
	file, err := os.OpenFile(hm.historyPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open history file: %v", err)
	}
	defer file.Close()

	if _, err := fmt.Fprintln(file, command); err != nil {
		return fmt.Errorf("failed to write to history file: %v", err)
	}

	return nil
}

// GetHistory returns the command history
func (hm *HistoryManager) GetHistory() []string {
	// Return a copy to prevent external modification
	result := make([]string, len(hm.history))
	copy(result, hm.history)
	return result
}

// SearchHistory searches for commands containing the given string
func (hm *HistoryManager) SearchHistory(query string) []string {
	if query == "" {
		return hm.GetHistory()
	}

	var matches []string
	queryLower := strings.ToLower(query)
	
	// Search in reverse order (most recent first)
	for i := len(hm.history) - 1; i >= 0; i-- {
		if strings.Contains(strings.ToLower(hm.history[i]), queryLower) {
			matches = append(matches, hm.history[i])
		}
	}
	
	return matches
}