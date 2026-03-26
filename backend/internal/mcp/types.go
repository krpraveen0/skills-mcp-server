package mcp

// JSON-RPC 2.0 types for the MCP protocol

// Request is an inbound JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string  `json:"jsonrpc"`
	ID      any     `json:"id"` // string | number | null
	Method  string  `json:"method"`
	Params  any     `json:"params,omitempty"`
}

// Response is an outbound JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Standard JSON-RPC error codes
const (
	ErrParse          = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternal       = -32603
)

// InitializeResult is returned by the initialize method.
type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
}

// Capabilities describes what the server supports.
type Capabilities struct {
	Tools *ToolsCapability `json:"tools,omitempty"`
}

// ToolsCapability signals tool support.
type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// ServerInfo identifies the server.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ToolsListResult is the response to tools/list.
type ToolsListResult struct {
	Tools []ToolDefinition `json:"tools"`
}

// ToolDefinition describes a single MCP tool.
type ToolDefinition struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema JSONSchema  `json:"inputSchema"`
}

// JSONSchema is a subset of JSON Schema for describing tool inputs.
type JSONSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]SchemaField `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

// SchemaField is a single field in a JSON schema.
type SchemaField struct {
	Type        string       `json:"type"`
	Description string       `json:"description,omitempty"`
	Default     any          `json:"default,omitempty"`
	Maximum     int          `json:"maximum,omitempty"`
	Items       *SchemaField `json:"items,omitempty"`
}

// CallToolParams is the params block for tools/call.
type CallToolParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// CallToolResult is the response for tools/call.
type CallToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock is a piece of content in a tool result.
type ContentBlock struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"`
}
