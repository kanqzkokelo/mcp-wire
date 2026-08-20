package proxy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type TargetProcess struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	mu        sync.Mutex
	running   bool
	exitErr   error
	doneChan  chan struct{}
}

// ParseCommandString splits command into binary name and slice of arguments respecting quotes.
func ParseCommandString(cmdStr string) (string, []string, error) {
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return "", nil, fmt.Errorf("command string cannot be empty")
	}

	var args []string
	var current strings.Builder
	inDouble := false
	inSingle := false
	escaped := false

	for _, r := range cmdStr {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' && !inSingle {
			escaped = true
			continue
		}

		if r == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}

		if r == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}

		if (r == ' ' || r == '\t') && !inDouble && !inSingle {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	if inDouble || inSingle || escaped {
		return "", nil, fmt.Errorf("unmatched quote or trailing escape in command string")
	}

	if len(args) == 0 {
		return "", nil, fmt.Errorf("command string evaluated to no tokens")
	}

	return args[0], args[1:], nil
}

// SpawnTargetProcess executes the remote MCP server command
func SpawnTargetProcess(ctx context.Context, cmdStr string, peerIDStr string) (*TargetProcess, error) {
	name, args, err := ParseCommandString(cmdStr)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, name, args...)

	// Pass-through environment and inject MCP_WIRE_PEER_ID
	cmd.Env = append(os.Environ(), fmt.Sprintf("MCP_WIRE_PEER_ID=%s", peerIDStr))

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	tp := &TargetProcess{
		cmd:      cmd,
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		running:  true,
		doneChan: make(chan struct{}),
	}

	// Capture process stderr and log via slog
	go tp.captureStderr()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start target process '%s': %w", cmdStr, err)
	}

	go tp.monitorProcess()

	slog.Info("spawned target MCP server subprocess", "cmd", cmdStr, "pid", cmd.Process.Pid)
	return tp, nil
}

func (tp *TargetProcess) captureStderr() {
	scanner := bufio.NewScanner(tp.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		slog.Info("[target-server stderr]", "line", line)
	}
}

func (tp *TargetProcess) monitorProcess() {
	err := tp.cmd.Wait()
	tp.mu.Lock()
	tp.running = false
	tp.exitErr = err
	tp.mu.Unlock()
	close(tp.doneChan)
	slog.Info("target MCP server subprocess exited", "error", err)
}

func (tp *TargetProcess) Stdin() io.WriteCloser {
	return tp.stdin
}

func (tp *TargetProcess) Stdout() io.ReadCloser {
	return tp.stdout
}

func (tp *TargetProcess) Stop() error {
	tp.mu.Lock()
	if !tp.running {
		tp.mu.Unlock()
		return nil
	}
	tp.mu.Unlock()

	if tp.cmd.Process != nil {
		// Send SIGINT first
		_ = tp.cmd.Process.Signal(syscall.SIGINT)

		select {
		case <-tp.doneChan:
			return nil
		case <-time.After(3 * time.Second):
			// Escalate to SIGKILL on timeout
			_ = tp.cmd.Process.Kill()
		}
	}
	return nil
}
