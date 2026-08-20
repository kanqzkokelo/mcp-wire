package auth

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadOnlyFilter(t *testing.T) {
	// Allowed call: tools/list
	allowedInput := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n"
	buf := bytes.NewBufferString(allowedInput)
	filter := NewReadOnlyFilterReader(buf, true)

	var out bytes.Buffer
	err := filter.ReadLineAndFilter(&out, &out)
	require.NoError(t, err)
	assert.Equal(t, allowedInput, out.String())

	// Blocked call: tools/call
	blockedInput := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"whisper"}}` + "\n"
	buf = bytes.NewBufferString(blockedInput)
	filter = NewReadOnlyFilterReader(buf, true)

	out.Reset()
	err = filter.ReadLineAndFilter(&out, &out)
	require.NoError(t, err)
	assert.Contains(t, out.String(), `-32601`)
	assert.Contains(t, out.String(), `Method 'tools/call' not allowed in read-only mode`)
}

func TestReadOnlyFilterBatchRequests(t *testing.T) {
	// Allowed batch
	batchAllowed := `[{"jsonrpc":"2.0","id":1,"method":"tools/list"},{"jsonrpc":"2.0","id":2,"method":"resources/list"}]` + "\n"
	buf := bytes.NewBufferString(batchAllowed)
	filter := NewReadOnlyFilterReader(buf, true)
	var out bytes.Buffer
	err := filter.ReadLineAndFilter(&out, &out)
	require.NoError(t, err)
	assert.Equal(t, batchAllowed, out.String())

	// Batch containing mutating method -> blocked
	batchMutating := `[{"jsonrpc":"2.0","id":1,"method":"tools/list"},{"jsonrpc":"2.0","id":2,"method":"tools/call"}]` + "\n"
	buf = bytes.NewBufferString(batchMutating)
	filter = NewReadOnlyFilterReader(buf, true)
	out.Reset()
	err = filter.ReadLineAndFilter(&out, &out)
	require.NoError(t, err)
	assert.Contains(t, out.String(), `-32601`)
	assert.Contains(t, out.String(), `tools/call`)
}

func TestReadOnlyFilterFailClosedInvalidJSON(t *testing.T) {
	invalidJSON := `{"jsonrpc":"2.0", invalid_json}` + "\n"
	buf := bytes.NewBufferString(invalidJSON)
	filter := NewReadOnlyFilterReader(buf, true)
	var out bytes.Buffer
	err := filter.ReadLineAndFilter(&out, &out)
	require.NoError(t, err)
	assert.Contains(t, out.String(), `-32700`)
	assert.Contains(t, out.String(), `Invalid or unparseable JSON-RPC request blocked in read-only mode`)
}

func TestReadLineBoundedLimit(t *testing.T) {
	// Generate payload exceeding max limit without newline
	largePayload := bytes.Repeat([]byte("A"), 100)
	buf := bytes.NewBuffer(largePayload)
	reader := bufio.NewReader(buf)

	_, err := ReadLineBounded(reader, 50)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `exceeds maximum limit`)
}
