package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/ui/completion"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// ConnectionOptions holds command-line connection options
type ConnectionOptions struct {
	Host                string
	Port                int
	Keyspace            string
	Username            string
	Password            string
	RequireConfirmation bool
}

// AISelectionResultMsg is sent when user completes a selection
type AISelectionResultMsg struct {
	Selection string
	Cancelled bool
}

// AIRequestUserSelectionMsg is sent when AI needs user to select from options
type AIRequestUserSelectionMsg struct {
	SelectionType string
	Options       []string
}

// AIRequestMoreInfoMsg is sent when AI needs more information from user
type AIRequestMoreInfoMsg struct {
	Message string
}

// AIInfoResponseMsg is sent when user provides additional information
type AIInfoResponseMsg struct {
	Response  string
	Cancelled bool
}

// MainModel is the main Bubble Tea model for the application.
type MainModel struct {
	historyViewport  viewport.Model  // For command history
	tableViewport    viewport.Model  // For current table display
	input            textinput.Model
	topBar           TopBarModel
	statusBar        StatusBarModel
	lastCommand      string
	commandHistory   []string
	historyIndex     int
	session          *db.Session
	styles           *Styles
	ready            bool
	lastQueryTime    time.Duration
	rowCount         int
	completionEngine *completion.CompletionEngine
	completions      []string
	completionIndex  int
	showCompletions  bool
	completionScrollOffset int  // Track scroll position in completion list
	confirmExit      bool
	modal            Modal
	aiModal          AIModal  // AI-powered CQL generation modal
	showAIModal      bool     // Whether AI modal is active
	aiConversationID string   // Current AI conversation ID for stateful interactions
	aiSelectionModal *AISelectionModal  // AI selection modal for user choices
	aiInfoRequestModal *AIInfoRequestModal  // AI info request modal for additional input
	showHistoryModal bool  // Whether to show command history modal
	historyModalIndex int  // Currently selected item in history modal
	historyModalScrollOffset int  // Track scroll position in history modal
	horizontalOffset int  // For horizontal scrolling of tables
	lastTableData    [][]string  // Store the last table data for horizontal scrolling
	tableWidth       int  // Width of the full table (before truncation)
	tableHeaders     []string  // Store column headers for sticky display
	columnWidths     []int  // Store column widths for proper alignment
	hasTable         bool  // Whether we're currently displaying a table
	viewMode         string  // "history" or "table"
	showDataTypes    bool  // Whether to show column data types in table headers
	columnTypes      []string  // Store column data types
	
	// Sliding window for large result sets
	slidingWindow    *SlidingWindowTable  // Manages memory-limited table data
	
	// Multi-line mode
	multiLineMode    bool     // Whether we're in multi-line mode
	multiLineBuffer  []string // Buffer for multi-line commands
	
	// History search
	historyManager   *HistoryManager
	historySearchMode bool  // Whether we're in Ctrl+R history search mode
	historySearchQuery string  // Current search query
	historySearchResults []string  // Filtered history results
	historySearchIndex int  // Currently selected item in search results
	historySearchScrollOffset int  // Scroll offset for history search modal
}

// NewMainModel creates a new MainModel.
func NewMainModel() (MainModel, error) {
	return NewMainModelWithOptions(false)
}

// NewMainModelWithOptions creates a new MainModel with options.
func NewMainModelWithOptions(noConfirm bool) (MainModel, error) {
	options := ConnectionOptions{
		RequireConfirmation: !noConfirm,
	}
	return NewMainModelWithConnectionOptions(options)
}

// NewMainModelWithConnectionOptions creates a new MainModel with connection options.
func NewMainModelWithConnectionOptions(options ConnectionOptions) (MainModel, error) {
	ti := textinput.New()
	ti.Placeholder = "Enter CQL command..."
	ti.Focus()
	ti.CharLimit = 256

	styles := DefaultStyles()

	ti.Prompt = styles.AccentText.Render("> ")
	ti.PlaceholderStyle = styles.MutedText

	session, err := db.NewSessionWithOptions(db.SessionOptions{
		Host:                options.Host,
		Port:                options.Port,
		Keyspace:            options.Keyspace,
		Username:            options.Username,
		Password:            options.Password,
		RequireConfirmation: options.RequireConfirmation,
	})
	if err != nil {
		return MainModel{}, err
	}

	completionEngine := completion.NewCompletionEngine(session)
	
	// Initialize history manager
	historyManager, err := NewHistoryManager()
	if err != nil {
		// Log warning but don't fail - history will work in-memory only
		fmt.Fprintf(os.Stderr, "Warning: could not initialize history manager: %v\n", err)
		historyManager = &HistoryManager{history: []string{}}
	}
	
	// Load command history from the history manager
	commandHistory := historyManager.GetHistory()

	return MainModel{
		topBar:           NewTopBarModel(),
		statusBar:        NewStatusBarModel(),
		input:            ti,
		session:          session,
		styles:           styles,
		commandHistory:   commandHistory,
		historyIndex:     -1,
		completionEngine: completionEngine,
		completions:      []string{},
		completionIndex:  -1,
		showCompletions:  false,
		completionScrollOffset: 0,
		horizontalOffset: 0,
		lastTableData:    nil,
		tableWidth:       0,
		tableHeaders:     nil,
		columnWidths:     nil,
		hasTable:         false,
		viewMode:         "history",
		historyManager:   historyManager,
		historySearchMode: false,
		historySearchQuery: "",
		historySearchResults: []string{},
		historySearchIndex: 0,
		historySearchScrollOffset: 0,
	}, nil
}

// Init initializes the main model.
func (m MainModel) Init() tea.Cmd {
	return textinput.Blink
}


// Update updates the main model.
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 1 // top bar
		footerHeight := 1 // status bar
		inputHeight := 1  // text input
		newWidth := msg.Width
		newHeight := msg.Height - headerHeight - footerHeight - inputHeight

		if !m.ready {
			// Initialize viewports
			m.historyViewport = viewport.New(newWidth, newHeight)
			m.tableViewport = viewport.New(newWidth, newHeight)
			m.historyViewport.SetContent(m.getWelcomeMessage())
			m.ready = true
		} else {
			// Resize viewports
			m.historyViewport.Width = newWidth
			m.historyViewport.Height = newHeight
			m.tableViewport.Width = newWidth
			m.tableViewport.Height = newHeight
		}

		m.input.Width = newWidth
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyboardInput(msg)
	
	case AICQLResultMsg:
		// Handle AI CQL generation result
		logger.DebugfToFile("AI", "Received AI result message")
		
		// Store the conversation ID if provided
		if msg.ConversationID != "" {
			m.aiConversationID = msg.ConversationID
			logger.DebugfToFile("AI", "Conversation ID: %s", msg.ConversationID)
		}
		
		if m.showAIModal && m.aiModal.State == AIModalStateGenerating {
			if msg.Error != nil {
				// Check if this is an interaction request
				if interactionReq, ok := msg.Error.(*ai.InteractionRequest); ok {
					if interactionReq.Type == "selection" {
						logger.DebugfToFile("AI", "User selection needed for: %s", interactionReq.SelectionType)
						// Hide the AI modal temporarily
						m.showAIModal = false
						// Show selection modal
						m.aiSelectionModal = NewAISelectionModal(interactionReq.SelectionType, interactionReq.SelectionOptions)
						return m, nil
					} else if interactionReq.Type == "info" {
						logger.DebugfToFile("AI", "More info needed: %s", interactionReq.InfoMessage)
						// Hide the AI modal temporarily
						m.showAIModal = false
						// Show info request modal
						m.aiInfoRequestModal = NewAIInfoRequestModal(interactionReq.InfoMessage)
						return m, nil
					}
				}
				// Regular error
				logger.DebugfToFile("AI", "AI generation failed: %v", msg.Error)
				m.aiModal.SetError(msg.Error)
			} else {
				logger.DebugfToFile("AI", "AI generation successful, showing preview")
				m.aiModal.SetResult(msg.Plan, msg.CQL)
			}
		}
		return m, nil
	
	case AIRequestUserSelectionMsg:
		// AI needs user to select from options
		logger.DebugfToFile("AI", "AI requesting user selection: type=%s, options=%v", msg.SelectionType, msg.Options)
		m.aiSelectionModal = NewAISelectionModal(msg.SelectionType, msg.Options)
		return m, nil
	
	case AISelectionResultMsg:
		// User completed selection (or cancelled)
		if msg.Cancelled {
			logger.DebugfToFile("AI", "User cancelled selection")
			// Cancel the AI operation and clear conversation
			m.showAIModal = false
			m.aiModal = AIModal{}
			m.aiSelectionModal = nil
			m.aiConversationID = ""
			// Add cancellation message
			historyContent := m.historyViewport.View() + "\n" + m.styles.MutedText.Render("AI generation cancelled.")
			m.historyViewport.SetContent(historyContent)
			m.historyViewport.GotoBottom()
		} else {
			logger.DebugfToFile("AI", "User selected: %s", msg.Selection)
			// Continue AI generation with the selected value
			m.aiSelectionModal = nil
			
			// Re-show the AI modal in generating state
			m.showAIModal = true
			m.aiModal.State = AIModalStateGenerating
			
			// Continue the existing conversation with the user's selection
			if m.aiConversationID != "" {
				logger.DebugfToFile("AI", "Continuing conversation %s with selection: %s", m.aiConversationID, msg.Selection)
				return m, continueAIConversation(m.aiConversationID, msg.Selection)
			} else {
				// Fallback if no conversation ID (shouldn't happen)
				logger.DebugfToFile("AI", "Warning: No conversation ID, starting new conversation")
				contextualRequest := fmt.Sprintf("%s\nUser selected: %s", m.aiModal.UserRequest, msg.Selection)
				return m, generateAICQL(m.session, contextualRequest)
			}
		}
		return m, nil
	
	case AIRequestMoreInfoMsg:
		// AI needs more information from user
		logger.DebugfToFile("AI", "AI requesting more info: %s", msg.Message)
		m.aiInfoRequestModal = NewAIInfoRequestModal(msg.Message)
		return m, nil
	
	case AIInfoResponseMsg:
		// User provided additional information (or cancelled)
		if msg.Cancelled {
			logger.DebugfToFile("AI", "User cancelled info request")
			// Cancel the AI operation and clear conversation
			m.showAIModal = false
			m.aiModal = AIModal{}
			m.aiInfoRequestModal = nil
			m.aiConversationID = ""
			// Add cancellation message
			historyContent := m.historyViewport.View() + "\n" + m.styles.MutedText.Render("AI generation cancelled.")
			m.historyViewport.SetContent(historyContent)
			m.historyViewport.GotoBottom()
		} else {
			logger.DebugfToFile("AI", "User provided info: %s", msg.Response)
			// Continue AI generation with the additional information
			m.aiInfoRequestModal = nil
			
			// Re-show the AI modal in generating state
			m.showAIModal = true
			m.aiModal.State = AIModalStateGenerating
			
			// Continue the existing conversation with the additional info
			if m.aiConversationID != "" {
				logger.DebugfToFile("AI", "Continuing conversation %s with info: %s", m.aiConversationID, msg.Response)
				return m, continueAIConversation(m.aiConversationID, msg.Response)
			} else {
				// Fallback if no conversation ID (shouldn't happen)
				logger.DebugfToFile("AI", "Warning: No conversation ID, starting new conversation")
				contextualRequest := fmt.Sprintf("%s\nAdditional info: %s", m.aiModal.UserRequest, msg.Response)
				m.aiModal.UserRequest = contextualRequest
				return m, generateAICQL(m.session, contextualRequest)
			}
		}
		return m, nil
	}

	// Only update viewport for mouse wheel events
	switch msg.(type) {
	case tea.MouseMsg:
		var vpCmd tea.Cmd
		// Update the appropriate viewport based on mode
		if m.viewMode == "table" && m.hasTable {
			m.tableViewport, vpCmd = m.tableViewport.Update(msg)
		} else {
			m.historyViewport, vpCmd = m.historyViewport.Update(msg)
		}
		m.input, _ = m.input.Update(msg)
		return m, vpCmd
	default:
		// Update input for other events
		var inputCmd tea.Cmd
		m.input, inputCmd = m.input.Update(msg)
		return m, inputCmd
	}
}