package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/router"
	"github.com/axonops/cqlai/internal/session"
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
	ConnectTimeout      int  // Connection timeout in seconds
	RequestTimeout      int  // Request timeout in seconds
	Debug               bool // Enable debug logging
}

// AISelectionResultMsg is sent when user completes a selection
type AISelectionResultMsg struct {
	Selection     string
	SelectionType string // The type of selection (e.g., "keyspace", "table")
	Cancelled     bool
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
	historyViewport          viewport.Model // For command history
	tableViewport            viewport.Model // For current table display
	input                    textinput.Model
	topBar                   TopBarModel
	statusBar                StatusBarModel
	lastCommand              string
	commandHistory           []string
	historyIndex             int
	fullHistoryContent       string // Full history content (not limited by viewport)
	session                  *db.Session
	sessionManager           *session.Manager // Application state manager
	aiConfig                 *config.AIConfig // AI configuration
	styles                   *Styles
	ready                    bool
	lastQueryTime            time.Duration
	rowCount                 int
	completionEngine         *completion.CompletionEngine
	completions              []string
	completionIndex          int
	showCompletions          bool
	completionScrollOffset   int // Track scroll position in completion list
	confirmExit              bool
	modal                    Modal
	aiModal                  AIModal             // AI-powered CQL generation modal
	showAIModal              bool                // Whether AI modal is active
	aiConversationID         string              // Current AI conversation ID for stateful interactions
	aiSelectionModal         *AISelectionModal   // AI selection modal for user choices
	showHistoryModal         bool                // Whether to show command history modal
	historyModalIndex        int                 // Currently selected item in history modal
	historyModalScrollOffset int                 // Track scroll position in history modal
	horizontalOffset         int                 // For horizontal scrolling of tables
	lastTableData            [][]string          // Store the last table data for horizontal scrolling
	tableWidth               int                 // Width of the full table (before truncation)
	tableHeaders             []string            // Store column headers for sticky display
	columnWidths             []int               // Store column widths for proper alignment
	hasTable                 bool                // Whether we're currently displaying a table
	viewMode                 string              // "history", "table", "trace", or "ai_info"
	showDataTypes            bool                // Whether to show column data types in table headers
	columnTypes              []string            // Store column data types
	
	// AI info request view
	aiInfoRequestActive      bool                // Whether AI info request view is active
	aiInfoRequestMessage     string              // The AI's message/question
	aiInfoRequestInput       textinput.Model     // Input for user response
	
	// Tracing support
	traceViewport            viewport.Model      // Viewport for trace results
	hasTrace                 bool                // Whether we have trace data to display
	traceData                [][]string          // Store trace results
	traceHeaders             []string            // Store trace column headers
	traceInfo                *db.TraceInfo       // Store trace session info
	traceHorizontalOffset    int                 // Horizontal scroll offset for trace table
	traceTableWidth          int                 // Full width of trace table
	traceColumnWidths        []int               // Column widths for trace table

	// Sliding window for large result sets
	slidingWindow *SlidingWindowTable // Manages memory-limited table data
	
	// Window dimensions
	windowWidth              int                 // Terminal window width
	windowHeight             int                 // Terminal window height

	// Multi-line mode
	multiLineMode   bool     // Whether we're in multi-line mode
	multiLineBuffer []string // Buffer for multi-line commands

	// History search
	historyManager            *HistoryManager
	historySearchMode         bool     // Whether we're in Ctrl+R history search mode
	historySearchQuery        string   // Current search query
	historySearchResults      []string // Filtered history results
	historySearchIndex        int      // Currently selected item in search results
	historySearchScrollOffset int      // Scroll offset for history search modal
}

// NewMainModel creates a new MainModel.
func NewMainModel() (*MainModel, error) {
	return NewMainModelWithOptions(false)
}

// NewMainModelWithOptions creates a new MainModel with options.
func NewMainModelWithOptions(noConfirm bool) (*MainModel, error) {
	options := ConnectionOptions{
		RequireConfirmation: !noConfirm,
	}
	return NewMainModelWithConnectionOptions(options)
}

// NewMainModelWithConnectionOptions creates a new MainModel with connection options.
func NewMainModelWithConnectionOptions(options ConnectionOptions) (*MainModel, error) {
	ti := textinput.New()
	ti.Placeholder = "Enter CQL command..."
	ti.Focus()
	ti.CharLimit = 256

	styles := DefaultStyles()

	ti.Prompt = styles.AccentText.Render("> ")
	ti.PlaceholderStyle = styles.MutedText

	infoReplyInput := textinput.New()
	infoReplyInput.Placeholder = "Type your response..."
	infoReplyInput.CharLimit = 500
	infoReplyInput.Width = 50

	// Load configuration from file and environment
	cfg, err := config.LoadConfig()
	if err != nil {
		// Use defaults if config file not found
		cfg = &config.Config{
			Host:                "127.0.0.1",
			Port:                9042,
			RequireConfirmation: true,
			AI: &config.AIConfig{
				Provider: "mock",
			},
		}
	}
	
	// Enable debug logging if configured (from config file or command-line)
	if cfg.Debug || options.Debug {
		logger.SetDebugEnabled(true)
	}

	// Override with command-line options
	if options.Host != "" {
		cfg.Host = options.Host
	}
	if options.Port != 0 {
		cfg.Port = options.Port
	}
	if options.Keyspace != "" {
		cfg.Keyspace = options.Keyspace
	}
	if options.Username != "" {
		cfg.Username = options.Username
	}
	if options.Password != "" {
		cfg.Password = options.Password
	}
	// RequireConfirmation handled specially since false is a valid override
	if options.Host != "" || options.Port != 0 || options.Keyspace != "" ||
		options.Username != "" || options.Password != "" {
		cfg.RequireConfirmation = options.RequireConfirmation
	}

	dbSession, err := db.NewSessionWithOptions(db.SessionOptions{
		Host:           cfg.Host,
		Port:           cfg.Port,
		Keyspace:       cfg.Keyspace,
		Username:       cfg.Username,
		Password:       cfg.Password,
		SSL:            cfg.SSL,
		ConnectTimeout: options.ConnectTimeout,
		RequestTimeout: options.RequestTimeout,
	})
	if err != nil {
		return nil, err
	}

	// Create session manager for application state
	sessionMgr := session.NewManager(cfg)

	// Initialize router with session manager
	router.InitRouter(sessionMgr)

	completionEngine := completion.NewCompletionEngine(dbSession, sessionMgr)

	// Initialize history manager
	historyManager, err := NewHistoryManager()
	if err != nil {
		// Log warning but don't fail - history will work in-memory only
		fmt.Fprintf(os.Stderr, "Warning: could not initialize history manager: %v\n", err)
		historyManager = &HistoryManager{history: []string{}}
	}

	// Load command history from the history manager
	commandHistory := historyManager.GetHistory()

	return &MainModel{
		topBar:                    NewTopBarModel(),
		statusBar:                 NewStatusBarModel(),
		input:                     ti,
		session:                   dbSession,
		sessionManager:            sessionMgr,
		aiConfig:                  cfg.AI,
		styles:                    styles,
		commandHistory:            commandHistory,
		historyIndex:              -1,
		fullHistoryContent:        "", // Will be initialized with welcome message in Init()
		completionEngine:          completionEngine,
		completions:               []string{},
		completionIndex:           -1,
		showCompletions:           false,
		completionScrollOffset:    0,
		horizontalOffset:          0,
		lastTableData:             nil,
		tableWidth:                0,
		tableHeaders:              nil,
		columnWidths:              nil,
		hasTable:                  false,
		viewMode:                  "history",
		hasTrace:                  false,
		traceData:                 nil,
		traceHeaders:              nil,
		historyManager:            historyManager,
		historySearchMode:         false,
		historySearchQuery:        "",
		historySearchResults:      []string{},
		historySearchIndex:        0,
		historySearchScrollOffset: 0,
	}, nil
}

// Init initializes the main model.
func (m *MainModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update updates the main model.
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Store the actual window dimensions
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		
		headerHeight := 1 // top bar
		footerHeight := 1 // status bar
		inputHeight := 1  // text input
		newWidth := msg.Width
		newHeight := msg.Height - headerHeight - footerHeight - inputHeight

		if !m.ready {
			// Initialize viewports
			m.historyViewport = viewport.New(newWidth, newHeight)
			m.tableViewport = viewport.New(newWidth, newHeight)
			m.traceViewport = viewport.New(newWidth, newHeight)
			welcomeMsg := m.getWelcomeMessage()
			m.fullHistoryContent = welcomeMsg
			m.historyViewport.SetContent(welcomeMsg)
			m.ready = true
		} else {
			// Resize viewports
			m.historyViewport.Width = newWidth
			m.historyViewport.Height = newHeight
			m.tableViewport.Width = newWidth
			m.tableViewport.Height = newHeight
			m.traceViewport.Width = newWidth
			m.traceViewport.Height = newHeight
		}

		m.input.Width = newWidth
		return m, nil

	case tea.KeyMsg:
		updatedModel, cmd := m.handleKeyboardInput(msg)
		return updatedModel, cmd

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
					switch interactionReq.Type {
					case "selection":
						logger.DebugfToFile("AI", "User selection needed for: %s", interactionReq.SelectionType)
						// Hide the AI modal temporarily
						m.showAIModal = false
						// Show selection modal
						m.aiSelectionModal = NewAISelectionModal(interactionReq.SelectionType, interactionReq.SelectionOptions)
						return m, nil
					case "info":
						logger.DebugfToFile("AI", "More info needed: %s", interactionReq.InfoMessage)
						// Hide the AI modal temporarily
						m.showAIModal = false
						// Show info request modal
						// Initialize the input if needed
						if m.aiInfoRequestInput.Value() == "" {
							input := textinput.New()
							input.Placeholder = "Type your response..."
							input.Focus()
							input.CharLimit = 500
							input.Width = 80
							m.aiInfoRequestInput = input
						}
						
						// Switch to AI info view
						m.aiInfoRequestActive = true
						m.aiInfoRequestMessage = interactionReq.InfoMessage
						m.viewMode = "ai_info"
						m.aiInfoRequestInput.Focus()
						m.showAIModal = false
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
			m.fullHistoryContent += "\n" + m.styles.MutedText.Render("AI generation cancelled.")
			m.historyViewport.SetContent(m.fullHistoryContent)
			m.historyViewport.GotoBottom()
		} else {
			logger.DebugfToFile("AI", "User selected %s: %s", msg.SelectionType, msg.Selection)
			// Continue AI generation with the selected value
			m.aiSelectionModal = nil

			// Re-show the AI modal in generating state
			m.showAIModal = true
			m.aiModal.State = AIModalStateGenerating

			// Build contextual response with type information
			contextualResponse := fmt.Sprintf("User selected %s: %s", msg.SelectionType, msg.Selection)

			// Continue the existing conversation with the user's selection
			if m.aiConversationID != "" {
				logger.DebugfToFile("AI", "Continuing conversation %s with selection: %s", m.aiConversationID, contextualResponse)
				return m, continueAIConversation(m.aiConfig, m.aiConversationID, contextualResponse)
			} else {
				// Fallback if no conversation ID (shouldn't happen)
				logger.DebugfToFile("AI", "Warning: No conversation ID, starting new conversation")
				contextualRequest := fmt.Sprintf("%s\n%s", m.aiModal.UserRequest, contextualResponse)
				return m, generateAICQL(m.session, m.aiConfig, contextualRequest)
			}
		}
		return m, nil

	case AIRequestMoreInfoMsg:
		// AI needs more information from user - use full-screen view
		logger.DebugfToFile("AI", "AI requesting more info: %s", msg.Message)
		
		// Initialize the input if needed
		if m.aiInfoRequestInput.Value() == "" {
			input := textinput.New()
			input.Placeholder = "Type your response..."
			input.Focus()
			input.CharLimit = 500
			input.Width = 80
			m.aiInfoRequestInput = input
		}
		
		// Switch to AI info view
		m.aiInfoRequestActive = true
		m.aiInfoRequestMessage = msg.Message
		m.viewMode = "ai_info"
		m.aiInfoRequestInput.Focus()
		
		// Close any modals
		m.showAIModal = false
		return m, nil

	case AIInfoResponseMsg:
		// User provided additional information (or cancelled)
		if msg.Cancelled {
			logger.DebugfToFile("AI", "User cancelled info request")
			// Cancel the AI operation and clear conversation
			m.showAIModal = false
			m.aiModal = AIModal{}
			m.aiInfoRequestActive = false
			m.aiInfoRequestInput.SetValue("")
			m.aiConversationID = ""
			// Add cancellation message
			m.fullHistoryContent += "\n" + m.styles.MutedText.Render("AI generation cancelled.")
			m.historyViewport.SetContent(m.fullHistoryContent)
			m.historyViewport.GotoBottom()
		} else {
			logger.DebugfToFile("AI", "User provided info: %s", msg.Response)
			// Continue AI generation with the additional information
			m.aiInfoRequestActive = false
			m.aiInfoRequestInput.SetValue("")

			// Re-show the AI modal in generating state
			m.showAIModal = true
			m.aiModal.State = AIModalStateGenerating

			// Continue the existing conversation with the additional info
			if m.aiConversationID != "" {
				logger.DebugfToFile("AI", "Continuing conversation %s with info: %s", m.aiConversationID, msg.Response)
				return m, continueAIConversation(m.aiConfig, m.aiConversationID, msg.Response)
			} else {
				// Fallback if no conversation ID (shouldn't happen)
				logger.DebugfToFile("AI", "Warning: No conversation ID, starting new conversation")
				contextualRequest := fmt.Sprintf("%s\nAdditional info: %s", m.aiModal.UserRequest, msg.Response)
				m.aiModal.UserRequest = contextualRequest
				return m, generateAICQL(m.session, m.aiConfig, contextualRequest)
			}
		}
		return m, nil
	}

	// Only update viewport for mouse wheel events
	switch msg.(type) {
	case tea.MouseMsg:
		var vpCmd tea.Cmd
		// Update the appropriate viewport based on mode
		switch {
		case m.viewMode == "trace" && m.hasTrace:
			m.traceViewport, vpCmd = m.traceViewport.Update(msg)
		case m.viewMode == "table" && m.hasTable:
			m.tableViewport, vpCmd = m.tableViewport.Update(msg)
		default:
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
