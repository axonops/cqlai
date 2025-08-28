package ai

import (
	"strings"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCmd  ToolName
		wantArg  string
		wantFound bool
	}{
		// JSON format tests
		{
			name:      "json fuzzy search",
			input:     `{"tool": "FUZZY_SEARCH", "params": {"query": "users"}}`,
			wantCmd:   ToolFuzzySearch,
			wantArg:   "users",
			wantFound: true,
		},
		{
			name:      "json get schema",
			input:     `{"tool": "GET_SCHEMA", "params": {"keyspace": "myapp", "table": "users"}}`,
			wantCmd:   ToolGetSchema,
			wantArg:   "myapp.users",
			wantFound: true,
		},
		{
			name:      "json list keyspaces",
			input:     `{"tool": "LIST_KEYSPACES", "params": {}}`,
			wantCmd:   ToolListKeyspaces,
			wantArg:   "",
			wantFound: true,
		},
		{
			name:      "json list tables",
			input:     `{"tool": "LIST_TABLES", "params": {"keyspace": "system"}}`,
			wantCmd:   ToolListTables,
			wantArg:   "system",
			wantFound: true,
		},
		{
			name:      "json user selection",
			input:     `{"tool": "USER_SELECTION", "params": {"type": "table", "options": ["users", "profiles", "settings"]}}`,
			wantCmd:   ToolUserSelection,
			wantArg:   "table:users,profiles,settings",
			wantFound: true,
		},
		{
			name:      "json not enough info",
			input:     `{"tool": "NOT_ENOUGH_INFO", "params": {"message": "Please specify the keyspace"}}`,
			wantCmd:   ToolNotEnoughInfo,
			wantArg:   "Please specify the keyspace",
			wantFound: true,
		},
		{
			name:      "json not relevant",
			input:     `{"tool": "NOT_RELEVANT", "params": {"message": "This is about MongoDB"}}`,
			wantCmd:   ToolNotRelevant,
			wantArg:   "This is about MongoDB",
			wantFound: true,
		},
		{
			name:      "json embedded in text",
			input:     "Let me search for that.\n{\"tool\": \"FUZZY_SEARCH\", \"params\": {\"query\": \"accounts\"}}\nSearching now...",
			wantCmd:   ToolFuzzySearch,
			wantArg:   "accounts",
			wantFound: true,
		},
		{
			name:      "no command found",
			input:     "This is just regular text",
			wantCmd:   "",
			wantArg:   "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, arg, found := ParseCommand(tt.input)
			if cmd != tt.wantCmd {
				t.Errorf("ParseCommand() cmd = %v, want %v", cmd, tt.wantCmd)
			}
			if arg != tt.wantArg {
				t.Errorf("ParseCommand() arg = %v, want %v", arg, tt.wantArg)
			}
			if found != tt.wantFound {
				t.Errorf("ParseCommand() found = %v, want %v", found, tt.wantFound)
			}
		})
	}
}

func TestExecuteCommand_UserSelection(t *testing.T) {
	// This test just verifies the formatting - actual execution requires globalAI to be set up
	cmd := ToolUserSelection
	arg := "table:users,accounts,sessions"
	
	result := ExecuteCommand(cmd, arg)
	
	// Without globalAI, should get error
	if result.Success {
		t.Error("Expected failure without globalAI initialized")
	}
	if result.Error == nil || !strings.Contains(result.Error.Error(), "AI system not initialized") {
		t.Errorf("Expected 'AI system not initialized' error, got: %v", result.Error)
	}
}

func TestExecuteCommand_NotEnoughInfo(t *testing.T) {
	// Test with the AI system not initialized - should get that error
	cmd := ToolNotEnoughInfo
	arg := "Please specify which table to query"
	
	result := ExecuteCommand(cmd, arg)
	
	// Without globalAI, should get error
	if result.Success {
		t.Error("Expected failure without globalAI initialized")
	}
	if result.Error == nil || !strings.Contains(result.Error.Error(), "AI system not initialized") {
		t.Errorf("Expected 'AI system not initialized' error, got: %v", result.Error)
	}
}