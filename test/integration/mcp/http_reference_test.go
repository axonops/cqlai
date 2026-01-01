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
	"strings"
	"testing"
	"time"

	"github.com/axonops/cqlai/internal/ai"
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

	// Check what API key is in the config file
	apiKeyFromConfig := ""
	if key, ok := jsonConfig["api_key"].(string); ok {
		apiKeyFromConfig = key
	}

	// Determine which API key to use:
	// 1. If config has hardcoded key (no ${VAR}), use that
	// 2. If config has ${VAR}, ensure TEST_MCP_API_KEY is set
	var apiKey string
	if apiKeyFromConfig != "" && !strings.Contains(apiKeyFromConfig, "${") {
		// Hardcoded key in config - use it directly
		apiKey = apiKeyFromConfig
	} else {
		// Config uses ${VAR} or no key - check TEST_MCP_API_KEY
		apiKey = os.Getenv("TEST_MCP_API_KEY")
		if apiKey == "" {
			// Generate new key
			var err error
			apiKey, err = ai.GenerateAPIKey()
			require.NoError(t, err)

			// Set environment variable so config ${TEST_MCP_API_KEY} expands
			os.Setenv("TEST_MCP_API_KEY", apiKey)
			t.Cleanup(func() { os.Unsetenv("TEST_MCP_API_KEY") })
		}
	}

	// Start server with config file
	cmd := fmt.Sprintf(".mcp start --config-file %s", configPath)

	result := mcpHandler.HandleMCPCommand(cmd)
	require.Contains(t, result, "started successfully", "MCP start failed: %s", result)

	time.Sleep(500 * time.Millisecond) // Wait for HTTP server to start

	baseURL := fmt.Sprintf("http://%s:%d/mcp", httpHost, httpPort)

	ctx := &HTTPTestContext{
		Session:      replSession,
		MCPHandler:   mcpHandler,
		HTTPHost:     httpHost,
		HTTPPort:     httpPort,
		APIKey:       apiKey,
		BaseURL:      baseURL,
		MCPSessionID: "", // Will be set by first request or SSE connection
	}

	// For POST-only tests, initialize session
	// For SSE tests, the GET connection will register the session
	sessionID, err := initializeMCPSession(t, ctx)
	if err == nil {
		ctx.MCPSessionID = sessionID
	}

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

	// Try to parse as pure JSON first (simple tool calls)
	var response map[string]any
	err = json.Unmarshal(respBody, &response)
	if err == nil {
		// Pure JSON response
		return response
	}

	// If JSON parse failed, try SSE format (responses with notifications)
	response = parseSSEResponse(t, respBody)
	if response == nil {
		t.Logf("Response body: %s", string(respBody))
		t.Fatalf("Failed to parse response as JSON or SSE")
		return nil
	}

	return response
}

// parseSSEResponse extracts the JSON-RPC response from SSE-formatted body
func parseSSEResponse(t *testing.T, body []byte) map[string]any {
	// SSE format:
	// event: message
	// data: {"jsonrpc":"2.0",...}
	// (blank line)
	bodyStr := string(body)
	lines := strings.Split(bodyStr, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if strings.HasPrefix(line, "data: ") {
			dataJSON := strings.TrimPrefix(line, "data: ")

			var msg map[string]any
			if err := json.Unmarshal([]byte(dataJSON), &msg); err != nil {
				continue
			}

			// Look for response (has "id"), not notification (has "method" but no "id")
			if _, hasID := msg["id"]; hasID {
				return msg
			}
		}
	}

	return nil
}

// callToolHTTPWithNotifications calls a tool and extracts both notifications and result
// Returns: first notification found (if any), result
func callToolHTTPWithNotifications(t *testing.T, ctx *HTTPTestContext, toolName string, args map[string]any) (notification map[string]any, result map[string]any) {
	// Build and send request (same as callToolHTTP)
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      time.Now().UnixNano(),
		"method":  "tools/call",
		"params": map[string]any{
			"name":      toolName,
			"arguments": args,
		},
	}

	requestJSON, _ := json.Marshal(request)
	httpReq, _ := http.NewRequest("POST", ctx.BaseURL, bytes.NewReader(requestJSON))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", ctx.APIKey)
	if ctx.MCPSessionID != "" {
		httpReq.Header.Set("MCP-Session-Id", ctx.MCPSessionID)
	}

	client := &http.Client{Timeout: 15 * time.Second} // Longer timeout for blocking operations
	httpResp, err := client.Do(httpReq)
	if err != nil {
		t.Logf("HTTP request failed: %v", err)
		return nil, nil
	}
	defer httpResp.Body.Close()

	// Read entire response body
	respBody, _ := io.ReadAll(httpResp.Body)

	// Parse all JSON-RPC messages from response (SSE format or plain JSON)
	messages := parseAllMessagesFromResponse(t, respBody)
	t.Logf("Parsed %d messages from response", len(messages))

	// Separate notifications from result
	for _, msg := range messages {
		if method, ok := msg["method"].(string); ok {
			// This is a notification (has "method", no "id")
			t.Logf("Found notification: method=%s", method)
			if notification == nil {
				// Capture first notification
				notification = msg
			}
		} else if _, hasID := msg["id"]; hasID {
			// This is the result (has "id")
			result = msg
		}
	}

	return notification, result
}

// parseAllMessagesFromResponse extracts all JSON-RPC messages from HTTP response
func parseAllMessagesFromResponse(t *testing.T, body []byte) []map[string]any {
	var messages []map[string]any

	// Try plain JSON first (single object)
	var singleMsg map[string]any
	if err := json.Unmarshal(body, &singleMsg); err == nil {
		return []map[string]any{singleMsg}
	}

	// Try SSE format (event: / data: lines)
	bodyStr := string(body)
	lines := strings.Split(bodyStr, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if strings.HasPrefix(line, "data: ") {
			dataJSON := strings.TrimPrefix(line, "data: ")

			var msg map[string]any
			if err := json.Unmarshal([]byte(dataJSON), &msg); err != nil {
				t.Logf("Failed to parse JSON: %v, data: %s", err, dataJSON)
				continue
			}

			messages = append(messages, msg)
		}
	}

	return messages
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

	t.Run("INSERT not allowed (readonly mode - IMMEDIATE error, NO streaming)", func(t *testing.T) {
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

		// Should get IMMEDIATE error (not streaming, not blocking)
		assertIsError(t, resp, "INSERT should be blocked in readonly mode")
		text := extractText(t, resp)

		// Verify error message explains the issue
		assert.Contains(t, text, "not allowed", "Should explain INSERT not allowed")

		t.Logf("NOT ALLOWED error message:\n%s", text)
		t.Logf("✅ NOT ALLOWED queries return immediate error (no streaming, no blocking)")
	})

	t.Run("NOT ALLOWED with runtime changes disabled", func(t *testing.T) {
		// Create config with runtime changes disabled
		tmpDir2 := t.TempDir()
		configPath2 := tmpDir2 + "/readonly_locked.json"

		apiKey2, _ := ai.GenerateAPIKey()
		configJSON2 := fmt.Sprintf(`{
			"http_host": "127.0.0.1",
			"http_port": 8896,
			"api_key": "%s",
			"mode": "readonly",
			"disable_runtime_permission_changes": true
		}`, apiKey2)

		os.WriteFile(configPath2, []byte(configJSON2), 0644)
		ctx2 := startMCPFromConfigHTTP(t, configPath2)
		defer stopMCPHTTP(ctx2)

		ensureTestDataExists(t, ctx2.Session)

		resp := callToolHTTP(t, ctx2, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000098",
				"name":  "Locked Test",
				"email": "locked@example.com",
			},
		})

		assertIsError(t, resp, "INSERT should be blocked")
		text := extractText(t, resp)

		// Should mention that runtime changes are disabled
		assert.Contains(t, text, "Runtime permission changes are disabled")
		assert.Contains(t, text, "restart the MCP server")
		assert.NotContains(t, text, "update_mcp_permissions tool", "Should NOT suggest tool when disabled")

		t.Logf("NOT ALLOWED (locked) message:\n%s", text)
		t.Logf("✅ Error message adapts based on whether runtime changes allowed")
	})
}

func TestHTTP_StreamingConfirmation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	tmpDir := t.TempDir()
	configPath := tmpDir + "/http_streaming.json"

	apiKey, _ := ai.GenerateAPIKey()
	configJSON := fmt.Sprintf(`{
		"http_host": "127.0.0.1",
		"http_port": 8894,
		"api_key": "%s",
		"mode": "readwrite",
		"confirm_queries": ["dml"],
		"allow_mcp_request_approval": true,
		"confirmation_timeout_seconds": 30
	}`, apiKey)

	err := os.WriteFile(configPath, []byte(configJSON), 0644)
	require.NoError(t, err)

	ctx := startMCPFromConfigHTTP(t, configPath)
	defer stopMCPHTTP(ctx)

	ensureTestDataExists(t, ctx.Session)

	t.Run("HTTP connection stays open during confirmation wait", func(t *testing.T) {
		// Channel to capture result from goroutine
		responseChan := make(chan map[string]any, 1)
		errorChan := make(chan error, 1)

		// Thread 1: Submit query (will BLOCK waiting for confirmation)
		// Should receive: 1) "waiting" notification, 2) (blocks), 3) final result
		go func() {
			t.Logf("Thread 1: Submitting INSERT (requires confirmation)...")

			// Use callToolHTTPWithNotifications to capture the initial notification
			notification, resp := callToolHTTPWithNotifications(t, ctx, "submit_query_plan", map[string]any{
				"operation": "INSERT",
				"keyspace":  "test_mcp",
				"table":     "users",
				"values": map[string]any{
					"id":    "00000000-0000-0000-0000-000000000066",
					"name":  "Streaming Test",
					"email": "streaming@test.com",
				},
			})

			if notification != nil {
				t.Logf("Thread 1: Received initial notification: %v", notification["method"])
				if method, ok := notification["method"].(string); ok && method == "confirmation/requested" {
					params := notification["params"].(map[string]any)
					t.Logf("  → Request ID: %s", params["request_id"])
					t.Logf("  → Timeout: %v seconds", params["timeout_seconds"])
					t.Logf("✅ Initial 'waiting for confirmation' notification received!")
				}
			}

			if resp == nil {
				errorChan <- fmt.Errorf("nil response")
				return
			}
			responseChan <- resp
			t.Logf("Thread 1: Got final response!")
		}()

		// Give submit_query_plan time to create confirmation request
		time.Sleep(1 * time.Second)

		// Get the request ID by checking pending
		t.Logf("Thread 2: Checking pending confirmations...")
		pendingResp := callToolHTTP(t, ctx, "get_pending_confirmations", map[string]any{})
		require.NotNil(t, pendingResp, "Should get pending confirmations")
		pendingText := extractText(t, pendingResp)
		requestID := extractRequestID(pendingText)
		require.NotEmpty(t, requestID, "Should have pending request")
		t.Logf("Thread 2: Found pending request: %s", requestID)

		// Thread 2: Approve the request (should unblock Thread 1)
		t.Logf("Thread 2: Approving request %s...", requestID)
		_, confirmResult := callToolHTTPWithNotifications(t, ctx, "confirm_request", map[string]any{
			"request_id":     requestID,
			"user_confirmed": true,
		})
		assertNotError(t, confirmResult, "confirm_request should succeed")
		t.Logf("Thread 2: Request approved!")

		// Thread 1 should now receive response with query execution results
		select {
		case resp := <-responseChan:
			t.Logf("Thread 1: Received response after confirmation!")
			assertNotError(t, resp, "Query should execute successfully after confirmation")

			text := extractText(t, resp)
			assert.Contains(t, text, "executed", "Should indicate query was executed")
			t.Logf("✅ STREAMING CONFIRMATION WORKS: Connection stayed open, query executed after approval")

		case err := <-errorChan:
			t.Fatalf("Error in submit thread: %v", err)

		case <-time.After(10 * time.Second):
			t.Fatal("Timeout: Thread 1 never received response after confirmation")
		}
	})

	t.Run("Timeout scenario with heartbeats", func(t *testing.T) {
		// This test uses a SHORT timeout (10 seconds) to verify timeout behavior
		// Recreate context with short timeout
		tmpDir2 := t.TempDir()
		configPath2 := tmpDir2 + "/short_timeout.json"

		apiKey2, _ := ai.GenerateAPIKey()
		configJSON2 := fmt.Sprintf(`{
			"http_host": "127.0.0.1",
			"http_port": 8895,
			"api_key": "%s",
			"mode": "readwrite",
			"confirm_queries": ["dml"],
			"confirmation_timeout_seconds": 10
		}`, apiKey2)

		os.WriteFile(configPath2, []byte(configJSON2), 0644)

		ctx2 := startMCPFromConfigHTTP(t, configPath2)
		defer stopMCPHTTP(ctx2)

		ensureTestDataExists(t, ctx2.Session)

		// Submit query and DON'T approve (let it timeout)
		t.Logf("Submitting INSERT without approval (will timeout in 10s)...")

		notification, resp := callToolHTTPWithNotifications(t, ctx2, "submit_query_plan", map[string]any{
			"operation": "INSERT",
			"keyspace":  "test_mcp",
			"table":     "users",
			"values": map[string]any{
				"id":    "00000000-0000-0000-0000-000000000055",
				"name":  "Timeout Test",
				"email": "timeout@test.com",
			},
		})

		// Should get initial notification
		require.NotNil(t, notification, "Should get confirmation/requested notification")
		t.Logf("Initial notification: %v", notification["method"])

		// After 10 seconds, should get TIMEOUT error
		require.NotNil(t, resp, "Should get response after timeout")
		assertIsError(t, resp, "Should get timeout error")

		text := extractText(t, resp)
		assert.Contains(t, text, "timed out", "Error should mention timeout")
		t.Logf("✅ Timeout behavior works: %s", text)
	})

	t.Run("Cancel scenario", func(t *testing.T) {
		responseChan := make(chan map[string]any, 1)

		// Thread 1: Submit and wait
		go func() {
			notification, resp := callToolHTTPWithNotifications(t, ctx, "submit_query_plan", map[string]any{
				"operation": "INSERT",
				"keyspace":  "test_mcp",
				"table":     "users",
				"values": map[string]any{
					"id":    "00000000-0000-0000-0000-000000000044",
					"name":  "Cancel Test",
					"email": "cancel@test.com",
				},
			})
			_ = notification
			responseChan <- resp
		}()

		time.Sleep(1 * time.Second)

		// Get request ID
		pendingResp := callToolHTTP(t, ctx, "get_pending_confirmations", map[string]any{})
		pendingText := extractText(t, pendingResp)
		requestID := extractRequestID(pendingText)
		require.NotEmpty(t, requestID)
		t.Logf("Found pending request: %s", requestID)

		// Thread 2: CANCEL the request
		t.Logf("Cancelling request %s...", requestID)
		cancelResp := callToolHTTP(t, ctx, "cancel_confirmation", map[string]any{
			"request_id": requestID,
		})
		assertNotError(t, cancelResp, "cancel_confirmation should succeed")

		// Thread 1 should receive CANCELLED error
		select {
		case resp := <-responseChan:
			assertIsError(t, resp, "Should get error for cancelled request")
			text := extractText(t, resp)
			assert.Contains(t, text, "cancel", "Error should mention cancellation")
			t.Logf("✅ Cancel message: %s", text)

		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for cancel response")
		}
	})
}

// listenForSSEEvents opens SSE connection and forwards events to channel
