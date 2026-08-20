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

// ReadLineBounded reads a newline-terminated line up to maxBytes without unbounded buffering.
func ReadLineBounded(r *bufio.Reader, maxBytes int) ([]byte, error) {
	var line []byte
	for {
		segment, isPrefix, err := r.ReadLine()
		if err != nil {
			if len(line) > 0 && err == io.EOF {
				line = append(line, '\n')
				return line, nil
			}
			return nil, err
		}
		line = append(line, segment...)
		if len(line) > maxBytes {
			return nil, fmt.Errorf("payload size exceeds maximum limit of %d bytes", maxBytes)
		}
		if !isPrefix {
			break
		}
	}
	line = append(line, '\n')
	return line, nil
}

// ReadLineAndFilter inspects newline-delimited JSON-RPC requests
func (f *ReadOnlyFilterReader) ReadLineAndFilter(allowedTarget io.Writer, blockedTarget io.Writer) error {
	if blockedTarget == nil {
		blockedTarget = allowedTarget
	}

	line, err := ReadLineBounded(f.r, MaxFrameBytes)
	if err != nil {
		return err
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

	// Try single JSON-RPC request
	var req JSONRPCRequest
	errSingle := json.Unmarshal(trimmed, &req)

	if errSingle == nil && req.Method != "" {
		if !IsReadOnlyMethodAllowed(req.Method) {
			return f.respondBlocked(blockedTarget, req.ID, req.Method, fmt.Sprintf("Method '%s' not allowed in read-only mode", req.Method), -32601)
		}
		_, err := allowedTarget.Write(line)
		return err
	}

	// Try JSON-RPC batch array request
	var batch []JSONRPCRequest
	errBatch := json.Unmarshal(trimmed, &batch)

	if errBatch == nil && len(batch) > 0 {
		for _, bReq := range batch {
			if bReq.Method == "" || !IsReadOnlyMethodAllowed(bReq.Method) {
				method := bReq.Method
				if method == "" {
					method = "unknown"
				}
				return f.respondBlocked(blockedTarget, bReq.ID, method, fmt.Sprintf("Method '%s' in batch request not allowed in read-only mode", method), -32601)
			}
		}
		_, err := allowedTarget.Write(line)
		return err
	}

	// Fail-closed for unparseable or invalid JSON-RPC in read-only mode
	return f.respondBlocked(blockedTarget, nil, "parse_error", "Invalid or unparseable JSON-RPC request blocked in read-only mode", -32700)
}

func (f *ReadOnlyFilterReader) respondBlocked(blockedTarget io.Writer, id interface{}, method string, msg string, code int) error {
	LogSecurityViolation(method)
	errResp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: msg,
		},
	}
	data, _ := json.Marshal(errResp)
	data = append(data, '\n')
	slog.Warn("blocked request in read-only mode", "method", method, "reason", msg)
	_, err := blockedTarget.Write(data)
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
