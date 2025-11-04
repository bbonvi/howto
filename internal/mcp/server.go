package mcp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync/atomic"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/yourusername/howto/internal/app"
	"github.com/yourusername/howto/internal/instructions"
)

const (
	jsonRPCVersion    = "2.0"
	methodInitialize  = "initialize"
	methodInitialized = "initialized"
	methodPing        = "ping"
	methodShutdown    = "shutdown"
	methodExit        = "exit"
	methodToolsList   = "tools/list"
	methodToolsCall   = "tools/call"
)

// Error codes aligned with JSON-RPC 2.0 specs.
const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternalError  = -32603
)

// Tool names exposed by the server.
const (
	ToolListPlaybooks = "list_playbooks"
	ToolGetPlaybook   = "get_playbook"
)

// Server implements a minimal MCP-compatible JSON-RPC server over stdio.
type Server struct {
	decoder *json.Decoder
	encoder *json.Encoder

	loader  app.RegistryLoader
	version string

	logger       *log.Logger
	shuttingDown atomic.Bool
}

// NewServer constructs an MCP server that reads from in and writes to out.
func NewServer(in io.Reader, out io.Writer, loader app.RegistryLoader, version string, logger *log.Logger) *Server {
	if logger == nil {
		logger = log.New(os.Stderr, "howto-mcp: ", log.LstdFlags)
	}

	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)

	return &Server{
		decoder: json.NewDecoder(bufio.NewReader(in)),
		encoder: enc,
		loader:  loader,
		version: version,
		logger:  logger,
	}
}

// Serve processes incoming JSON-RPC requests until EOF or an exit notification.
func (s *Server) Serve() error {
	for {
		var req rawMessage
		if err := s.decoder.Decode(&req); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("failed to decode request: %w", err)
		}

		if req.JSONRPC != jsonRPCVersion {
			if req.HasID() {
				if err := s.sendError(req.ID, codeInvalidRequest, "unsupported JSON-RPC version", nil); err != nil {
					return err
				}
			}
			continue
		}

		if req.Method == "" {
			if req.HasID() {
				if err := s.sendError(req.ID, codeInvalidRequest, "missing method", nil); err != nil {
					return err
				}
			}
			continue
		}

		if !req.HasID() {
			if stop, err := s.handleNotification(req); err != nil {
				s.logger.Printf("notification error for %s: %v", req.Method, err)
			} else if stop {
				return nil
			}
			continue
		}

		if err := s.handleRequest(req); err != nil {
			return err
		}
	}
}

func (s *Server) handleNotification(msg rawMessage) (bool, error) {
	switch msg.Method {
	case methodInitialized:
		return false, nil
	case methodExit:
		if s.shuttingDown.Load() {
			return true, nil
		}
		return true, nil
	default:
		s.logger.Printf("ignoring unknown notification %q", msg.Method)
		return false, nil
	}
}

func (s *Server) handleRequest(msg rawMessage) error {
	switch msg.Method {
	case methodInitialize:
		return s.handleInitialize(msg)
	case methodPing:
		return s.sendResult(msg.ID, map[string]any{})
	case methodShutdown:
		s.shuttingDown.Store(true)
		return s.sendResult(msg.ID, map[string]any{})
	case methodToolsList:
		return s.handleToolsList(msg)
	case methodToolsCall:
		return s.handleToolsCall(msg)
	default:
		return s.sendError(msg.ID, codeMethodNotFound, fmt.Sprintf("unknown method %q", msg.Method), nil)
	}
}

func (s *Server) handleInitialize(msg rawMessage) error {
	var params initializeParams
	if len(msg.Params) > 0 {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return s.sendError(msg.ID, codeInvalidParams, "invalid initialize params", map[string]any{"error": err.Error()})
		}
	}

	result := initializeResult{
		ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
		ServerInfo: serverInfo{
			Name:    "howto-mcp",
			Version: s.version,
		},
		Capabilities: capabilities{
			Tools: toolsCapability{
				ListChanged: true,
			},
		},
		Instructions: instructions.MCPUsageInstructions(),
	}

	return s.sendResult(msg.ID, result)
}

func (s *Server) handleToolsList(msg rawMessage) error {
	var params struct{}
	if len(msg.Params) > 0 && string(msg.Params) != "{}" {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			// params not used but ensure valid JSON object
			return s.sendError(msg.ID, codeInvalidParams, "tools/list expects an object", nil)
		}
	}

	result := toolsListResult{
		Tools: []toolDefinition{
			{
				Name:        ToolListPlaybooks,
				Description: "List available playbooks with their descriptions and origin.",
				InputSchema: jsonSchema{
					Type:                 "object",
					Properties:           map[string]any{},
					Required:             []string{},
					AdditionalProperties: false,
				},
			},
			{
				Name:        ToolGetPlaybook,
				Description: "Fetch a specific playbook by name and return its Markdown content.",
				InputSchema: jsonSchema{
					Type: "object",
					Properties: map[string]any{
						"name": map[string]any{
							"type":        "string",
							"description": "Playbook name from the howto registry.",
						},
					},
					Required:             []string{"name"},
					AdditionalProperties: false,
				},
			},
		},
	}

	return s.sendResult(msg.ID, result)
}

func (s *Server) handleToolsCall(msg rawMessage) error {
	var params toolsCallParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.sendError(msg.ID, codeInvalidParams, "invalid tools/call params", map[string]any{"error": err.Error()})
	}

	arguments := params.Arguments
	if arguments == nil {
		arguments = map[string]any{}
	}

	switch params.Name {
	case ToolListPlaybooks:
		if len(arguments) != 0 {
			return s.sendError(msg.ID, codeInvalidParams, "list_playbooks does not accept arguments", nil)
		}
		return s.executeListPlaybooks(msg.ID)
	case ToolGetPlaybook:
		rawName, ok := arguments["name"]
		if !ok {
			return s.sendError(msg.ID, codeInvalidParams, "get_playbook requires a name argument", nil)
		}
		name, ok := rawName.(string)
		if !ok {
			return s.sendError(msg.ID, codeInvalidParams, "name must be a string", nil)
		}
		return s.executeGetPlaybook(msg.ID, strings.TrimSpace(name))
	default:
		return s.sendError(msg.ID, codeInvalidParams, fmt.Sprintf("unknown tool %q", params.Name), nil)
	}
}

func (s *Server) executeListPlaybooks(id json.RawMessage) error {
	reg, err := s.loader.Load()
	if err != nil {
		s.logger.Printf("failed to load registry: %v", err)
		return s.sendError(id, codeInternalError, "failed to load playbook registry", nil)
	}

	docs := app.DocumentsToList(reg)
	var builder strings.Builder

	if len(docs) == 0 {
		builder.WriteString("No playbooks available.")
	} else {
		builder.WriteString("Available playbooks:\n")
		for _, doc := range docs {
			builder.WriteString(fmt.Sprintf("- %s — %s\n", doc.Name, oneLine(doc.Description)))
		}
	}

	return s.sendResult(id, toolResponse{
		Content: []responseContent{
			{
				Type: "text",
				Text: strings.TrimRight(builder.String(), "\n"),
			},
		},
	})
}

func (s *Server) executeGetPlaybook(id json.RawMessage, name string) error {
	if name == "" {
		return s.sendError(id, codeInvalidParams, "name cannot be empty", nil)
	}

	reg, err := s.loader.Load()
	if err != nil {
		s.logger.Printf("failed to load registry: %v", err)
		return s.sendError(id, codeInternalError, "failed to load playbook registry", nil)
	}

	doc, ok := reg.Get(name)
	if !ok {
		return s.sendError(id, codeInvalidParams, fmt.Sprintf("unknown playbook %q", name), nil)
	}

	text := doc.Content
	if strings.TrimSpace(text) == "" {
		text = "(empty playbook)"
	}

	return s.sendResult(id, toolResponse{
		Content: []responseContent{
			{
				Type: "text",
				Text: text,
			},
		},
		Metadata: map[string]any{
			"name":        doc.Name,
			"description": doc.Description,
			"source":      doc.Source.String(),
		},
	})
}

func (s *Server) sendResult(id json.RawMessage, result any) error {
	resp := response{
		JSONRPC: jsonRPCVersion,
		ID:      &id,
		Result:  result,
	}
	return s.encoder.Encode(resp)
}

func (s *Server) sendError(id json.RawMessage, code int, message string, data any) error {
	resp := response{
		JSONRPC: jsonRPCVersion,
		ID:      &id,
		Error: &responseError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	return s.encoder.Encode(resp)
}

func oneLine(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "(no description)"
	}
	fields := strings.Fields(trimmed)
	return strings.Join(fields, " ")
}

type rawMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

func (r rawMessage) HasID() bool {
	return len(r.ID) > 0
}

type initializeParams struct {
	ClientInfo map[string]any `json:"clientInfo"`
}

type initializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	ServerInfo      serverInfo   `json:"serverInfo"`
	Capabilities    capabilities `json:"capabilities"`
	Instructions    string       `json:"instructions,omitempty"`
}

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type capabilities struct {
	Tools toolsCapability `json:"tools"`
}

type toolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type toolsListResult struct {
	Tools []toolDefinition `json:"tools"`
}

type toolDefinition struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema jsonSchema `json:"inputSchema"`
}

type jsonSchema struct {
	Type                 string         `json:"type"`
	Properties           map[string]any `json:"properties"`
	Required             []string       `json:"required"`
	AdditionalProperties bool           `json:"additionalProperties"`
}

type toolsCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type toolResponse struct {
	Content  []responseContent `json:"content"`
	Metadata map[string]any    `json:"metadata,omitempty"`
}

type responseContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type response struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Result  any              `json:"result,omitempty"`
	Error   *responseError   `json:"error,omitempty"`
}

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
