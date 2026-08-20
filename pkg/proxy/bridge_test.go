package proxy

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCommandString(t *testing.T) {
	name, args, err := ParseCommandString("python server.py --port 8080")
	require.NoError(t, err)
	assert.Equal(t, "python", name)
	assert.Equal(t, []string{"server.py", "--port", "8080"}, args)

	// Quotes and escaped spaces
	name, args, err = ParseCommandString(`python -c "import time; print('hello world')" --flag 'value with spaces'`)
	require.NoError(t, err)
	assert.Equal(t, "python", name)
	assert.Equal(t, []string{"-c", "import time; print('hello world')", "--flag", "value with spaces"}, args)

	_, _, err = ParseCommandString("")
	assert.Error(t, err)

	_, _, err = ParseCommandString(`python -c "unclosed quote`)
	assert.Error(t, err)
}

func TestSubprocessExecutionAndLoopback(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Launch simple echo binary 'cat'
	tp, err := SpawnTargetProcess(ctx, "cat", "test-peer-id")
	require.NoError(t, err)
	defer tp.Stop()

	// Write payload to process stdin
	inputMsg := []byte("hello mcp-wire stdio\n")
	_, err = tp.Stdin().Write(inputMsg)
	require.NoError(t, err)

	// Read echoed response from process stdout
	readBuf := make([]byte, len(inputMsg))
	_, err = io.ReadFull(tp.Stdout(), readBuf)
	require.NoError(t, err)

	assert.Equal(t, string(inputMsg), string(readBuf))
}
