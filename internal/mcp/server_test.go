package mcp

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"strings"
	"sync"
	"testing"

	"github.com/yourusername/howto/internal/parser"
	"github.com/yourusername/howto/internal/registry"
)

func TestServerHandlesHandshakeAndTools(t *testing.T) {
	loader := &stubLoader{
		reg: registry.Registry{
			"core-principles": {
				Name:        "core-principles",
				Description: "Core guidance for agents.",
				Content:     "Always follow the plays.",
				Source:      parser.SourceProjectScoped,
			},
		},
	}

	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"clientInfo":{"name":"tester"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_playbooks","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_playbook","arguments":{"name":"core-principles"}}}`,
	}, "\n")

	var output bytes.Buffer
	server := NewServer(strings.NewReader(input), &output, loader, "test", log.New(io.Discard, "", 0))

	if err := server.Serve(); err != nil {
		t.Fatalf("Serve() returned error: %v", err)
	}

	messages := decodeLines(t, output.String())
	if len(messages) != 4 {
		t.Fatalf("expected 4 responses, got %d", len(messages))
	}

	// initialize
	if messages[0].Error != nil {
		t.Fatalf("initialize returned error: %+v", messages[0].Error)
	}
	if result := messages[0].Result; result != nil {
		if protocol, ok := result["protocolVersion"].(string); !ok || protocol == "" {
			t.Fatalf("initialize response missing protocolVersion")
		}
	} else {
		t.Fatalf("initialize response missing result")
	}

	// tools/list
	if messages[1].Error != nil {
		t.Fatalf("tools/list returned error: %+v", messages[1].Error)
	}
	result, ok := messages[1].Result["tools"].([]any)
	if !ok || len(result) != 2 {
		t.Fatalf("expected two tools, got %#v", messages[1].Result["tools"])
	}

	// list_playbooks
	if messages[2].Error != nil {
		t.Fatalf("list_playbooks returned error: %+v", messages[2].Error)
	}
	verifyContentContains(t, messages[2].Result, "Available playbooks:")
	verifyContentContains(t, messages[2].Result, "core-principles")

	// get_playbook
	if messages[3].Error != nil {
		t.Fatalf("get_playbook returned error: %+v", messages[3].Error)
	}
	verifyContentContains(t, messages[3].Result, "Always follow the plays.")
}

func TestServerGetPlaybookError(t *testing.T) {
	loader := &stubLoader{
		reg: registry.Registry{},
	}

	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_playbook","arguments":{"name":"missing"}}}`
	var output bytes.Buffer
	server := NewServer(strings.NewReader(input), &output, loader, "test", log.New(io.Discard, "", 0))

	if err := server.Serve(); err != nil {
		t.Fatalf("Serve() returned error: %v", err)
	}

	messages := decodeLines(t, output.String())
	if len(messages) != 1 {
		t.Fatalf("expected 1 response, got %d", len(messages))
	}
	if messages[0].Error == nil {
		t.Fatalf("expected error response for missing playbook")
	}
	if messages[0].Error.Code != codeInvalidParams {
		t.Fatalf("expected invalid params code, got %d", messages[0].Error.Code)
	}
}

type stubLoader struct {
	mu  sync.Mutex
	reg registry.Registry
	err error
}

func (s *stubLoader) Load() (registry.Registry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.err != nil {
		return nil, s.err
	}

	copy := make(registry.Registry, len(s.reg))
	for k, v := range s.reg {
		copy[k] = v
	}
	return copy, nil
}

type message struct {
	ID     any            `json:"id"`
	Result map[string]any `json:"result"`
	Error  *messageError  `json:"error"`
}

type messageError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func decodeLines(t *testing.T, raw string) []message {
	t.Helper()
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	lines := strings.Split(raw, "\n")
	out := make([]message, 0, len(lines))
	for _, line := range lines {
		var msg message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			t.Fatalf("failed to decode response %q: %v", line, err)
		}
		out = append(out, msg)
	}
	return out
}

func verifyContentContains(t *testing.T, result map[string]any, expected string) {
	t.Helper()
	contentRaw, ok := result["content"]
	if !ok {
		t.Fatalf("result missing content field: %#v", result)
	}
	contentSlice, ok := contentRaw.([]any)
	if !ok || len(contentSlice) == 0 {
		t.Fatalf("content has unexpected shape: %#v", contentRaw)
	}

	first, ok := contentSlice[0].(map[string]any)
	if !ok {
		t.Fatalf("content entry has unexpected type: %#v", contentSlice[0])
	}

	text, _ := first["text"].(string)
	if !strings.Contains(text, expected) {
		t.Fatalf("expected content to contain %q, got %q", expected, text)
	}
}
