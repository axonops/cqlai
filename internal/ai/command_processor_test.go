package ai

import (
	"strings"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCmd  CommandType
		wantArg  string
		wantFound bool
	}{
		{
			name:      "fuzzy search command",
			input:     "FUZZY_SEARCH:graph",
			wantCmd:   CommandFuzzySearch,
			wantArg:   "graph",
			wantFound: true,
		},
		{
			name:      "get schema command",
			input:     "GET_SCHEMA:graphql_test.users",
			wantCmd:   CommandGetSchema,
			wantArg:   "graphql_test.users",
			wantFound: true,
		},
		{
			name:      "list keyspaces command",
			input:     "LIST_KEYSPACES",
			wantCmd:   CommandListKeyspaces,
			wantArg:   "",
			wantFound: true,
		},
		{
			name:      "list tables command",
			input:     "LIST_TABLES:graphql_test",
			wantCmd:   CommandListTables,
			wantArg:   "graphql_test",
			wantFound: true,
		},
		{
			name:      "user selection command",
			input:     "USER_SELECTION:keyspace:test1,test2,test3",
			wantCmd:   CommandUserSelection,
			wantArg:   "keyspace:test1,test2,test3",
			wantFound: true,
		},
		{
			name:      "not enough info with message",
			input:     "NOT_ENOUGH_INFO:Please specify which table you want to query",
			wantCmd:   CommandNotEnoughInfo,
			wantArg:   "Please specify which table you want to query",
			wantFound: true,
		},
		{
			name:      "not enough info with detailed message",
			input:     "NOT_ENOUGH_INFO:I found multiple tables that could match. Please specify the exact keyspace and table name you want to query.",
			wantCmd:   CommandNotEnoughInfo,
			wantArg:   "I found multiple tables that could match. Please specify the exact keyspace and table name you want to query.",
			wantFound: true,
		},
		{
			name:      "no command found",
			input:     "This is just regular text",
			wantCmd:   "",
			wantArg:   "",
			wantFound: false,
		},
		{
			name:      "command in multiline response",
			input:     "Let me search for that.\nFUZZY_SEARCH:users\nSearching now...",
			wantCmd:   CommandFuzzySearch,
			wantArg:   "users",
			wantFound: true,
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
	cmd := CommandUserSelection
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
	cmd := CommandNotEnoughInfo
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