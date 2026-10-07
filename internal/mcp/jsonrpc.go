// Package mcp is cqlai's Model Context Protocol server: JSON-RPC 2.0 over
// stdin and stdout, one message per line.
//
// It speaks the protocol and nothing else. Which tools there are, what they
// do and what they may do are decided in internal/ai and internal/policy, so
// the CHAT view and an MCP client get the same tools, and every security
// decision is in one place.
//
// It is written on encoding/json, which is in the binary anyway. The official
// SDK would add 1.74 MiB, most of it a second and a third JSON library; it is
// used in the tests only, to check this server against the protocol.
package mcp

import (
	"encoding/json"
)

// The JSON-RPC error codes.
const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternalError  = -32603
)

// message is anything that comes in: a request has an id and a method, a
// notification a method and no id, a response an id and no method.
type message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

func (m message) isRequest() bool      { return m.Method != "" && len(m.ID) > 0 }
func (m message) isNotification() bool { return m.Method != "" && len(m.ID) == 0 }

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}
