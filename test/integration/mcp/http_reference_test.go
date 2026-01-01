//go:build integration
// +build integration

package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// HTTP Reference Implementation
// ============================================================================
//
// This file contains reference implementations for HTTP-based MCP testing.
// Once these work correctly, we'll use this pattern to update all other tests.
//

// HTTPTestContext holds test context for HTTP-based tests
type HTTPTestContext struct {
	Session      *db.Session
	MCPHandler   *router.MCPHandler
	HTTPHost     string
	HTTPPort     int
	APIKey       string
	BaseURL      string
	MCPSessionID string // MCP protocol session ID
}

// startMCPFromConfigHTTP starts MCP server using HTTP transport
func startMCPFromConfigHTTP(t *testing.T, configPath string) *HTTPTestContext {
	// Create session
	replSession, err := createTestREPLSession(t)
	require.NoError(t, err)

	// Init MCP handler
	err = router.InitMCPHandler(replSession)
	require.NoError(t, err)

	mcpHandler := router.GetMCPHandler()
	require.NotNil(t, mcpHandler)

	// Extract HTTP config from JSON file
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var jsonConfig map[string]any
	err = json.Unmarshal(data, &jsonConfig)
	require.NoError(t, err)

	// Get HTTP settings (with defaults)
	httpHost := "127.0.0.1"
	if h, ok := jsonConfig["http_host"].(string); ok {
		httpHost = h
	}

	httpPort := 8888
	if p, ok := jsonConfig["http_port"].(float64); ok {
		httpPort = int(p)
	}

	// Generate API key if not in config
	apiKey := ""
	if key, ok := jsonConfig["api_key"].(string); ok {
		apiKey = ai.ExpandEnvVar(key)
	}
	if apiKey == "" {
		// Generate for test
		apiKey, err = ai.GenerateAPIKey()
		require.NoError(t, err)
	}

	// Start server with config file
	cmd := fmt.Sprintf(".mcp start --config-file %s", configPath)
	if apiKey != "" && jsonConfig["api_key"] == nil {
		// If key not in config, pass via flag
		cmd += fmt.Sprintf(" --api-key %s", apiKey)
	}

	result := mcpHandler.HandleMCPCommand(cmd)
	require.Contains(t, result, "started successfully", "MCP start failed: %s", result)

	time.Sleep(500 * time.Millisecond) // Wait for HTTP server to start

	baseURL := fmt.Sprintf("http://%s:%d/mcp", httpHost, httpPort)

	ctx := &HTTPTestContext{
		Session:    replSession,
		MCPHandler: mcpHandler,
		HTTPHost:   httpHost,
		HTTPPort:   httpPort,
		APIKey:     apiKey,
		BaseURL:    baseURL,
	}

	// Initialize MCP session
	sessionID, err := initializeMCPSession(t, ctx)
	require.NoError(t, err, "Failed to initialize MCP session")
	ctx.MCPSessionID = sessionID

	return ctx
}

// initializeMCPSession sends MCP initialize request and returns session ID
func initializeMCPSession(t *testing.T, ctx *HTTPTestContext) (string, error) {
	// Build MCP initialize request
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-11-25",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "cqlai-test-client",
				"version": "1.0.0",
			},
		},
	}

	requestJSON, _ := json.Marshal(request)

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", ctx.BaseURL, bytes.NewReader(requestJSON))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", ctx.APIKey)

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return "", fmt.Errorf("initialize failed with status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	// Read response
	respBody, _ := io.ReadAll(httpResp.Body)
	var response map[string]any
	json.Unmarshal(respBody, &response)

	// Extract session ID from headers or response
	sessionID := httpResp.Header.Get("MCP-Session-Id")
	if sessionID == "" {
		// Try to get from response if not in header
		if result, ok := response["result"].(map[string]any); ok {
			if sid, ok := result["sessionId"].(string); ok {
				sessionID = sid
			}
		}
	}

	t.Logf("MCP session initialized: %s", sessionID)
	return sessionID, nil
}

// stopMCPHTTP stops MCP server and cleans up
func stopMCPHTTP(ctx *HTTPTestContext) {
	if ctx == nil {
		return
	}
	if ctx.MCPHandler != nil {
		ctx.MCPHandler.HandleMCPCommand(".mcp stop")
	}
	if ctx.Session != nil {
		ctx.Session.Close()
	}
	time.Sleep(100 * time.Millisecond) // Brief pause for HTTP server to shutdown
}

// callToolHTTP calls MCP tool via HTTP
func callToolHTTP(t *testing.T, ctx *HTTPTestContext, toolName string, args map[string]any) map[string]any {
	// Build JSON-RPC request
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
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
		return nil
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", ctx.BaseURL, bytes.NewReader(requestJSON))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
		return nil
	}

	// Add required headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", ctx.APIKey)
	if ctx.MCPSessionID != "" {
		httpReq.Header.Set("MCP-Session-Id", ctx.MCPSessionID)
	}

	// Send request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
		return nil
	}
	defer httpResp.Body.Close()

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		t.Fatalf("HTTP request failed with status %d: %s", httpResp.StatusCode, string(bodyBytes))
		return nil
	}

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
		return nil
	}

	// Parse JSON-RPC response
	var response map[string]any
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
		return nil
	}

	return response
}

// ============================================================================
// Reference Tests - Simple Cases to Validate HTTP Works
// ============================================================================

func TestHTTP_ListKeyspaces(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create minimal HTTP config
	tmpDir := t.TempDir()
	configPath := tmpDir + "/http_test.json"

	apiKey, _ := ai.GenerateAPIKey()
	configJSON := fmt.Sprintf(`{
		"http_host": "127.0.0.1",
		"http_port": 8889,
		"api_key": "%s",
		"mode": "readonly"
	}`, apiKey)

	err := os.WriteFile(configPath, []byte(configJSON), 0644)
	require.NoError(t, err)

	ctx := startMCPFromConfigHTTP(t, configPath)
	defer stopMCPHTTP(ctx)

	t.Run("list_keyspaces via HTTP", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "list_keyspaces", map[string]any{})
		assertNotError(t, resp, "list_keyspaces should work in readonly mode")

		// Verify we got keyspaces
		text := extractText(t, resp)
		assert.Contains(t, text, "system", "Should list system keyspace")
	})
}

func TestHTTP_SelectQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	tmpDir := t.TempDir()
	configPath := tmpDir + "/http_select.json"

	apiKey, _ := ai.GenerateAPIKey()
	configJSON := fmt.Sprintf(`{
		"http_host": "127.0.0.1",
		"http_port": 8890,
		"api_key": "%s",
		"mode": "readonly"
	}`, apiKey)

	err := os.WriteFile(configPath, []byte(configJSON), 0644)
	require.NoError(t, err)

	ctx := startMCPFromConfigHTTP(t, configPath)
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	t.Run("SELECT via HTTP", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "SELECT",
			"keyspace":  "test_mcp",
			"table":     "users",
		})
		assertNotError(t, resp, "SELECT should work in readonly mode")
	})
}

func TestHTTP_ConfirmationRequired(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	tmpDir := t.TempDir()
	configPath := tmpDir + "/http_confirm.json"

	apiKey, _ := ai.GenerateAPIKey()
	configJSON := fmt.Sprintf(`{
		"http_host": "127.0.0.1",
		"http_port": 8891,
		"api_key": "%s",
		"mode": "readonly"
	}`, apiKey)

	err := os.WriteFile(configPath, []byte(configJSON), 0644)
	require.NoError(t, err)

	ctx := startMCPFromConfigHTTP(t, configPath)
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	t.Run("INSERT requires confirmation (readonly mode)", func(t *testing.T) {
		resp := callToolHTTP(t, ctx, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000099",
				"name":  "Test User",
				"email": "test@example.com",
			},
		})
		assertIsError(t, resp, "INSERT should be blocked in readonly mode")
		text := extractText(t, resp)
		assert.Contains(t, text, "not allowed", "Should explain INSERT not allowed")
	})
}
