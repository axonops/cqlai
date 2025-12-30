package mcp

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startCQLAIWithMCP starts CQLAI with MCP auto-start using the specified config file
func startCQLAIWithMCP(t *testing.T, configFile string) *exec.Cmd {
	// Binary is in project root, tests run from test/integration/mcp
	// Correct syntax: ./cqlai 127.0.0.1 -u cassandra -p cassandra
	cmd := exec.Command("../../../cqlai",
		"127.0.0.1",
		"-u", "cassandra",
		"-p", "cassandra",
		"--mcpstart",
		"--mcpconfig", configFile,
	)

	err := cmd.Start()
	require.NoError(t, err, "Failed to start cqlai")

	return cmd
}

// stopCQLAI stops the CQLAI process
func stopCQLAI(cmd *exec.Cmd) {
	if cmd != nil && cmd.Process != nil {
		cmd.Process.Kill()
		cmd.Wait() // Clean up zombie
	}
}

// callTool calls an MCP tool via Unix socket and returns the response
func callTool(t *testing.T, toolName string, args map[string]any) map[string]any {
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      time.Now().UnixNano(),
		"method":  "tools/call",
		"params": map[string]any{
			"name":      toolName,
			"arguments": args,
		},
	}

	requestJSON, err := json.Marshal(request)
	require.NoError(t, err)

	// Call via nc with timeout
	shellCmd := fmt.Sprintf("echo '%s' | timeout 5 nc -U /tmp/cqlai-mcp-test.sock 2>&1", string(requestJSON))
	output, err := exec.Command("sh", "-c", shellCmd).CombinedOutput()

	if err != nil {
		t.Logf("nc command failed: %v, output: %s", err, string(output))
		return nil
	}

	var response map[string]any
	if err := json.Unmarshal(output, &response); err != nil {
		t.Logf("Failed to parse response: %s", string(output))
		return nil
	}

	return response
}

// assertIsError asserts that the response is an error
func assertIsError(t *testing.T, resp map[string]any, msg string) {
	if resp == nil {
		t.Fatalf("nil response for: %s", msg)
	}
	result, ok := resp["result"].(map[string]any)
	require.True(t, ok, "Response should have result")

	isError, _ := result["isError"].(bool)
	assert.True(t, isError, msg)
}

// assertNotError asserts that the response is successful
func assertNotError(t *testing.T, resp map[string]any, msg string) {
	if resp == nil {
		t.Fatalf("nil response for: %s", msg)
	}
	result, ok := resp["result"].(map[string]any)
	require.True(t, ok, "Response should have result")

	isError, _ := result["isError"].(bool)
	assert.False(t, isError, msg)
}

// assertContains asserts that the response text contains a substring
func assertContains(t *testing.T, resp map[string]any, substring string) {
	if resp == nil {
		return
	}
	result := resp["result"].(map[string]any)
	content := result["content"].([]any)
	if len(content) > 0 {
		text := content[0].(map[string]any)["text"].(string)
		assert.Contains(t, text, substring)
	}
}

// extractText extracts the text content from a response
func extractText(t *testing.T, resp map[string]any) string {
	if resp == nil {
		return ""
	}
	result := resp["result"].(map[string]any)
	content := result["content"].([]any)
	if len(content) > 0 {
		return content[0].(map[string]any)["text"].(string)
	}
	return ""
}

// extractRequestID extracts a request ID (req_NNN) from error text
func extractRequestID(text string) string {
	// Simple extraction - look for "req_" pattern
	if idx := findIndex(text, "req_"); idx >= 0 {
		end := idx + 7 // "req_NNN" is 7 chars
		if end <= len(text) {
			return text[idx:end]
		}
	}
	return ""
}

func findIndex(s, substr string) int {
	// Simple string search
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
