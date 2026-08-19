package auth

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/mcp-wire/mcp-wire/pkg/config"
)

const MaxFrameBytes = 16 * 1024 * 1024 // 16MB max JSON-RPC payload size

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

var allowedReadOnlyMethods = map[string]bool{
	"tools/list":          true,
	"resources/list":      true,
	"resources/read":      true,
	"resources/templates": true,
	"prompts/list":        true,
	"prompts/get":         true,
	"ping":                true,
	"initialize":          true,
}

type ReadOnlyFilterReader struct {
	r        *bufio.Reader
	readOnly bool
}

func NewReadOnlyFilterReader(r io.Reader, readOnly bool) *ReadOnlyFilterReader {
	return &ReadOnlyFilterReader{
		r:        bufio.NewReaderSize(r, 64*1024),
		readOnly: readOnly,
	}
}

// ReadLineAndFilter inspects newline-delimited JSON-RPC requests
func (f *ReadOnlyFilterReader) ReadLineAndFilter(allowedTarget io.Writer, blockedTarget io.Writer) error {
	if blockedTarget == nil {
		blockedTarget = allowedTarget
	}

	line, err := f.r.ReadBytes('\n')
	if err != nil {
		return err
	}

	if len(line) > MaxFrameBytes {
		return fmt.Errorf("payload size exceeds maximum limit of 16MB")
	}

	if !f.readOnly {
		_, err := allowedTarget.Write(line)
		return err
	}

	// Try parsing JSON-RPC request
	trimmed := bytes.TrimSpace(line)
	if len(trimmed) == 0 {
		_, err := allowedTarget.Write(line)
		return err
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(trimmed, &req); err != nil || req.Method == "" {
		// Pass through unparseable or notifications
		_, err := allowedTarget.Write(line)
		return err
	}

	// Check if method is permitted in read-only mode
	if !allowedReadOnlyMethods[req.Method] {
		LogSecurityViolation(req.Method)
		// Return JSON-RPC error response -32601
		errResp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32601,
				Message: fmt.Sprintf("Method '%s' not allowed in read-only mode", req.Method),
			},
		}
		data, _ := json.Marshal(errResp)
		data = append(data, '\n')
		slog.Warn("blocked mutating request in read-only mode", "method", req.Method)
		_, err := blockedTarget.Write(data)
		return err
	}

	_, err = allowedTarget.Write(line)
	return err
}

func LogSecurityViolation(method string) {
	logPath, err := config.GetSecurityLogPath()
	if err != nil {
		return
	}

	entry := fmt.Sprintf("[%s] SECURITY VIOLATION: Mutating method '%s' blocked by read-only guard\n",
		time.Now().Format(time.RFC3339), method)

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err == nil {
		defer f.Close()
		_, _ = f.WriteString(entry)
	}
}

func IsReadOnlyMethodAllowed(method string) bool {
	return allowedReadOnlyMethods[strings.ToLower(method)]
}
