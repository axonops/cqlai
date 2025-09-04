package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/axonops/cqlai/internal/logger"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// handleAICommand handles .ai commands
func (m *MainModel) handleAICommand(command string) (*MainModel, tea.Cmd) {
	// Log the AI command
	logger.DebugfToFile("AI", "User AI command: %s", command)

	// Add AI command to command history
	m.commandHistory = append(m.commandHistory, command)
	m.historyIndex = -1
	m.lastCommand = command

	// Save to persistent history
	if m.historyManager != nil {
		if err := m.historyManager.SaveCommand(command); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not save AI command to history: %v\n", err)
		}
	}

	// Add command to history
	m.fullHistoryContent += "\n" + m.styles.AccentText.Render("> "+command)
	m.historyViewport.SetContent(m.fullHistoryContent)
	m.historyViewport.GotoBottom()

	// Extract the natural language request
	userRequest := strings.TrimSpace(command[3:])
	if userRequest == "" {
		// Show error for empty request
		m.fullHistoryContent += "\n" + m.styles.ErrorText.Render("Error: Please provide a request after .ai")
		m.historyViewport.SetContent(m.fullHistoryContent)
		m.historyViewport.GotoBottom()
		m.input.Reset()
		return m, nil
	}

	// Log the extracted request
	logger.DebugfToFile("AI", "Extracted AI request: %s", userRequest)

	// Clear any existing conversation ID for new .ai command
	// The new conversation ID will be set when we receive the response
	m.aiConversationID = ""

	// Initialize AI conversation view if needed
	if m.aiConversationInput.Value() == "" {
		input := textinput.New()
		input.Placeholder = ""
		input.Prompt = "> "
		input.Focus()
		input.CharLimit = 500
		input.Width = m.historyViewport.Width - 10
		m.aiConversationInput = input
		
		// Initialize conversation viewport
		m.aiConversationViewport = viewport.New(m.historyViewport.Width, m.historyViewport.Height)
		// Clear messages for new conversation
		m.aiConversationMessages = []AIMessage{}
	}
	
	// Add user's initial request to raw messages
	m.aiConversationMessages = append(m.aiConversationMessages, AIMessage{
		Role:    "user",
		Content: userRequest,
	})
	// Rebuild the conversation with proper wrapping
	m.rebuildAIConversation()
	
	// Switch to AI conversation view
	m.aiConversationActive = true
	m.viewMode = "ai"
	m.aiProcessing = true
	m.aiConversationInput.SetValue("")
	m.aiConversationInput.Focus()
	m.input.Reset()

	// Start AI generation in background
	return m, generateAICQL(m.session, m.aiConfig, userRequest)
}